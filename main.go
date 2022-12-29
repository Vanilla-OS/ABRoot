package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/cmd"
	"github.com/vanilla-os/abroot/core"
)

var (
	Version = "1.0.2"
)

func help(cmd *cobra.Command, args []string) {
	fmt.Print(`Usage: 
abroot [options] [command]

Options:
	--help/-h		show this message
	--verbose/-v		show more verbosity
	--version/-V		show version

Commands:
	_update-boot		update the boot partition (for advanced users only)
	get			outputs the present or future root partition state
	shell			enter a transactional shell in the future root partition and switch root on the next boot
	exec			execute a command in a transactional shell in the future root partition and switch to it on the next boot
`)
}

func newABRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "abroot",
		Short:   "ABRoot provides full immutability and atomicity by performing transactions between 2 root partitions (A<->B).",
		Version: Version,
	}
}

func main() {
	rootCmd := newABRootCommand()

	rootCmd.AddCommand(cmd.NewUpdateBootCommand())
	rootCmd.AddCommand(cmd.NewGetCommand())
	rootCmd.AddCommand(cmd.NewShellCommand())
	rootCmd.AddCommand(cmd.NewExecCommand())
	rootCmd.SetHelpFunc(help)
	rootCmd.Execute()

	core.CheckABRequirements()
}
