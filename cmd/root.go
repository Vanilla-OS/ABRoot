package cmd

import (
	"embed"

	"github.com/vanilla-os/orchid/cmdr"
)

var abroot *cmdr.App

const (
	verboseFlag string = "verbose"
)

func New(version string, fs embed.FS) *cmdr.App {
	abroot = cmdr.NewApp("abroot", version, fs)
	return abroot
}
func NewRootCommand(version string) *cmdr.Command {
	root := cmdr.NewCommand(
		abroot.Trans("abroot.use"),
		abroot.Trans("abroot.long"),
		abroot.Trans("abroot.short"),
		nil).
		WithPersistentBoolFlag(
			cmdr.NewBoolFlag(
				verboseFlag,
				"v",
				abroot.Trans("abroot.verboseFlag"),
				false))
	root.Version = version

	return root
}
