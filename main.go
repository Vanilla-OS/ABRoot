package main

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
	"embed"

	"github.com/containers/storage/pkg/reexec"
	"github.com/vanilla-os/abroot/cmd"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/orchid/cmdr"
)

var Version = "development"

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
	abroot.CreateRootCommand(root, abroot.Trans("abroot.msg.help"), abroot.Trans("abroot.msg.version"))

	msgs := cmdr.UsageStrings{
		Usage:                abroot.Trans("abroot.msg.usage"),
		Aliases:              abroot.Trans("abroot.msg.aliases"),
		Examples:             abroot.Trans("abroot.msg.examples"),
		AvailableCommands:    abroot.Trans("abroot.msg.availableCommands"),
		AdditionalCommands:   abroot.Trans("abroot.msg.additionalCommands"),
		Flags:                abroot.Trans("abroot.msg.flags"),
		GlobalFlags:          abroot.Trans("abroot.msg.globalFlags"),
		AdditionalHelpTopics: abroot.Trans("abroot.msg.additionalHelpTopics"),
		MoreInfo:             abroot.Trans("abroot.msg.moreInfo"),
	}
	abroot.SetUsageStrings(msgs)

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

	cnf := cmd.NewConfCommand()
	root.AddCommand(cnf)

	unlockVar := cmd.NewUnlockVarCommand()
	root.AddCommand(unlockVar)

	mntSys := cmd.NewMountSysCommand()
	root.AddCommand(mntSys)

	rebase := cmd.NewRebaseCommand()
	root.AddCommand(rebase)

	// run the app
	err := abroot.Run()
	if err != nil {
		cmdr.Error.Println(err)
	}
}
