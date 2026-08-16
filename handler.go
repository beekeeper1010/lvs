package main

import (
	"fmt"
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

// boolInt 布尔转 0/1，用于 SQLite 布尔字段
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// maxLoginFails 连续输错密码锁定阈值
const maxLoginFails = 5

func handleLogin(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		CaptchaID   string `json:"captcha_id"`
		CaptchaCode string `json:"captcha_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		respond(c, 1, "用户名和密码不能为空", nil)
		return
	}
	// 验证码校验（一次性，校验后即失效）
	if !verifyCaptcha(req.CaptchaID, req.CaptchaCode) {
		respond(c, 1, "验证码错误", nil)
		return
	}
	var (
		id             int64
		hash           string
		nickname       string
		role           string
		avatar         string
		pwdVer         int64
		locked         int
		loginFailCount int
	)
	err := db.QueryRow(`SELECT id, password, nickname, role, avatar, pwd_ver, locked, login_fail_count FROM users WHERE username = ?`, req.Username).
		Scan(&id, &hash, &nickname, &role, &avatar, &pwdVer, &locked, &loginFailCount)
	if err != nil {
		// 区分"用户名不存在"与"密码错误"
		respond(c, 1, "用户名不存在", nil)
		return
	}
	// 锁定账户：即使密码正确也拒绝登录，需管理员解锁
	if locked == 1 {
		respond(c, 1, "账户已锁定，请联系管理员", gin.H{"locked": true})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		newCount := loginFailCount + 1
		if newCount >= maxLoginFails {
			// 连续输错达到上限，锁定账户
			if _, err := db.Exec(`UPDATE users SET login_fail_count = ?, locked = 1 WHERE id = ?`, newCount, id); err != nil {
				respond(c, 1, "登录失败", nil)
				return
			}
			respond(c, 1, "密码错误次数过多，账户已锁定，请联系管理员", gin.H{"locked": true, "remaining": 0})
			return
		}
		if _, err := db.Exec(`UPDATE users SET login_fail_count = ? WHERE id = ?`, newCount, id); err != nil {
			respond(c, 1, "登录失败", nil)
			return
		}
		remaining := maxLoginFails - newCount
		respond(c, 1, fmt.Sprintf("密码错误，还有 %d 次机会", remaining), gin.H{"remaining": remaining})
		return
	}
	// 登录成功：清零连续失败次数
	if _, err := db.Exec(`UPDATE users SET login_fail_count = 0 WHERE id = ?`, id); err != nil {
		respond(c, 1, "登录失败", nil)
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

// handleAdminUserLock 管理员锁定/解锁用户（仅 admin）。锁定 admin 账号被禁止
func handleAdminUserLock(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		respond(c, 1, "无效的用户ID", nil)
		return
	}
	var req struct {
		Locked bool `json:"locked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, 1, "请求参数错误", nil)
		return
	}
	var role string
	if err := db.QueryRow(`SELECT role FROM users WHERE id = ?`, id).Scan(&role); err != nil {
		respond(c, 1, "用户不存在", nil)
		return
	}
	if req.Locked && role == "admin" {
		respond(c, 1, "管理员账号不可锁定", nil)
		return
	}
	if _, err := db.Exec(`UPDATE users SET locked = ? WHERE id = ?`, boolInt(req.Locked), id); err != nil {
		respond(c, 1, "操作失败", nil)
		return
	}
	// 解锁时清零连续失败次数
	if !req.Locked {
		if _, err := db.Exec(`UPDATE users SET login_fail_count = 0 WHERE id = ?`, id); err != nil {
			respond(c, 1, "操作失败", nil)
			return
		}
	}
	respond(c, 0, "ok", nil)
}

