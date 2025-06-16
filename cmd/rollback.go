package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewRollbackCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"rollback",
		abroot.Trans("rollback.long"),
		abroot.Trans("rollback.short"),
		func(cmd *cobra.Command, args []string) error {
			err := rollback(cmd, args)
			if err != nil {
				os.Exit(1)
			}
			return nil
		},
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"check-only",
			"c",
			abroot.Trans("rollback.checkOnlyFlag"),
			false))

	cmd.Example = "abroot rollback"

	return cmd
}

func rollback(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("rollback.rootRequired"))
		return nil
	}

	checkOnly, err := cmd.Flags().GetBool("check-only")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	aBsys, err := core.NewABSystem()
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	response, err := aBsys.Rollback(checkOnly)
	if err != nil {
		cmdr.Error.Println(err)
		os.Exit(2)
		return err
	}
	switch response {
	case core.ROLLBACK_RES_YES:
		// NOTE: the following strings could lead to misinterpretation, with
		// "can" and "cannot", we don't mean "is it possible to rollback?",
		// but "is it necessary to rollback?"
		cmdr.Info.Println(abroot.Trans("rollback.canRollback"))
		os.Exit(0)
	case core.ROLLBACK_RES_NO:
		cmdr.Info.Println(abroot.Trans("rollback.cannotRollback"))
		os.Exit(1)
	case core.ROLLBACK_UNNECESSARY:
		cmdr.Info.Println(abroot.Trans("rollback.rollbackUnnecessary"))
		os.Exit(0)
	case core.ROLLBACK_SUCCESS:
		cmdr.Info.Println(abroot.Trans("rollback.rollbackSuccess"))
		os.Exit(0)
	}

	return nil
}
