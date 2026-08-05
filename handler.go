package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// respond 统一响应格式 {code, msg, data}, code=0 成功
func respond(c *gin.Context, code int, msg string, data any) {
	c.JSON(http.StatusOK, gin.H{"code": code, "msg": msg, "data": data})
}

func handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		respond(c, 1, "用户名或密码不能为空", nil)
		return
	}
	var (
		id   int64
		hash string
	)
	err := db.QueryRow(`SELECT id, password FROM users WHERE username = ?`, req.Username).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		respond(c, 1, "用户名或密码错误", nil)
		return
	}
	token, err := generateToken(id, req.Username)
	if err != nil {
		respond(c, 1, "登录失败", nil)
		return
	}
	respond(c, 0, "ok", gin.H{"token": token, "username": req.Username})
}

// handleLogout 由前端删除本地 token，服务端仅返回成功
func handleLogout(c *gin.Context) {
	respond(c, 0, "ok", nil)
}

type videoItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	ThumbPath string `json:"thumb_path"`
	CreatedAt string `json:"created_at"`
}

func handleVideoList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	var total int64
	if err := db.QueryRow(`SELECT COUNT(*) FROM videos`).Scan(&total); err != nil {
		respond(c, 1, "查询失败", nil)
		return
	}
	rows, err := db.Query(
		`SELECT id, name, path, thumb_path, created_at FROM videos ORDER BY id DESC LIMIT ? OFFSET ?`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		respond(c, 1, "查询失败", nil)
		return
	}
	defer rows.Close()
	list := make([]videoItem, 0, pageSize)
	for rows.Next() {
		var v videoItem
		var created time.Time
		if err := rows.Scan(&v.ID, &v.Name, &v.Path, &v.ThumbPath, &created); err != nil {
			continue
		}
		v.CreatedAt = created.Format("2006-01-02 15:04:05")
		list = append(list, v)
	}
	if err := rows.Err(); err != nil {
		respond(c, 1, "查询失败", nil)
		return
	}
	respond(c, 0, "ok", gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

// handleVideoPlay 基于 http.ServeContent 支持 Range 分片下载播放
func handleVideoPlay(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if id <= 0 {
		respond(c, 1, "无效的视频ID", nil)
		return
	}
	var path string
	if err := db.QueryRow(`SELECT path FROM videos WHERE id = ?`, id).Scan(&path); err != nil {
		respond(c, 1, "视频不存在", nil)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		respond(c, 1, "视频文件不存在", nil)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		respond(c, 1, "读取视频失败", nil)
		return
	}
	http.ServeContent(c.Writer, c.Request, filepath.Base(path), st.ModTime(), f)
}

func handleVideoThumb(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if id <= 0 {
		respond(c, 1, "无效的视频ID", nil)
		return
	}
	var thumb string
	if err := db.QueryRow(`SELECT thumb_path FROM videos WHERE id = ?`, id).Scan(&thumb); err != nil || thumb == "" {
		respond(c, 1, "缩略图不存在", nil)
		return
	}
	if _, err := os.Stat(thumb); err != nil {
		respond(c, 1, "缩略图文件不存在", nil)
		return
	}
	c.File(thumb)
}
