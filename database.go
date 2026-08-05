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

// createDB 创建数据库文件、建表、内置账号并生成 JWT secret
func createDB(dbPath, username, password string) error {
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
		`INSERT INTO users(username, password, role, created_at) VALUES(?, ?, 'admin', ?)`,
		username, string(hash), time.Now()); err != nil {
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
	// 迁移：旧库补充 role 列，并确保 admin 为管理员
	_, _ = db.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'`)
	_, _ = db.Exec(`UPDATE users SET role = 'admin' WHERE username = 'admin'`)
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
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS videos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			thumb_path TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL
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

// upsertUser 创建用户或更新密码, 返回是否新建。admin 用户名创建即为管理员
func upsertUser(username, password string) (bool, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, username).Scan(&count); err != nil {
		return false, err
	}
	created := count == 0
	if created {
		role := "user"
		if username == "admin" {
			role = "admin"
		}
		_, err = db.Exec(`INSERT INTO users(username, password, role, created_at) VALUES(?, ?, ?, ?)`,
			username, string(hash), role, time.Now())
	} else {
		_, err = db.Exec(`UPDATE users SET password = ? WHERE username = ?`, string(hash), username)
	}
	if err != nil {
		return false, err
	}
	return created, nil
}

// userInfo 用户列表项
type userInfo struct {
	ID        int64
	Username  string
	Role      string
	CreatedAt string
}

// listUsers 查询全部用户
func listUsers() ([]userInfo, error) {
	rows, err := db.Query(`SELECT id, username, role, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]userInfo, 0)
	for rows.Next() {
		var u userInfo
		var created time.Time
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &created); err != nil {
			return nil, err
		}
		u.CreatedAt = created.Format("2006-01-02 15:04:05")
		list = append(list, u)
	}
	return list, rows.Err()
}

// deleteUser 删除用户, 至少保留一个账号
func deleteUser(username string) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count <= 1 {
		return fmt.Errorf("至少保留一个用户, 删除被拒绝")
	}
	res, err := db.Exec(`DELETE FROM users WHERE username = ?`, username)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("用户 %q 不存在", username)
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
