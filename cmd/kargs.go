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

var validKargsArgs = []string{"edit", "list"}

func NewKargsCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"kargs edit|show",
		abroot.Trans("kargs.long"),
		abroot.Trans("kargs.short"),
		kargs,
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"check-only",
			"c",
			abroot.Trans("kargs.checkOnlyFlag"),
			false))

	cmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	cmd.ValidArgs = validKargsArgs
	cmd.Example = "abroot kargs edit"

	return cmd
}

func kargs(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("kargs.rootRequired"))
		return nil
	}

	return nil
}
