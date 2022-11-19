package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/ABRoot/core"
)

func execUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	Execute a command in a transactional shell in the future root and switch to it on next boot.

Usage:
	exec [command]

Options:
	--help/-h		show this message

Examples:
	abroot exec ls -l /
`)
	return nil
}

func NewExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a transactional shell in the future root and switch to it on next boot.",
		RunE:  execCommand,
	}
	cmd.SetUsageFunc(execUsage)
	return cmd
}

func execCommand(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	return nil
}
