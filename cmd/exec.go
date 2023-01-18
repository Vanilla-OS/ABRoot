package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewExecCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"exec [command]",
		abroot.Trans("exec.long"),
		abroot.Trans("exec.short"),
		execCommand,
	).WithBoolFlag(
		cmdr.NewBoolFlag(
			assumeYesFlag,
			"y",
			abroot.Trans("exec.assumeYesFlag"),
			false))
	cmd.Args = cobra.MinimumNArgs(1)
	cmd.Example = "abroot exec apt-get update"
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func execCommand(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("exec.rootRequired"))
		return nil
	}

	assumeYes := cmdr.FlagValBool(assumeYesFlag)
	if !assumeYes {
		b, err := cmdr.Confirm.Show(abroot.Trans("exec.confirm"))
		if err != nil {
			return err
		}
		if !b {
			return nil
		}
	}

	cmdr.Warning.Println(abroot.Trans("exec.start"))

	command := ""
	for _, arg := range args {
		command += arg + " "
	}

	if _, err := core.TransactionalExec(command); err != nil {
		cmdr.Error.Println(abroot.Trans("exec.failed"), err)
		os.Exit(1)
	}
	core.TransactionDiff()

	cmdr.Success.Println(abroot.Trans("exec.success"))

	return nil
}
