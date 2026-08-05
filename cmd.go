package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	port      int
	dbPath    string
	username  string
	password  string
	nickname  string
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
			fmt.Println("用法: lvs init -p <密码> [-d <数据库>]")
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

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "用户管理",
}

var userUpsertCmd = &cobra.Command{
	Use:   "upsert",
	Short: "创建用户或修改密码",
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" {
			fmt.Println("用法: lvs user upsert -u <用户名> (-p <密码> | -n <昵称> | 两者) [-d <数据库>]")
			os.Exit(1)
		}
		setPassword := cmd.Flags().Changed("password")
		setNickname := cmd.Flags().Changed("nickname")
		if !setPassword && !setNickname {
			fmt.Println("错误: 至少要指定 -p <密码> 或 -n <昵称> 之一")
			os.Exit(1)
		}
		if err := openDB(dbPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer db.Close()
		created, err := upsertUser(username, password, nickname, setPassword, setNickname)
		if err != nil {
			fmt.Println("操作失败:", err)
			os.Exit(1)
		}
		if created {
			fmt.Printf("用户 %q 已创建\n", username)
		} else {
			changed := make([]string, 0, 2)
			if setPassword {
				changed = append(changed, "密码")
			}
			if setNickname {
				changed = append(changed, "昵称")
			}
			fmt.Printf("用户 %q 的%s已更新\n", username, strings.Join(changed, "、"))
		}
	},
}

var userDelCmd = &cobra.Command{
	Use:   "del",
	Short: "删除用户",
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" {
			fmt.Println("用法: lvs user del -u <用户名> [-d <数据库>]")
			os.Exit(1)
		}
		if err := openDB(dbPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer db.Close()
		if err := deleteUser(username); err != nil {
			fmt.Println("操作失败:", err)
			os.Exit(1)
		}
		fmt.Printf("用户 %q 已删除\n", username)
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询用户列表",
	Run: func(cmd *cobra.Command, args []string) {
		if err := openDB(dbPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer db.Close()
		users, err := listUsers()
		if err != nil {
			fmt.Println("查询失败:", err)
			os.Exit(1)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t用户名\t昵称\t角色\t创建时间")
		for _, u := range users {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", u.ID, u.Username, u.Nickname, u.Role, u.CreatedAt)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&port, "port", "P", 8900, "监听端口")
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", "lvs.db", "sqlite 数据库文件路径")
	initCmd.Flags().StringVarP(&password, "password", "p", "", "管理员密码")
	scanCmd.Flags().StringVarP(&scanDir, "dir", "D", "", "要扫描的视频目录")
	scanCmd.Flags().StringVarP(&thumbsDir, "thumbs", "t", "data/thumbs", "缩略图输出目录")
	userUpsertCmd.Flags().StringVarP(&username, "username", "u", "", "用户名")
	userUpsertCmd.Flags().StringVarP(&password, "password", "p", "", "密码")
	userUpsertCmd.Flags().StringVarP(&nickname, "nickname", "n", "", "昵称")
	userDelCmd.Flags().StringVarP(&username, "username", "u", "", "用户名")
	userCmd.AddCommand(userUpsertCmd, userDelCmd, userListCmd)
	rootCmd.AddCommand(initCmd, scanCmd, serveCmd, userCmd)
}
