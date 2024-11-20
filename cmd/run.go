package cmd

import (
	"github.com/beekeeper1010/lvs2/server"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run server",
	Long:  "Run server",
	Run: func(cmd *cobra.Command, args []string) {
		addr, _ := cmd.Flags().GetString("addr")
		dbfile, _ := cmd.Flags().GetString("db")
		logfile, _ := cmd.Flags().GetString("log")
		server.Run(addr, dbfile, logfile)
	},
}

func init() {
	runCmd.Flags().StringP("addr", "a", ":8080", "server listen address")
	runCmd.Flags().StringP("db", "d", "lvs2.db", "sqlite db file for server")
	runCmd.Flags().StringP("log", "l", "lvs2.log", "log file for server")
	rootCmd.AddCommand(runCmd)
}
