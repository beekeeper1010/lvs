package cmd

import (
	"fmt"
	"log"

	"github.com/beekeeper1010/lvs2/initialize"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run server",
	Long:  "Run server",
	Run: func(cmd *cobra.Command, args []string) {
		addr, _ := cmd.Flags().GetString("addr")
		dbfile, _ := cmd.Flags().GetString("db")
		cfgfile, _ := cmd.Flags().GetString("cfg")
		logfile, _ := cmd.Flags().GetString("log")
		run(addr, dbfile, cfgfile, logfile)
	},
}

func init() {
	runCmd.Flags().StringP("addr", "a", ":8080", "server listen address")
	runCmd.Flags().StringP("cfg", "c", "config.yaml", "config file for server")
	runCmd.Flags().StringP("db", "d", "lvs2.db", "sqlite db file for server")
	runCmd.Flags().StringP("log", "l", "lvs2.log", "log file for server")
	rootCmd.AddCommand(runCmd)
}

func printLogo() {
	fmt.Println(`
 _     __   __  ___   ___ 
| |    \ \ / / / __| |_  )
| |__   \ V /  \__ \  / /
|____|   \_/   |___/ /___|`)
}

func run(addr, dbfile, cfgfile, logfile string) {
	initialize.InitializeBase(dbfile, cfgfile, logfile)
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()
	g.Use(cors.Default())
	g.LoadHTMLGlob("web/templates/*")
	g.Static("/static", "web/assets")
	initialize.InitializeRouter(g)
	log.Println("server running on", addr)
	printLogo()
	log.Fatal(g.Run(addr))
}
