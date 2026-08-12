package main

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
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
		avatar   string
		pwdVer   int64
	)
	err := db.QueryRow(`SELECT id, password, nickname, role, avatar, pwd_ver FROM users WHERE username = ?`, req.Username).
		Scan(&id, &hash, &nickname, &role, &avatar, &pwdVer)
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
	respond(c, 0, "ok", gin.H{"token": token, "username": req.Username, "nickname": nickname, "role": role, "avatar": avatar})
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

// handleUserInfo 返回当前登录用户的最新信息（昵称/头像可能已被修改）
func handleUserInfo(c *gin.Context) {
	claims := c.MustGet("claims").(*Claims)
	var (
		nickname string
		role     string
		avatar   string
	)
	if err := db.QueryRow(`SELECT nickname, role, avatar FROM users WHERE id = ?`, claims.UserID).Scan(&nickname, &role, &avatar); err != nil {
		respond(c, 1, "用户不存在", nil)
		return
	}
	if nickname == "" {
		nickname = claims.Username
	}
	respond(c, 0, "ok", gin.H{"username": claims.Username, "nickname": nickname, "role": role, "avatar": avatar})
}

const avatarDir = "data/avatars"

var avatarExts = []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}

// saveAvatar 保存上传的头像文件并更新用户记录
func saveAvatar(c *gin.Context, username string) {
	file, err := c.FormFile("avatar")
	if err != nil {
		respond(c, 1, "缺少头像文件", nil)
		return
	}
	if file.Size > 2*1024*1024 {
		respond(c, 1, "头像文件不能超过 2MB", nil)
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !slices.Contains(avatarExts, ext) {
		respond(c, 1, "不支持的图片格式", nil)
		return
	}
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		respond(c, 1, "创建头像目录失败", nil)
		return
	}
	path := filepath.Join(avatarDir, username+ext)
	if err := c.SaveUploadedFile(file, path); err != nil {
		respond(c, 1, "保存头像失败", nil)
		return
	}
	// 清理同用户旧的其它扩展头像
	for _, e := range avatarExts {
		if e != ext {
			_ = os.Remove(filepath.Join(avatarDir, username+e))
		}
	}
	if _, err := db.Exec(`UPDATE users SET avatar = ? WHERE username = ?`, path, username); err != nil {
		respond(c, 1, "更新头像失败", nil)
		return
	}
	respond(c, 0, "ok", nil)
}

// handleUserAvatarUpload 上传当前登录用户头像
func handleUserAvatarUpload(c *gin.Context) {
	claims := c.MustGet("claims").(*Claims)
	saveAvatar(c, claims.Username)
}

// handleAdminUserAvatarUpload 管理员上传指定用户头像
func handleAdminUserAvatarUpload(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var username string
	if err := db.QueryRow(`SELECT username FROM users WHERE id = ?`, id).Scan(&username); err != nil {
		respond(c, 1, "用户不存在", nil)
		return
	}
	saveAvatar(c, username)
}

// handleUserAvatarGet 返回用户头像图片，无头像返回 204
func handleUserAvatarGet(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		respond(c, 1, "缺少用户名", nil)
		return
	}
	var avatar string
	if err := db.QueryRow(`SELECT avatar FROM users WHERE username = ?`, username).Scan(&avatar); err != nil || avatar == "" {
		c.Status(http.StatusNoContent)
		return
	}
	if _, err := os.Stat(avatar); err != nil {
		c.Status(http.StatusNoContent)
		return
	}
	// 禁用缓存，确保更换头像后立即生效
	c.Header("Cache-Control", "no-cache")
	c.File(avatar)
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
	id, err := createUser(req.Username, req.Password, req.Nickname)
	if err != nil {
		respond(c, 1, err.Error(), nil)
		return
	}
	respond(c, 0, "ok", gin.H{"id": id})
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
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Path      string  `json:"path"`
	ThumbPath string  `json:"thumb_path"`
	Duration  float64 `json:"duration"`
	LikeCount int64   `json:"like_count"`
	Liked     bool    `json:"liked"` // 当前登录用户是否已点赞
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
	claims := c.MustGet("claims").(*Claims)
	var total int64
	if err := db.QueryRow(`SELECT COUNT(*) FROM videos`).Scan(&total); err != nil {
		respond(c, 1, "查询失败", nil)
		return
	}
	// 当前用户已点赞的视频 id 集合
	likedIDs := map[int64]bool{}
	if rows, err := db.Query(`SELECT video_id FROM video_likes WHERE user_id = ?`, claims.UserID); err == nil {
		for rows.Next() {
			var vid int64
			if rows.Scan(&vid) == nil {
				likedIDs[vid] = true
			}
		}
		rows.Err()
		rows.Close()
	}
	rows, err := db.Query(
		`SELECT id, name, path, thumb_path, duration, like_count FROM videos ORDER BY id ASC LIMIT ? OFFSET ?`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		respond(c, 1, "查询失败", nil)
		return
	}
	defer rows.Close()
	list := make([]videoItem, 0, pageSize)
	for rows.Next() {
		var v videoItem
		if err := rows.Scan(&v.ID, &v.Name, &v.Path, &v.ThumbPath, &v.Duration, &v.LikeCount); err != nil {
			continue
		}
		v.Liked = likedIDs[v.ID]
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

// handleVideoLike 点赞/取消点赞（每用户仅一次），返回最新点赞数与点赞状态
func handleVideoLike(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		respond(c, 1, "无效的视频ID", nil)
		return
	}
	var req struct {
		Liked bool `json:"liked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, 1, "请求参数错误", nil)
		return
	}
	// 校验视频存在
	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM videos WHERE id = ?`, id).Scan(&exists); err != nil || exists == 0 {
		respond(c, 1, "视频不存在", nil)
		return
	}
	claims := c.MustGet("claims").(*Claims)
	// 在事务中更新点赞记录与计数，保证一致性
	tx, err := db.Begin()
	if err != nil {
		respond(c, 1, "操作失败", nil)
		return
	}
	defer tx.Rollback()
	if req.Liked {
		// 已存在则忽略（幂等），不存在才插入并 +1
		if _, err := tx.Exec(
			`INSERT INTO video_likes(video_id, user_id, created_at) VALUES(?, ?, ?) ON CONFLICT(video_id, user_id) DO NOTHING`,
			id, claims.UserID, time.Now()); err != nil {
			respond(c, 1, "点赞失败", nil)
			return
		}
		if _, err := tx.Exec(
			`UPDATE videos SET like_count = (SELECT COUNT(*) FROM video_likes WHERE video_id = ?) WHERE id = ?`, id, id); err != nil {
			respond(c, 1, "点赞失败", nil)
			return
		}
	} else {
		if _, err := tx.Exec(
			`DELETE FROM video_likes WHERE video_id = ? AND user_id = ?`, id, claims.UserID); err != nil {
			respond(c, 1, "取消点赞失败", nil)
			return
		}
		if _, err := tx.Exec(
			`UPDATE videos SET like_count = (SELECT COUNT(*) FROM video_likes WHERE video_id = ?) WHERE id = ?`, id, id); err != nil {
			respond(c, 1, "取消点赞失败", nil)
			return
		}
	}
	var count int64
	if err := tx.QueryRow(`SELECT like_count FROM videos WHERE id = ?`, id).Scan(&count); err != nil {
		respond(c, 1, "操作失败", nil)
		return
	}
	if err := tx.Commit(); err != nil {
		respond(c, 1, "操作失败", nil)
		return
	}
	respond(c, 0, "ok", gin.H{"like_count": count, "liked": req.Liked})
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
