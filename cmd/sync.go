package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/ABRoot/core"
)

func syncUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	Sync the future root with the present root.

Usage:
	_sync-future

Options:
	--help/-h		show this message

Examples:
	abroot sync
`)
	return nil
}

func NewSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "_sync-future",
		Short: "Sync the future root with the present root",
		RunE:  sync,
	}
	cmd.SetUsageFunc(syncUsage)
	return cmd
}

func sync(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}
	return nil
}
