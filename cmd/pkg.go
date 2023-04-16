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
	"fmt"

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
		err := pkgM.Add(args[1])
		if err != nil {
			cmdr.Error.Println(err)
			return err
		} else {
			cmdr.Info.Printf("Package %s added\n", args[1])
		}
	case "remove":
		err := pkgM.Remove(args[1])
		if err != nil {
			cmdr.Error.Println(err)
			return err
		} else {
			cmdr.Info.Printf("Package %s removed\n", args[1])
		}
	case "list":
		added, err := pkgM.GetAddPackages()
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}

		removed, err := pkgM.GetRemovePackages()
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}

		addedStr := ""
		for _, pkg := range added {
			addedStr += fmt.Sprintf("%s\n", pkg)
		}

		removedStr := ""
		for _, pkg := range removed {
			removedStr += fmt.Sprintf("%s\n", pkg)
		}

		cmdr.Info.Printf("Added packages:\n%s\nRemoved packages:\n%s\n", addedStr, removedStr)
		return nil
	}

	return nil
}
