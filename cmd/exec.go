package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/core"
)

func execUsage(*cobra.Command) error {
	fmt.Print(`Description:
	Execute a command in a transactional shell in the future root partition and switch to it on the next boot.

Usage:
	exec [command]

Options:
	--help/-h		show this message
	--assume-yes/-y		assume yes to all questions

Examples:
	abroot exec ls -l /
	abroot exec apt install git 
`)

	return nil
}

func NewExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a transactional shell in the future root and switch to it on next boot.",
		RunE:  execCommand,
	}
	cmd.SetUsageFunc(execUsage)
	cmd.Flags().BoolP("assume-yes", "y", false, "assume yes to all questions")
	cmd.Flags().SetInterspersed(false)

	return cmd
}

func execCommand(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	assumeYes, _ := cmd.Flags().GetBool("assume-yes")
	if !assumeYes {
		if !core.AskConfirmation(`
===============================================================================
PLEASE READ CAREFULLY BEFORE PROCEEDING
===============================================================================
Changes made in the shell will be applied to the future root on next boot on
successful.
Running a command in a transactional shell is meant to be used by advanced users 
for maintenance purposes.

If you ended up here trying to install an application, consider using 
Flatpak/Appimage or Apx (apx install pacakge) instead.

Read more about ABRoot at [https://documentation.vanillaos.org/docs/ABRoot/].

Are you sure you want to proceed?`) {
			return nil
		}
	}

	fmt.Println(`New transaction started. This may take a while...
Do not reboot or cancel the transaction until it is finished.`)

	command := ""
	for _, arg := range args {
		command += arg + " "
	}

	if _, err := core.TransactionalExec(command); err != nil {
		fmt.Println("Failed to start transactional shell:", err)
		os.Exit(1)
	}

	fmt.Println("Transaction completed successfully. Reboot to apply changes.")

	return nil
}
