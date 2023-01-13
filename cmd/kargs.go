package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/core"
)

func kargsUsage(*cobra.Command) error {
	fmt.Print(`Description:
	Manage kernel parameters.

Usage:
	kargs [action]

Options:
	--help/-h				show this message

Actions:
	get [present|future] 	get present/future root partition parameters
	edit

Examples:
	abroot kargs edit
	abroot kargs get future
`)

	return nil
}

func NewKargsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kargs",
		Short: "Manage kernel parameters.",
		RunE:  kargsCommand,
	}
	cmd.SetUsageFunc(kargsUsage)
	cmd.Flags().SetInterspersed(false)

	return cmd
}

func kargsCommand(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	if args[0] == "get" && len(args) == 1 {
		fmt.Println("Please specify a state (present or future)")
		return nil
	}

	switch args[0] {
	case "get":
		switch args[1] {
		case "present":
			kargs, err := core.GetCurrentKargs()
			if err != nil {
				return err
			}

			fmt.Printf("Current partition's parameters:\n%s\n", kargs)
		case "future":
			kargs, err := core.GetFutureKargs()
			if err != nil {
				return err
			}

			fmt.Printf("Future partition's parameters:\n%s\n", kargs)
		default:
			fmt.Printf("Unknown state: %s\n", args[1])
		}
	case "edit":
		kargs_edit()
	default:
		fmt.Printf("Unknown parameter: %s\n", args[0])
	}

	return nil
}

func kargs_edit() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	// Create custom kargs file if non-existent
	if _, err := os.Stat(core.KargsPath); os.IsNotExist(err) {
		cmd := exec.Command("cp", core.KargsDefaultPath, core.KargsPath)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Open a temporary file, so editors installed via apx can also be used
	cmd := exec.Command("cp", core.KargsPath, "/tmp/kargs-temp")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Call $EDITOR on temp file
	cmd = exec.Command(editor, "/tmp/kargs-temp")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Copy temp file back to /etc
	cmd = exec.Command("cp", "/tmp/kargs-temp", core.KargsPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Run something just to trigger a transaction
	if _, err := core.TransactionalExec("echo"); err != nil {
		fmt.Println("Failed to start transactional shell:", err)
		os.Exit(1)
	}

	fmt.Println("Kernel parameters will be applied on next boot.")
	return nil
}
