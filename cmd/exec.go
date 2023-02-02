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
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"skip-diff",
			"s",
			abroot.Trans("exec.skipDiffFlag"),
			false))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"force-run",
			"f",
			abroot.Trans("exec.forceRunFlag"),
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

	forceRun := cmdr.FlagValBool("force-run")
	if !forceRun {
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

	skipDiff := cmdr.FlagValBool("skip-diff")
	if !skipDiff {
		core.TransactionDiff()
	}

	cmdr.Success.Println(abroot.Trans("exec.success"))

	return nil
}
