package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{Use: "syncfile"}

func Execute() {
	rootCmd.AddCommand(clientCommand, serverCommand, versionCommand)
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
	}
}
