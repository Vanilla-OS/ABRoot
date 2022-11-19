package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/ABRoot/core"
)

func shellUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	Enter a transactional shell in the future root and switch root on next boot

Usage:
	shell

Options:
	--help/-h		show this message

Examples:
	abroot shell
`)
	return nil
}

func NewShellCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Enter a transactional shell in the future root and switch root on next boot",
		RunE:  shell,
	}
	cmd.SetUsageFunc(shellUsage)
	return cmd
}

func shell(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	return nil
}
