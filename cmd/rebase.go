package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewRebaseCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"rebase <name>",
		abroot.Trans("rebase.long"),
		abroot.Trans("rebase.short"),
		rebase,
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"dry-run",
			"d",
			abroot.Trans("rebase.dryRunFlag"),
			false,
		))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"remove-packages",
			"r",
			abroot.Trans("rebase.removePackagesShort"),
			false,
		))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"keep-packages",
			"k",
			abroot.Trans("rebase.keepPackages"),
			false,
		))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"rebase-only",
			"n",
			abroot.Trans("rebase.rebaseOnly"),
			false,
		))

	cmd.Args = cobra.ExactArgs(1)
	cmd.Example = "abroot rebase ghcr.io/vanilla-os/desktop:main"

	return cmd
}

func rebase(cmd *cobra.Command, args []string) error {

	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("rebase.rootRequired"))
		return nil
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	removePackages, err := cmd.Flags().GetBool("remove-packages")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	keepPackages, err := cmd.Flags().GetBool("keep-packages")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	if removePackages && keepPackages {
		flagError := errors.New(abroot.Trans("rebase.flagError"))
		cmdr.Error.Println(flagError)
		return err
	}

	name := args[0]
	abSys, err := core.NewABSystem()
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	pkgM, err := core.NewPackageManager(dryRun)
	if pkgM.Status == core.PKG_MNG_ENABLED {
		var removePackagesPrompt string
		if !keepPackages && !removePackages {
			cmdr.Info.Print(abroot.Trans("rebase.removePackagesLong"), " (y/N) ", " ")
			fmt.Scanln(&removePackagesPrompt)
		}
		if strings.Contains(removePackagesPrompt, "y") || removePackages {
			addedPackages, err := pkgM.GetAddPackages()
			if err != nil {
				return err
			}

			core.PrintVerboseInfo("cmd.Rebase", "Removing packages: ", addedPackages)

			if !dryRun {
				for _, v := range addedPackages {
					pkgM.Remove(v)
				}
			}
			cmdr.Info.Println(abroot.Trans("rebase.pkgRemoveSuccess"))
		}
	}

	err = abSys.Rebase(name, dryRun)
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	if dryRun {
		cmdr.Info.Println(abroot.Trans("rebase.dryRunSuccess"))
	}

	rebaseOnly, err := cmd.Flags().GetBool("rebase-only")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}
	if rebaseOnly {
		cmdr.Info.Println(abroot.Trans("rebase.success"))
		return nil
	}

	cmdr.Info.Println(abroot.Trans("rebase.successUpdate"))
	return abSys.RunOperation("UPGRADE", false)
}
