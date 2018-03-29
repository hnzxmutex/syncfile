package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	BuildDate string
)

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "编译时间",
	Long:  "显示项目编译时间",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("BuildDate:%s\n", BuildDate)
	},
}
