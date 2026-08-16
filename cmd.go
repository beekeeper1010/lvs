package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	port      int
	dbPath    string
	password  string
	username  string
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
	Short: "初始化数据库并创建管理员账号",
	Run: func(cmd *cobra.Command, args []string) {
		if password == "" {
			fmt.Println("用法: lvs init -p <密码> [-D <数据库>]")
			os.Exit(1)
		}
		if err := createDB(dbPath, password); err != nil {
			fmt.Println("init 失败:", err)
			os.Exit(1)
		}
		fmt.Printf("初始化完成: 数据库 %s, 管理员账号 \"admin\"\n", dbPath)
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "递归扫描目录下的 mp4 视频并生成缩略图入库",
	Run: func(cmd *cobra.Command, args []string) {
		if scanDir == "" {
			fmt.Println("用法: lvs scan -d <目录> [-D <数据库>] [-t <缩略图目录>]")
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
	Run:   rootCmd.Run, // 无子命令的 root 与 serve 等价
}

var resetPwdCmd = &cobra.Command{
	Use:   "resetpwd",
	Short: "重置用户密码",
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" || password == "" {
			fmt.Println("用法: lvs resetpwd -u <用户名> -p <新密码> [-D <数据库>]")
			os.Exit(1)
		}
		if err := openDB(dbPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer db.Close()
		if err := resetUserPassword(username, password); err != nil {
			fmt.Println("resetpwd 失败:", err)
			os.Exit(1)
		}
		fmt.Printf("已重置用户 %q 的密码\n", username)
	},
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&port, "port", "P", 8900, "监听端口")
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "D", "lvs.db", "sqlite 数据库文件路径")
	initCmd.Flags().StringVarP(&password, "password", "p", "", "管理员密码")
	scanCmd.Flags().StringVarP(&scanDir, "dir", "d", "", "要扫描的视频目录")
	scanCmd.Flags().StringVarP(&thumbsDir, "thumbs", "t", "data/thumbs", "缩略图输出目录")
	resetPwdCmd.Flags().StringVarP(&username, "user", "u", "", "要重置密码的用户名")
	resetPwdCmd.Flags().StringVarP(&password, "password", "p", "", "新密码")
	rootCmd.AddCommand(initCmd, scanCmd, serveCmd, resetPwdCmd)
}
