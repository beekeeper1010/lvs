package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:web/dist
var webFS embed.FS

func startServer(port int, dbPath string) {
	if err := openDB(dbPath); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	api.POST("/login", handleLogin)
	authed := api.Group("")
	authed.Use(authMiddleware())
	authed.POST("/logout", handleLogout)
	authed.GET("/video/list", handleVideoList)
	authed.GET("/video/play", handleVideoPlay)
	authed.GET("/video/thumb", handleVideoThumb)

	// 嵌入的前端静态资源与 SPA fallback
	if sub, err := fs.Sub(webFS, "web/dist"); err == nil {
		if assets, err := fs.Sub(sub, "assets"); err == nil {
			r.StaticFS("/assets", http.FS(assets))
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

	log.Printf("lvs 服务已启动: http://localhost:%d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal(err)
	}
}
