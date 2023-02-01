package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewRollbackCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"rollback",
		abroot.Trans("rollback.long"),
		abroot.Trans("rollback.short"),
		rollbackCommand,
	)
	cmd.Example = "abroot rollback"
	cmd.Flags().SetInterspersed(false)

	return cmd
}

func rollbackCommand(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		cmdr.Error.Println(abroot.Trans("rollback.rootRequired"))
		return nil
	}

	err := core.Rollback()
	if err != nil {
		return err
	}

	return nil
}
