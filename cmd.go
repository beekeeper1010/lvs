package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	port      int
	dbPath    string
	username  string
	password  string
	scanDir   string
	thumbsDir string
)

var rootCmd = &cobra.Command{
	Use:   "lvs",
	Short: "本地视频播放服务",
	Long:  "基于 gin + jwt + sqlite 的本地视频播放服务，前端产物嵌入单文件部署",
	// 无子命令时等价于 serve
	Run: func(cmd *cobra.Command, args []string) {
		startServer(port, dbPath)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化数据库并创建登录账号",
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" || password == "" {
			fmt.Println("用法: lvs init --username <用户名> --password <密码> [--db <路径>]")
			os.Exit(1)
		}
		if err := createDB(dbPath, username, password); err != nil {
			fmt.Println("init 失败:", err)
			os.Exit(1)
		}
		fmt.Printf("初始化完成: 数据库 %s, 账号 %q\n", dbPath, username)
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "递归扫描目录下的 mp4 视频并生成缩略图入库",
	Run: func(cmd *cobra.Command, args []string) {
		if scanDir == "" {
			fmt.Println("用法: lvs scan --dir <目录> [--db <路径>] [--thumbs <目录>]")
			os.Exit(1)
		}
		if err := openDB(dbPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer db.Close()
		if err := scanDirectory(scanDir, thumbsDir); err != nil {
			fmt.Println("scan 失败:", err)
			os.Exit(1)
		}
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动服务",
	Run: func(cmd *cobra.Command, args []string) {
		startServer(port, dbPath)
	},
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&port, "port", "P", 8900, "监听端口")
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", "lvs.db", "sqlite 数据库文件路径")
	initCmd.Flags().StringVarP(&username, "username", "u", "", "登录用户名")
	initCmd.Flags().StringVarP(&password, "password", "p", "", "登录密码")
	scanCmd.Flags().StringVarP(&scanDir, "dir", "D", "", "要扫描的视频目录")
	scanCmd.Flags().StringVarP(&thumbsDir, "thumbs", "t", "data/thumbs", "缩略图输出目录")
	rootCmd.AddCommand(initCmd, scanCmd, serveCmd)
}
