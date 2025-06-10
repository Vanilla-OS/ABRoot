package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

var validPkgArgs = []string{"add", "remove", "list", "apply"}

func NewPkgCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"pkg add|remove|list|apply",
		abroot.Trans("pkg.long"),
		abroot.Trans("pkg.short"),
		pkg,
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"dry-run",
			"d",
			abroot.Trans("pkg.dryRunFlag"),
			false))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"force-enable-user-agreement",
			"f",
			abroot.Trans("pkg.forceEnableUserAgreementFlag"),
			false))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"force-apply",
			"",
			abroot.Trans("pkg.forceApply"),
			false))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"delete-old-system",
			"",
			abroot.Trans("upgrade.deleteOld"),
			false))

	cmd.Args = cobra.MinimumNArgs(1)
	cmd.ValidArgs = validPkgArgs
	cmd.Example = "abroot pkg add <pkg>"

	return cmd
}

func pkg(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("pkg.rootRequired"))
		return nil
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	freeSpace, err := cmd.Flags().GetBool("delete-old-system")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	forceEnableUserAgreement, err := cmd.Flags().GetBool("force-enable-user-agreement")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	forceApply, err := cmd.Flags().GetBool("force-apply")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	pkgM, err := core.NewPackageManager(false)
	if err != nil {
		cmdr.Error.Println(abroot.Trans("pkg.failedGettingPkgManagerInstance", err))
		return err
	}

	// Check for user agreement, here we could simply call the CheckStatus
	// function which also checks if the package manager is enabled or not
	// since this pkg command is not even added to the root command if the
	// package manager is disabled, but we want to be explicit here to avoid
	// potential hard to debug errors in the future in weird development
	// scenarios. Yeah, trust me, I've been there.
	if pkgM.Status == core.PKG_MNG_REQ_AGREEMENT {
		err = pkgM.CheckStatus()
		if err != nil {
			if !forceEnableUserAgreement {
				cmdr.Info.Println(abroot.Trans("pkg.agreementMsg"))
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(answer)
				if answer == "y" || answer == "Y" {
					err := pkgM.AcceptUserAgreement()
					if err != nil {
						cmdr.Error.Println(abroot.Trans("pkg.agreementSignFailed"), err)
						return err
					}
				} else {
					cmdr.Info.Println(abroot.Trans("pkg.agreementDeclined"))
					return nil
				}
			} else {
				err := pkgM.AcceptUserAgreement()
				if err != nil {
					cmdr.Error.Println(abroot.Trans("pkg.agreementSignFailed"), err)
					return err
				}
			}
		}
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			return errors.New(abroot.Trans("pkg.noPackageNameProvided"))
		}
		for _, pkg := range args[1:] {
			err := pkgM.Add(pkg)
			if err != nil {
				cmdr.Error.Println(err)
				return err
			}
		}
		cmdr.Info.Printf(abroot.Trans("pkg.addedMsg"), strings.Join(args[1:], ", "))
	case "remove":
		if len(args) < 2 {
			return errors.New(abroot.Trans("pkg.noPackageNameProvided"))
		}
		for _, pkg := range args[1:] {
			err := pkgM.Remove(pkg)
			if err != nil {
				cmdr.Error.Println(err)
				return err
			}
		}
		cmdr.Info.Printf(abroot.Trans("pkg.removedMsg"), strings.Join(args[1:], ", "))
	case "list":
		added, err := pkgM.GetAddPackagesString("\n")
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}

		removed, err := pkgM.GetRemovePackagesString("\n")
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}

		cmdr.Info.Printf(abroot.Trans("pkg.listMsg"), added, removed)
		return nil
	case "apply":
		unstagedAdded, unstagedRemoved, err := pkgM.GetUnstagedPackages("/")
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}

		if !forceApply && len(unstagedAdded) == 0 && len(unstagedRemoved) == 0 {
			cmdr.Info.Println(abroot.Trans("pkg.noChanges"))
			return nil
		}

		aBsys, err := core.NewABSystem()
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}

		if dryRun {
			err = aBsys.RunOperation(core.DRY_RUN_APPLY, freeSpace)
		} else {
			err = aBsys.RunOperation(core.APPLY, freeSpace)
		}
		if err != nil {
			cmdr.Error.Printf(abroot.Trans("pkg.applyFailed"), err)
			return err
		}
		cmdr.Info.Println(abroot.Trans("pkg.applySuccess"))
	default:
		cmdr.Error.Println(abroot.Trans("pkg.unknownCommand", args[0]))
		return nil
	}

	return nil
}
