package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

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
		rollback,
	)

	cmd.Example = "abroot rollback"

	return cmd
}

func rollback(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("rollback.rootRequired"))
		return nil
	}

	return nil
}
