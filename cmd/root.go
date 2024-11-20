package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version_flag = false

var rootCmd = &cobra.Command{
	Use:     "lvs2",
	Short:   "lvs2 is a Local Video Service",
	Long:    "lvs2 is a Local Video Service",
	Version: "v1.0.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
