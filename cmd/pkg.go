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

var validPkgArgs = []string{"add", "remove", "list"}

func NewPkgCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"pkg add|remove|list",
		abroot.Trans("pkg.long"),
		abroot.Trans("pkg.short"),
		pkg,
	)

	cmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	cmd.ValidArgs = validPkgArgs
	cmd.Example = "abroot pkg add <pkg>"

	return cmd
}

func pkg(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("pkg.rootRequired"))
		return nil
	}

	return nil
}
