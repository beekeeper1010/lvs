package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/beekeeper1010/lvs2/server"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management",
	Long:  "User management",
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user",
	Long:  "Add a new user",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		nickname, _ := cmd.Flags().GetString("nickname")
		password, _ := cmd.Flags().GetString("password")
		admin, _ := cmd.Flags().GetBool("admin")
		dbfile, _ := cmd.Flags().GetString("db")
		if err := addUser(username, nickname, password, admin, dbfile); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("ok")
	},
}

var userDelCmd = &cobra.Command{
	Use:   "del",
	Short: "Delete a user",
	Long:  "Delete a user",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		dbfile, _ := cmd.Flags().GetString("db")
		if err := delUser(username, dbfile); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("ok")
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Long:  "List all users",
	Run: func(cmd *cobra.Command, args []string) {
		dbfile, _ := cmd.Flags().GetString("db")
		if err := listUsers(dbfile); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	/* userAddCmd */
	userAddCmd.Flags().BoolP("admin", "a", false, "is admin")
	userAddCmd.Flags().StringP("db", "d", "lvs2.db", "sqlite db file")
	userAddCmd.Flags().StringP("nickname", "n", "", "nickname")
	userAddCmd.MarkFlagRequired("nickname")
	userAddCmd.Flags().StringP("password", "p", "", "password")
	userAddCmd.MarkFlagRequired("password")
	userAddCmd.Flags().StringP("username", "u", "", "username to add")
	userAddCmd.MarkFlagRequired("username")
	/* userDelCmd */
	userDelCmd.Flags().StringP("db", "d", "lvs2.db", "sqlite db file")
	userDelCmd.Flags().StringP("username", "u", "", "username to delete")
	userDelCmd.MarkFlagRequired("username")
	/* userListCmd */
	userListCmd.Flags().StringP("db", "d", "lvs2.db", "sqlite db file")
	/* userCmd */
	userCmd.AddCommand(userAddCmd, userDelCmd, userListCmd)
	rootCmd.AddCommand(userCmd)
}

func addUser(username, nickname, password string, admin bool, dbfile string) error {
	if err := server.InitializeDbAndTable(dbfile); err != nil {
		return err
	}
	pwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return server.DB.Create(&server.User{
		Username: username,
		Nickname: nickname,
		Password: string(pwd),
		Admin:    admin,
	}).Error
}

func delUser(username, dbfile string) error {
	if err := server.InitializeDbAndTable(dbfile); err != nil {
		return err
	}
	result := server.DB.Unscoped().Where("username = ?", username).Delete(&server.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user '%s' not found", username)
	}
	return nil
}

func listUsers(dbfile string) error {
	if err := server.InitializeDbAndTable(dbfile); err != nil {
		return err
	}
	var users []server.User
	if err := server.DB.Find(&users).Error; err != nil {
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"No.", "Username", "Nickname", "Password", "Admin"})
	for i, user := range users {
		admin := "N"
		if user.Admin {
			admin = "Y"
		}
		table.Append([]string{strconv.Itoa(i + 1), user.Username, user.Nickname, user.Password, admin})
	}
	table.Render()
	return nil
}
