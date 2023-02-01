package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewShellCommand() *cmdr.Command {
	shell := cmdr.NewCommand(
		"shell",
		abroot.Trans("shell.long"),
		abroot.Trans("shell.short"),
		shell,
	).WithBoolFlag(
		cmdr.NewBoolFlag(
			"force-open",
			"f",
			abroot.Trans("shell.forceOpenFlag"),
			false))
	shell.Example = "abroot shell"

	return shell
}

func shell(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("shell.rootRequired"))
		return nil
	}

	forceOpen := cmdr.FlagValBool("force-open")
	if !forceOpen {
		b, err := cmdr.Confirm.Show(abroot.Trans("shell.confirm"))

		if err != nil {
			return err
		}

		if !b {
			return nil
		}
	}

	cmdr.Warning.Println(abroot.Trans("shell.start"))

	if _, err := core.NewTransactionalShell(); err != nil {
		cmdr.Error.Println(abroot.Trans("shell.failed"), err)
		os.Exit(1)
	}

	core.TransactionDiff()

	cmdr.Success.Println(abroot.Trans("shell.success"))

	return nil
}
