package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewUpdateBootCommand() *cmdr.Command {
	upd := cmdr.NewCommand(
		"_update-boot",
		abroot.Trans("update.long"),
		abroot.Trans("update.short"),
		status).WithBoolFlag(
		cmdr.NewBoolFlag(
			"force-update",
			"f",
			abroot.Trans("update.forceUpdateFlag"),
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

	forceUpdate := cmdr.FlagValBool("force-update")
	if !forceUpdate {
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

	if updateErr := core.UpdateRootBoot(false, kargs); updateErr != nil {
		return updateErr
	}

	return nil
}
