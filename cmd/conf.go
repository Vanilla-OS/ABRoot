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

func NewConfCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"config-editor",
		abroot.Trans("cnf.long"),
		abroot.Trans("cnf.short"),
		func(cmd *cobra.Command, args []string) error {
			err := cnf(cmd, args)
			if err != nil {
				os.Exit(1)
			}
			return nil
		},
	)

	return cmd
}

func cnf(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("cnf.rootRequired"))
		return nil
	}

	result, err := core.ConfEdit()
	switch result {
	case core.CONF_CHANGED:
		cmdr.Info.Println(abroot.Trans("cnf.changed"))
	case core.CONF_UNCHANGED:
		cmdr.Info.Println(abroot.Trans("cnf.unchanged"))
	case core.CONF_FAILED:
		cmdr.Error.Println(abroot.Trans("cnf.failed", err))
	}

	return nil
}
