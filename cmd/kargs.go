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
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

var validKargsArgs = []string{"edit", "show"}

func NewKargsCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"kargs edit|show",
		abroot.Trans("kargs.long"),
		abroot.Trans("kargs.short"),
		func(cmd *cobra.Command, args []string) error {
			err := kargs(cmd, args)
			if err != nil {
				os.Exit(1)
			}
			return nil
		},
	)

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

	switch args[0] {
	case "edit":
		changed, err := core.KargsEdit()
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}

		if !changed {
			cmdr.Info.Println(abroot.Trans("kargs.notChanged"))
			return nil
		}

		aBsys, err := core.NewABSystem()
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}
		err = aBsys.RunOperation(core.APPLY, false)
		if err != nil {
			cmdr.Error.Println(abroot.Trans("pkg.applyFailed"))
			return err
		}
	case "show":
		kargsStr, err := core.KargsRead()
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}
		cmdr.Info.Println(kargsStr)
	default:
		return errors.New(abroot.Trans("kargs.unknownCommand", args[0]))
	}

	return nil
}
