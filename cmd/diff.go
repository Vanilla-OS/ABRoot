package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/core"
)

func diffUsage(*cobra.Command) error {
	fmt.Print(`Description:
	List modifications made to the filesystem in the latest transiction.

Usage:
	diff

Options:
	--help/-h		show this message

Examples:
	abroot diff
`)

	return nil
}

func NewDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "List modifications made to the filesystem in the latest transaction.",
		RunE:  diffCommand,
	}
	cmd.SetUsageFunc(diffUsage)
	cmd.Flags().SetInterspersed(false)

	return cmd
}

func diffCommand(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	core.TransactionDiff()

	return nil
}
