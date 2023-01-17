package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

const (
	assumeYesFlag string = "assume-yes"
)

func NewUpdateBootCommand() *cmdr.Command {
	upd := cmdr.NewCommand(
		"_update-boot",
		abroot.Trans("update.long"),
		abroot.Trans("update.short"),
		status).WithBoolFlag(
		cmdr.NewBoolFlag(
			assumeYesFlag,
			"y",
			abroot.Trans("update.assumeYesFlag"),
			false))
	// don't show this command in usage/help unless specified
	upd.Hidden = true
	return upd
}

func status(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("update.rootRequired"))
		return nil
	}
	assumeYes := cmdr.FlagValBool(assumeYesFlag)

	if !assumeYes {
		b, err := cmdr.Confirm.Show(abroot.Trans("update.confirm"))
		if err != nil {
			return err
		}
		if !b {
			return nil
		}
	}

	kargs, err := core.ReadKargsFile()
	if err != nil {
		return err
	}
	if update_err := core.UpdateRootBoot(false, kargs); update_err != nil {
		return update_err
	}

	return nil
}
