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
		id       int64
		hash     string
		nickname string
		role     string
		pwdVer   int64
	)
	err := db.QueryRow(`SELECT id, password, nickname, role, pwd_ver FROM users WHERE username = ?`, req.Username).
		Scan(&id, &hash, &nickname, &role, &pwdVer)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		respond(c, 1, "用户名或密码错误", nil)
		return
	}
	if nickname == "" {
		nickname = req.Username
	}
	token, err := generateToken(id, req.Username, nickname, role, pwdVer)
	if err != nil {
		respond(c, 1, "登录失败", nil)
		return
	}
	respond(c, 0, "ok", gin.H{"token": token, "username": req.Username, "nickname": nickname, "role": role})
}

// handleLogout 由前端删除本地 token，服务端仅返回成功
func handleLogout(c *gin.Context) {
	respond(c, 0, "ok", nil)
}

// handleUserProfile 修改当前用户昵称/密码。改密码须校验当前密码
func handleUserProfile(c *gin.Context) {
	claims := c.MustGet("claims").(*Claims)
	var req struct {
		Nickname    string `json:"nickname"`
		Password    string `json:"password"`
		OldPassword string `json:"old_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, 1, "请求参数错误", nil)
		return
	}
	if req.Nickname == "" && req.Password == "" {
		respond(c, 1, "至少要修改一项", nil)
		return
	}
	if req.Password != "" {
		var hash string
		if err := db.QueryRow(`SELECT password FROM users WHERE id = ?`, claims.UserID).Scan(&hash); err != nil {
			respond(c, 1, "用户不存在", nil)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.OldPassword)) != nil {
			respond(c, 1, "当前密码错误", nil)
			return
		}
	}
	if err := updateUser(claims.UserID, req.Nickname, req.Password); err != nil {
		respond(c, 1, err.Error(), nil)
		return
	}
	respond(c, 0, "ok", nil)
}

// handleUserInfo 返回当前登录用户的最新信息（昵称可能已被修改）
func handleUserInfo(c *gin.Context) {
	claims := c.MustGet("claims").(*Claims)
	var (
		nickname string
		role     string
	)
	if err := db.QueryRow(`SELECT nickname, role FROM users WHERE id = ?`, claims.UserID).Scan(&nickname, &role); err != nil {
		respond(c, 1, "用户不存在", nil)
		return
	}
	if nickname == "" {
		nickname = claims.Username
	}
	respond(c, 0, "ok", gin.H{"username": claims.Username, "nickname": nickname, "role": role})
}

// adminAuthMiddleware 仅允许 admin 角色访问
func adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Get("claims")
		if !ok || claims.(*Claims).Role != "admin" {
			respond(c, 2, "无权限", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

// ---------- 用户管理接口（仅 admin） ----------

func handleAdminUsers(c *gin.Context) {
	users, err := listUsers()
	if err != nil {
		respond(c, 1, "查询失败", nil)
		return
	}
	respond(c, 0, "ok", users)
}

func handleAdminUserCreate(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		respond(c, 1, "用户名和密码不能为空", nil)
		return
	}
	if err := createUser(req.Username, req.Password, req.Nickname); err != nil {
		respond(c, 1, err.Error(), nil)
		return
	}
	respond(c, 0, "ok", nil)
}

func handleAdminUserUpdate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		respond(c, 1, "无效的用户ID", nil)
		return
	}
	var req struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, 1, "请求参数错误", nil)
		return
	}
	if err := updateUser(id, req.Nickname, req.Password); err != nil {
		respond(c, 1, err.Error(), nil)
		return
	}
	respond(c, 0, "ok", nil)
}

func handleAdminUserDelete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		respond(c, 1, "无效的用户ID", nil)
		return
	}
	if err := deleteUserByID(id); err != nil {
		respond(c, 1, err.Error(), nil)
		return
	}
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

// handleVideoDelete 删除视频数据库记录与缩略图（不删除源视频文件），仅 admin
func handleVideoDelete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		respond(c, 1, "无效的视频ID", nil)
		return
	}
	var thumb string
	if err := db.QueryRow(`SELECT thumb_path FROM videos WHERE id = ?`, id).Scan(&thumb); err != nil {
		respond(c, 1, "视频不存在", nil)
		return
	}
	if _, err := db.Exec(`DELETE FROM videos WHERE id = ?`, id); err != nil {
		respond(c, 1, "删除失败", nil)
		return
	}
	// 清理缩略图文件（衍生产物），源视频文件保留
	if thumb != "" {
		_ = os.Remove(thumb)
	}
	respond(c, 0, "ok", nil)
}

// handleVideoAdjacent 返回指定视频的前一个与后一个（按 id 排序），用于播放页切换
func handleVideoAdjacent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if id <= 0 {
		respond(c, 1, "无效的视频ID", nil)
		return
	}
	type item struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	var (
		prev *item
		next *item
	)
	var (
		pid int64
		pn  string
	)
	if err := db.QueryRow(`SELECT id, name FROM videos WHERE id < ? ORDER BY id DESC LIMIT 1`, id).Scan(&pid, &pn); err == nil {
		prev = &item{pid, pn}
	}
	var (
		nid int64
		nn  string
	)
	if err := db.QueryRow(`SELECT id, name FROM videos WHERE id > ? ORDER BY id ASC LIMIT 1`, id).Scan(&nid, &nn); err == nil {
		next = &item{nid, nn}
	}
	respond(c, 0, "ok", gin.H{"prev": prev, "next": next})
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
