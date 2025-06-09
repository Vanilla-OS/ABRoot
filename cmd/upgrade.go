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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/differ/diff"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewUpgradeCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"upgrade",
		abroot.Trans("upgrade.long"),
		abroot.Trans("upgrade.short"),
		upgrade,
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"check-only",
			"c",
			abroot.Trans("upgrade.checkOnlyFlag"),
			false))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"dry-run",
			"d",
			abroot.Trans("upgrade.dryRunFlag"),
			false))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"force",
			"f",
			abroot.Trans("upgrade.forceFlag"),
			false))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"delete-old-system",
			"",
			abroot.Trans("upgrade.deleteOld"),
			false))

	cmd.Example = "abroot upgrade"

	return cmd
}

func upgrade(cmd *cobra.Command, args []string) error {
	checkOnly, err := cmd.Flags().GetBool("check-only")
	if err != nil {
		cmdr.Error.Println(err)
		return err
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

	aBsys, err := core.NewABSystem()
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	if checkOnly {
		_, raw := os.LookupEnv("ABROOT_JSON_OUTPUT")
		if !raw {
			cmdr.Info.Println(abroot.Trans("upgrade.checkingSystemUpdate"))
		}

		// Check for image updates
		newDigest, res, err := aBsys.CheckUpdate()
		if err != nil {
			cmdr.Error.Println(err)
			return err
		}
		sysAdded, sysUpgraded, sysDowngraded, sysRemoved := []diff.PackageDiff{}, []diff.PackageDiff{}, []diff.PackageDiff{}, []diff.PackageDiff{}
		if res {
			if !raw {
				cmdr.Info.Println(abroot.Trans("upgrade.systemUpdateAvailable"))
			}

			sysAdded, sysUpgraded, sysDowngraded, sysRemoved, err = core.BaseImagePackageDiff(aBsys.CurImage.Digest, newDigest)
			if err != nil {
				return err
			}
			if !raw {
				err = renderPackageDiff(sysAdded, sysUpgraded, sysDowngraded, sysRemoved)
				if err != nil {
					return err
				}
			}
		} else if !raw {
			cmdr.Info.Println(abroot.Trans("upgrade.noUpdateAvailable"))
		}

		// Check for package updates
		if !raw {
			cmdr.Info.Println(abroot.Trans("upgrade.checkingPackageUpdate"))
		}
		ovlAdded, ovlUpgraded, ovlDowngraded, ovlRemoved, err := core.OverlayPackageDiff()
		if err != nil {
			return err
		}

		sumChanges := len(ovlAdded) + len(ovlUpgraded) + len(ovlDowngraded) + len(ovlRemoved)
		if sumChanges == 0 && !raw {
			cmdr.Info.Println(abroot.Trans("upgrade.noUpdateAvailable"))
		} else if !raw {
			cmdr.Info.Sprintf(abroot.Trans("upgrade.packageUpdateAvailable"), sumChanges)

			err = renderPackageDiff(ovlAdded, ovlUpgraded, ovlDowngraded, ovlRemoved)
			if err != nil {
				return err
			}
		}

		if raw {
			newDigestIfHasUpdate := ""
			if res {
				newDigestIfHasUpdate = newDigest
			}

			out, err := json.Marshal(map[string]any{
				"hasUpdate": res,
				"newDigest": newDigestIfHasUpdate,
				"systemPackageDiff": map[string][]diff.PackageDiff{
					"added":      sysAdded,
					"upgraded":   sysUpgraded,
					"downgraded": sysDowngraded,
					"removed":    sysRemoved,
				},
				"overlayPackageDiff": map[string][]diff.PackageDiff{
					"added":      ovlAdded,
					"upgraded":   ovlUpgraded,
					"downgraded": ovlDowngraded,
					"removed":    ovlRemoved,
				},
			})
			if err != nil {
				cmdr.Error.Println(err)
			}

			fmt.Println(string(out))
		}

		if !res && sumChanges == 0 {
			os.Exit(1) // No update available
		} else {
			os.Exit(0) // Update available
		}
	}

	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("upgrade.rootRequired"))
		return nil
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		cmdr.Error.Println(err)
		return err
	}

	var operation core.ABSystemOperation
	if force {
		operation = core.FORCE_UPGRADE
	} else if dryRun {
		operation = core.DRY_RUN_UPGRADE
	} else {
		operation = core.UPGRADE
	}

	cmdr.Info.Println(abroot.Trans("upgrade.checkingSystemUpdate"))
	err = aBsys.RunOperation(operation, freeSpace)
	if err != nil {
		if err == core.ErrNoUpdate {
			cmdr.Info.Println(abroot.Trans("upgrade.noUpdateAvailable"))
			return err
		}

		cmdr.Error.Println(err)
		return err
	}

	if dryRun {
		cmdr.Info.Println(abroot.Trans("upgrade.dryRunSuccess"))
	}

	cmdr.Info.Println(abroot.Trans("upgrade.success"))
	os.Exit(0)
	return nil
}

func renderPackageDiff(added, upgraded, downgraded, removed []diff.PackageDiff) error {
	pkgFmt := "%s  '%s' -> '%s'"

	// Calculate largest string for proper alignment
	largestPkgName := 0
	for _, pkgSet := range [][]diff.PackageDiff{added, upgraded, downgraded, removed} {
		for _, pkg := range pkgSet {
			if len(pkg.Name) > largestPkgName {
				largestPkgName = len(pkg.Name)
			}
		}
	}

	for _, pkgSet := range []struct {
		Set    []diff.PackageDiff
		Header string
		Color  cmdr.Color
	}{
		{added, abroot.Trans("upgrade.added"), cmdr.FgGreen},
		{upgraded, abroot.Trans("upgrade.upgraded"), cmdr.FgBlue},
		{downgraded, abroot.Trans("upgrade.downgraded"), cmdr.FgYellow},
		{removed, abroot.Trans("upgrade.removed"), cmdr.FgRed},
	} {
		cmdr.NewStyle(cmdr.Bold, pkgSet.Color).Println(pkgSet.Header + ":")
		bulletItems := []cmdr.BulletListItem{}
		for _, pkg := range pkgSet.Set {
			bulletItems = append(bulletItems, cmdr.BulletListItem{
				Level: 1,
				Text:  fmt.Sprintf(pkgFmt, pkg.Name+strings.Repeat(" ", largestPkgName-len(pkg.Name)), pkg.PreviousVersion, pkg.NewVersion),
			})
		}
		err := cmdr.BulletList.WithItems(bulletItems).Render()
		if err != nil {
			return err
		}
	}

	return nil
}
