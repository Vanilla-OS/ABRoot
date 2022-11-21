package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/ABRoot/core"
)

func getUsage(*cobra.Command) error {
	fmt.Print(`Description:
	Get the present or future root partition.

Usage:
	get [partition]

Options:
	--help/-h		show this message

Partition:
	present			get the present root partition
	future			get the future root partition

Examples:
	abroot get present
	abroot get future`)

	return nil
}

func NewGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get the present or future root partition",
		RunE:  get,
	}
	cmd.SetUsageFunc(getUsage)

	return cmd
}

func get(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	return nil
}
