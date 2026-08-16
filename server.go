package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed all:web/dist
var webFS embed.FS

// newEngine 构建 gin 引擎并注册全部路由（独立出来便于测试复用）
func newEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	api.GET("/captcha", handleCaptcha)
	api.POST("/login", handleLogin)
	api.GET("/video/play", handleVideoPlay) // 公开路由：内部用短时播放票据鉴权
	authed := api.Group("")
	authed.Use(authMiddleware())
	authed.POST("/logout", handleLogout)
	authed.GET("/user/info", handleUserInfo)
	authed.PUT("/user/profile", handleUserProfile)
	authed.POST("/user/avatar", handleUserAvatarUpload)
	authed.GET("/user/avatar", handleUserAvatarGet)
	authed.GET("/video/list", handleVideoList)
	authed.POST("/video/ticket", handleVideoTicket)
	authed.GET("/video/thumb", handleVideoThumb)
	authed.GET("/video/adjacent", handleVideoAdjacent)
	authed.PUT("/video/:id/like", handleVideoLike)

	// 用户管理（仅 admin）
	admin := authed.Group("")
	admin.Use(adminAuthMiddleware())
	admin.GET("/admin/users", handleAdminUsers)
	admin.POST("/admin/users", handleAdminUserCreate)
	admin.PUT("/admin/users/:id", handleAdminUserUpdate)
	admin.DELETE("/admin/users/:id", handleAdminUserDelete)
	admin.DELETE("/video/:id", handleVideoDelete)
	admin.POST("/admin/users/:id/avatar", handleAdminUserAvatarUpload)
	admin.PUT("/admin/users/:id/lock", handleAdminUserLock)

	// 嵌入的前端静态资源与 SPA fallback
	if sub, err := fs.Sub(webFS, "web/dist"); err == nil {
		if assets, err := fs.Sub(sub, "assets"); err == nil {
			r.StaticFS("/assets", http.FS(assets))
		}
		// 站点图标（需显式服务，避免被 SPA fallback 拦截）
		if favicon, err := fs.ReadFile(sub, "favicon.svg"); err == nil {
			r.GET("/favicon.svg", func(c *gin.Context) {
				c.Data(http.StatusOK, "image/svg+xml", favicon)
			})
		}
		index, _ := fs.ReadFile(sub, "index.html")
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				respond(c, 404, "接口不存在", nil)
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", index)
		})
	}

	return r
}

func startServer(port int, dbPath string) {
	if err := openDB(dbPath); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := newEngine()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("lvs 服务已启动: http://localhost:%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 等待退出信号（Ctrl+C / kill），优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("收到退出信号, 正在优雅关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务关闭异常: %v", err)
	}
	log.Println("lvs 服务已退出")
}
