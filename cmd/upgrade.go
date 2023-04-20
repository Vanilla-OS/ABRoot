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

func NewUpgradeCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"upgrade",
		abroot.Trans("upgrade.long"),
		abroot.Trans("upgrade.short"),
		upgrade,
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"check-only",
			"c",
			abroot.Trans("upgrade.checkOnlyFlag"),
			false))

	cmd.Example = "abroot upgrade"

	return cmd
}

func upgrade(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("upgrade.rootRequired"))
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

	if checkOnly {
		if aBsys.CheckUpdate() {
			cmdr.Info.Println(abroot.Trans("upgrade.updateAvailable"))
		} else {
			cmdr.Info.Println(abroot.Trans("upgrade.noUpdateAvailable"))
		}
		return nil
	}

	err = aBsys.Upgrade()
	if err != nil {
		cmdr.Error.Println(err)
		err = aBsys.RunCleanUpQueue()
		if err != nil {
			cmdr.Error.Println(err)
		}
		return err
	}

	return nil
}
