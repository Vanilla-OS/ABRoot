package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewKargsCommand() *cmdr.Command {
	kargs := cmdr.NewCommand("kargs [edit|get] <partition>", abroot.Trans("kargs.long"), abroot.Trans("kargs.short"), kargsCommand)
	kargs.Example = "abroot kargs edit\nabroot kargs get future"

	return kargs
}

func kargsCommand(cmd *cobra.Command, args []string) error {
	if args[0] == "get" && len(args) == 1 {
		cmdr.Error.Println(abroot.Trans("kargs.stateRequired"))
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

			cmdr.Info.Printf(abroot.Trans("kargs.params"), kargs)
		case "future":
			kargs, err := core.GetFutureKargs()
			if err != nil {
				return err
			}

			cmdr.Info.Printf(abroot.Trans("kargs.futureParams"), kargs)
		default:
			cmdr.Error.Printf(abroot.Trans("kargs.unknownState"), args[1])
		}
	case "edit":
		err := kargsEdit()
		if err != nil {
			return nil
		}
	default:
		cmdr.Error.Printf(abroot.Trans("kargs.unknownParam"), args[0])
	}

	return nil
}

func kargsEdit() error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("kargs.rootRequired"))
		return nil
	}

	if core.AreTransactionsLocked() {
		cmdr.Error.Printf(abroot.Trans("kargs.transactionsLocked"))
		return nil
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	// Create custom kargs file if non-existent
	if _, err := os.Stat(core.KargsPath); os.IsNotExist(err) {
		cmd := exec.Command("cp", core.KargsDefaultPath, core.KargsPath)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	// Open a temporary file, so editors installed via apx can also be used
	cmd := exec.Command("cp", core.KargsPath, "/tmp/kargs-temp")

	err := cmd.Run()
	if err != nil {
		return err
	}

	// Call $EDITOR on temp file
	cmd = exec.Command(editor, "/tmp/kargs-temp")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	// Copy temp file back to /etc
	cmd = exec.Command("cp", "/tmp/kargs-temp", core.KargsPath)

	err = cmd.Run()
	if err != nil {
		return err
	}

	// Run something just to trigger a transaction
	if _, err := core.TransactionalExec("echo"); err != nil {
		cmdr.Error.Println(abroot.Trans("kargs.failedTransaction"), err)
		os.Exit(1)
	}

	cmdr.Info.Println(abroot.Trans("kargs.nextReboot"))

	return nil
}
