package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

import (
	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
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

	cmd.Example = "abroot upgrade"

	return cmd
}

func upgrade(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("upgrade.rootRequired"))
		return nil
	}

	checkOnly, err := cmd.Flags().GetBool("check-only")
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
		if aBsys.CheckUpdate() {
			cmdr.Info.Println(abroot.Trans("upgrade.updateAvailable"))
		} else {
			cmdr.Info.Println(abroot.Trans("upgrade.noUpdateAvailable"))
		}
		return nil
	}

	// NOTE: This is just a test, to see if the code works
	// p := core.NewPodman()
	// cf := p.NewContainerFile(
	// 	"docker.io/library/alpine:latest",
	// 	map[string]string{
	// 		"LABEL": "test",
	// 	},
	// 	map[string]string{},
	// 	`RUN echo "test" > /test.txt`,
	// )
	// p.GenerateRootfs("testing", cf, "test")

	// diskM := core.NewDiskManager()
	// disk, err := diskM.GetDisk("nvme0n1")
	// if err != nil {
	// 	return err
	// }
	// for _, partition := range disk.Partitions {
	// 	fmt.Println(partition.Label)
	// }

	// a := core.NewABRootManager()
	// present, err := a.GetPresent()
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("present:", present.Label)

	// future, err := a.GetFuture()
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("future:", future.Label)

	// c := core.NewChecks()
	// err := c.PerformAllChecks()
	// if err != nil {
	// 	fmt.Println(err)
	// }

	return nil
}
