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
	--assume-yes/-y		assume yes to all questions

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

	assumeYes, _ := cmd.Flags().GetBool("assume-yes")
	if !assumeYes {
		if !core.AskConfirmation(`Are you sure you want to proceed?
The transactional shell is meant to be used by advanced users for maintenance purposes.`) {
			return nil
		}
	}

	if _, err := core.NewTransactionalShell(); err != nil {
		return err
	}

	return nil
}
