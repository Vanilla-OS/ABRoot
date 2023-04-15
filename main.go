package main

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
	"embed"

	"github.com/vanilla-os/abroot/cmd"
	"github.com/vanilla-os/orchid/cmdr"
)

var (
	Version = "2.0.0-alpha.1"
)

//go:embed locales/*.yml
var fs embed.FS
var abroot *cmdr.App

func main() {
	abroot = cmd.New(Version, fs)

	// root command
	root := cmd.NewRootCommand(Version)
	abroot.CreateRootCommand(root)

	upgrade := cmd.NewUpgradeCommand()
	root.AddCommand(upgrade)

	kargs := cmd.NewKargsCommand()
	root.AddCommand(kargs)

	pkg := cmd.NewPkgCommand()
	root.AddCommand(pkg)

	rollback := cmd.NewRollbackCommand()
	root.AddCommand(rollback)

	status := cmd.NewStatusCommand()
	root.AddCommand(status)

	// run the app
	err := abroot.Run()
	if err != nil {
		cmdr.Error.Println(err)
	}
}
