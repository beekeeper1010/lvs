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
		`INSERT INTO users(username, password, created_at) VALUES(?, ?, ?)`,
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

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
