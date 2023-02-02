package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewDiffCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"diff",
		abroot.Trans("diff.long"),
		abroot.Trans("diff.short"),
		diffCommand,
	)
	cmd.Example = "abroot diff"
	cmd.Flags().SetInterspersed(false)

	return cmd
}

func diffCommand(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		cmdr.Error.Println(abroot.Trans("diff.rootRequired"))
		return nil
	}

	core.TransactionDiff()

	return nil
}
