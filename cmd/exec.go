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
	--assume-yes/-y		assume yes to all questions

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

	assumeYes, _ := cmd.Flags().GetBool("assume-yes")
	if !assumeYes {
		if !core.AskConfirmation(`Are you sure you want to proceed?
Running a command in a transactional shell is meant to be used by advanced users for maintenance purposes.`) {
			return nil
		}
	}

	command := args[0]
	if _, err := core.TransactionalExec(command); err != nil {
		return err
	}

	return nil
}