type videoItem struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
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
		`SELECT id, name, thumb_path, duration, like_count FROM videos ORDER BY id ASC LIMIT ? OFFSET ?`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		respond(c, 1, "查询失败", nil)
		return
	}
	defer rows.Close()
	list := make([]videoItem, 0, pageSize)
	for rows.Next() {
		var v videoItem
		if err := rows.Scan(&v.ID, &v.Name, &v.ThumbPath, &v.Duration, &v.LikeCount); err != nil {
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

// handleVideoTicket 为播放签发短时票据（需登录），绑定用户/视频/客户端 IP
func handleVideoTicket(c *gin.Context) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		respond(c, 1, "无效的视频ID", nil)
		return
	}
	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM videos WHERE id = ?`, req.ID).Scan(&exists); err != nil || exists == 0 {
		respond(c, 1, "视频不存在", nil)
		return
	}
	claims := c.MustGet("claims").(*Claims)
	ticket, err := generatePlayTicket(claims.UserID, req.ID, c.ClientIP())
	if err != nil {
		respond(c, 1, "签发播放凭证失败", nil)
		return
	}
	respond(c, 0, "ok", gin.H{"ticket": ticket})
}

// parseSingleRange 解析形如 "bytes=start-end" 的单个 Range，返回实际 [start, end]（闭区间）。
// 不支持多段 Range；后缀范围 "bytes=-N" 返回最后 N 字节。
func parseSingleRange(hdr string, size int64) (int64, int64, bool) {
	hdr = strings.TrimPrefix(hdr, "bytes=")
	if hdr == "" || strings.Contains(hdr, ",") {
		return 0, 0, false
	}
	parts := strings.SplitN(hdr, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])
	var (
		start, end int64
		err        error
	)
	switch {
	case startStr == "" && endStr == "":
		return 0, 0, false
	case startStr == "":
		n, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil || n <= 0 {
			return 0, 0, false
		}
		if n > size {
			n = size
		}
		return size - n, size - 1, true
	case endStr == "":
		start, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil || start < 0 || start >= size {
			return 0, 0, false
		}
		return start, size - 1, true
	default:
		start, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		end, err = strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		if start < 0 || end < start || start >= size {
			return 0, 0, false
		}
		if end >= size {
			end = size - 1
		}
		return start, end, true
	}
}

// handleVideoPlay 基于短时票据 + Range 分片流式播放。
// 防下载措施：必须携带票据（绑定用户/视频/IP、短时有效）；必须携带 Range（拒绝整文件）；
// 单次响应限制大小（滴水式）；按会话校验拉取顺序（大跨度跳转需换新票据）。
func handleVideoPlay(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if id <= 0 {
		respond(c, 1, "无效的视频ID", nil)
		return
	}
	ticket := c.Query("ticket")
	if ticket == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 1, "msg": "缺少播放凭证", "data": nil})
		return
	}
	claims, err := parsePlayTicket(ticket)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 1, "msg": "播放凭证无效或已过期", "data": nil})
		return
	}
	if claims.VideoID != id || claims.IP != c.ClientIP() {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 1, "msg": "播放凭证不匹配", "data": nil})
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

	// HEAD 仅返回元信息，不涉及内容下载
	if c.Request.Method == http.MethodHead {
		http.ServeContent(c.Writer, c.Request, filepath.Base(path), st.ModTime(), f)
		return
	}

	// 必须携带 Range，禁止整文件拉取
	rangeHdr := c.Request.Header.Get("Range")
	if rangeHdr == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 1, "msg": "不支持整文件下载", "data": nil})
		return
	}
	start, end, ok := parseSingleRange(rangeHdr, st.Size())
	if !ok {
		c.AbortWithStatusJSON(http.StatusRequestedRangeNotSatisfiable, gin.H{"code": 1, "msg": "无效的 Range", "data": nil})
		return
	}
	// 单次响应限制大小（滴水式）：允许任意位置拉取，但单次最多 playStepBytes
	servedEnd := min(end, start+int64(playStepBytes)-1)
	c.Request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, servedEnd))
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
