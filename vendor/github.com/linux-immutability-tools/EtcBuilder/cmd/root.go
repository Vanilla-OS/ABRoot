package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "EtcBuilder",
	Short: "A tool to generate an etc based on multiple etc states",
}

func init() {
	rootCmd.AddCommand(NewBuildCommand())
}

func Execute() error {
	return rootCmd.Execute()
}
