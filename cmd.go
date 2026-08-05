package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func runInit(args []string) {
	fs := pflag.NewFlagSet("init", pflag.ContinueOnError)
	username := fs.String("username", "", "登录用户名")
	password := fs.String("password", "", "登录密码")
	dbPath := fs.String("db", "lvs.db", "sqlite 数据库文件路径")
	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *username == "" || *password == "" {
		fmt.Println("用法: lvs init --username <用户名> --password <密码> [--db <路径>]")
		os.Exit(1)
	}
	if err := createDB(*dbPath, *username, *password); err != nil {
		fmt.Println("init 失败:", err)
		os.Exit(1)
	}
	fmt.Printf("初始化完成: 数据库 %s, 账号 %q\n", *dbPath, *username)
}

func runScan(args []string) {
	fs := pflag.NewFlagSet("scan", pflag.ContinueOnError)
	dir := fs.String("dir", "", "要扫描的视频目录")
	dbPath := fs.String("db", "lvs.db", "sqlite 数据库文件路径")
	thumbsDir := fs.String("thumbs", "data/thumbs", "缩略图输出目录")
	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *dir == "" {
		fmt.Println("用法: lvs scan --dir <目录> [--db <路径>] [--thumbs <目录>]")
		os.Exit(1)
	}
	if err := openDB(*dbPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()
	if err := scanDirectory(*dir, *thumbsDir); err != nil {
		fmt.Println("scan 失败:", err)
		os.Exit(1)
	}
}

func runServe(args []string) {
	fs := pflag.NewFlagSet("serve", pflag.ContinueOnError)
	port := fs.Int("port", 8080, "监听端口")
	dbPath := fs.String("db", "lvs.db", "sqlite 数据库文件路径")
	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	startServer(*port, *dbPath)
}
