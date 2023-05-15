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
	"errors"
	"strings"

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

	cmd.Args = cobra.MinimumNArgs(1)
	cmd.ValidArgs = validPkgArgs
	cmd.Example = "abroot pkg add <pkg>"

	return cmd
}

func pkg(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("pkg.rootRequired"))
		return nil
	}

	pkgM := core.NewPackageManager()

	switch args[0] {
	case "add":
		if len(args) < 2 {
			return errors.New(abroot.Trans("pkg.noPackageNameProvided"))
		}
		for _, pkg := range args[1:] {
			err := pkgM.Add(pkg)
			if err != nil {
				cmdr.Error.Println(err)
				return err
			}
		}
		cmdr.Info.Printf(abroot.Trans("pkg.addedMsg"), strings.Join(args[1:], ", "))
	case "remove":
		if len(args) < 2 {
			return errors.New(abroot.Trans("pkg.noPackageNameProvided"))
		}
		for _, pkg := range args[1:] {
			err := pkgM.Remove(pkg)
			if err != nil {
				cmdr.Error.Println(err)
				return err
			}
		}
		cmdr.Info.Printf(abroot.Trans("pkg.removedMsg"), strings.Join(args[1:], ", "))
	case "list":
		added, err := pkgM.GetAddPackagesString("\n")
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}

		removed, err := pkgM.GetRemovePackagesString("\n")
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}

		cmdr.Info.Printf(abroot.Trans("pkg.listMsg"), added, removed)
		return nil
	}

	return nil
}
