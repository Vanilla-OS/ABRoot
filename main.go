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

	"github.com/containers/storage/pkg/reexec"
	"github.com/vanilla-os/abroot/cmd"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/orchid/cmdr"
)

var (
	Version = "2.0.0-alpha.1"
)

//go:embed locales/*.yml
var fs embed.FS
var abroot *cmdr.App

func main() {
	if reexec.Init() {
		return
	}

	abroot = cmd.New(Version, fs)

	// root command
	root := cmd.NewRootCommand(Version)
	abroot.CreateRootCommand(root)

	upgrade := cmd.NewUpgradeCommand()
	root.AddCommand(upgrade)

	kargs := cmd.NewKargsCommand()
	root.AddCommand(kargs)

	// we only add the pkg command if the ABRoot configuration
	// has the IPkgMng enabled in any way (1 or 2)
	if settings.Cnf.IPkgMngStatus > 0 {
		pkg := cmd.NewPkgCommand()
		root.AddCommand(pkg)
	}

	rollback := cmd.NewRollbackCommand()
	root.AddCommand(rollback)

	status := cmd.NewStatusCommand()
	root.AddCommand(status)

	updateInitramfs := cmd.NewUpdateInitfsCommand()
	root.AddCommand(updateInitramfs)

	// run the app
	err := abroot.Run()
	if err != nil {
		cmdr.Error.Println(err)
	}
}
