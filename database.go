package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var (
	db        *sql.DB
	jwtSecret []byte
)

// createDB 创建数据库文件、建表、内置 admin 账号并生成 JWT secret
func createDB(dbPath, password string) error {
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("数据库 %q 已存在", dbPath)
	}
	d, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer d.Close()
	if err := createTables(d); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := d.Exec(
		`INSERT INTO users(username, password, nickname, role, created_at) VALUES('admin', ?, '管理员', 'admin', ?)`,
		string(hash), time.Now()); err != nil {
		return err
	}
	if _, err := d.Exec(
		`INSERT INTO settings(key, value) VALUES('jwt_secret', ?)`,
		randomHex(32)); err != nil {
		return err
	}
	return nil
}

// openDB 打开现有数据库并加载 JWT secret
func openDB(dbPath string) error {
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("数据库 %q 不存在, 请先运行 'lvs init'", dbPath)
	}
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1) // sqlite 单写者，限制并发连接
	var s string
	if err := db.QueryRow(`SELECT value FROM settings WHERE key='jwt_secret'`).Scan(&s); err != nil {
		return fmt.Errorf("读取 jwt_secret 失败: %w", err)
	}
	jwtSecret = []byte(s)
	return nil
}

func createTables(d *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			nickname TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL DEFAULT 'user',
			pwd_ver INTEGER NOT NULL DEFAULT 0,
			avatar TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS videos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			thumb_path TEXT NOT NULL DEFAULT '',
			duration REAL NOT NULL DEFAULT 0,
			like_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL
		)`,
		// 点赞记录：联合主键保证每个用户对同一视频最多点赞一次
		`CREATE TABLE IF NOT EXISTS video_likes (
			video_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (video_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	}
	for _, s := range stmts {
		if _, err := d.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

// createUser 创建用户（用户名重复时返回错误），返回新用户 id
func createUser(username, password, nickname string) (int64, error) {
	if nickname == "" {
		nickname = username
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	role := "user"
	if username == "admin" {
		role = "admin"
	}
	res, err := db.Exec(`INSERT INTO users(username, password, nickname, role, created_at) VALUES(?, ?, ?, ?, ?)`,
		username, string(hash), nickname, role, time.Now())
	if err != nil {
		return 0, fmt.Errorf("创建用户失败: %w", err)
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// updateUser 按 id 更新昵称, 密码非空时同时更新密码
func updateUser(id int64, nickname, password string) error {
	// 昵称为空时不覆盖原值
	update := `UPDATE users SET nickname = CASE WHEN ? = '' THEN nickname ELSE ? END`
	args := []any{nickname, nickname}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		update += `, password = ?, pwd_ver = pwd_ver + 1` // 改密使旧 token 失效
		args = append(args, string(hash))
	}
	update += ` WHERE id = ?`
	args = append(args, id)
	_, err := db.Exec(update, args...)
	return err
}

// resetUserPassword 重置指定用户的密码，改密使旧 token 失效
func resetUserPassword(username, newPassword string) error {
	var id int64
	if err := db.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id); err != nil {
		return fmt.Errorf("用户 %q 不存在", username)
	}
	return updateUser(id, "", newPassword)
}

// userInfo 用户列表项
type userInfo struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Role      string `json:"role"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"created_at"`
}

// listUsers 查询全部用户
func listUsers() ([]userInfo, error) {
	rows, err := db.Query(`SELECT id, username, nickname, role, avatar, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]userInfo, 0)
	for rows.Next() {
		var u userInfo
		var created time.Time
		if err := rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.Role, &u.Avatar, &created); err != nil {
			return nil, err
		}
		u.CreatedAt = created.Format("2006-01-02 15:04:05")
		list = append(list, u)
	}
	return list, rows.Err()
}

// deleteUserByID 按 id 删除用户，管理员不可删且至少保留一个账号
func deleteUserByID(id int64) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count <= 1 {
		return fmt.Errorf("至少保留一个用户, 删除被拒绝")
	}
	var role string
	if err := db.QueryRow(`SELECT role FROM users WHERE id = ?`, id).Scan(&role); err != nil {
		return fmt.Errorf("用户不存在")
	}
	if role == "admin" {
		return fmt.Errorf("管理员账号不可删除")
	}
	if _, err := db.Exec(`DELETE FROM users WHERE id = ?`, id); err != nil {
		return err
	}
	return nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
