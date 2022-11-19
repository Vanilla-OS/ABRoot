package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/ABRoot/core"
)

func updateBootUsage(*cobra.Command) error {
	fmt.Print(`Description: 
	Update the boot partition.

Usage:
	_update-boot

Options:
	--help/-h		show this message`)
	return nil
}

func NewUpdateBootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "_update-boot",
		Short: "Update the boot partition",
		RunE:  status,
	}
	cmd.SetUsageFunc(updateBootUsage)
	return cmd
}

func status(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	return nil
}
