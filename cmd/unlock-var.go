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
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/orchid/cmdr"
)

type VarConfigError struct{}

func (e *VarConfigError) Error() string {
	return "reading the var disk from config is not implemented yet"
}

type VarInvalidError struct {
	passedDisk string
}

func (e *VarInvalidError) Error() string {
	return "the /var disk " + e.passedDisk + " does not exist"
}

type NotEncryptedError struct{}

func (e *NotEncryptedError) Error() string {
	return "the var partition is not encrypted"
}

func NewUnlockVarCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"unlock-var",
		"",
		"",
		unlockVarCmd,
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"dry-run",
			"d",
			"perform a dry run of the operation",
			false,
		),
	)

	cmd.WithStringFlag(
		cmdr.NewStringFlag(
			"var-disk",
			"m",
			"pass /var disk directly instead of reading from configuration",
			"",
		),
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"check-encrypted",
			"c",
			"check if drive is encrypted and return",
			false,
		),
	)

	cmd.Example = "abroot unlock-var"

	cmd.Hidden = true

	return cmd
}

// helper function which only returns syntax errors and prints other ones
func unlockVarCmd(cmd *cobra.Command, args []string) error {
	err := unlockVar(cmd, args)
	if err != nil {
		cmdr.Error.Println(err)
		os.Exit(1)
		return nil
	}
	return nil
}

func unlockVar(cmd *cobra.Command, _ []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println("You must be root to run this command.")
		os.Exit(2)
		return nil
	}

	varDisk, err := cmd.Flags().GetString("var-disk")
	if err != nil {
		return err
	}

	check_only, err := cmd.Flags().GetBool("check-encrypted")
	if err != nil {
		return err
	}

	_, err = os.Stat(filepath.Join("/dev/disk/by-label/", settings.Cnf.PartLabelVar))
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		return &NotEncryptedError{}
	}
	if check_only {
		cmdr.Info.Println("The var partition is encrypted.")
		return nil
	}

	if varDisk == "" {
		return &VarConfigError{}
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	partitions, err := core.NewDiskManager().GetPartitions("")
	if err != nil {
		return err
	}

	var varLuksPart core.Partition
	foundPart := false

	for _, partition := range partitions {
		devName := "/dev/"
		if partition.IsDevMapper() {
			devName += "mapper/"
		}
		devName += partition.Device

		if devName == varDisk {
			varLuksPart = partition
			foundPart = true
			break
		}
	}
	if !foundPart {
		return &VarInvalidError{varDisk}
	}

	uuid := varLuksPart.Uuid
	cmdr.FgDefault.Println("unlocking", varDisk)

	if dryRun {
		cmdr.Info.Println("Dry run complete.")
	} else {
		cryptsetupCmd := exec.Command("/usr/sbin/cryptsetup", "luksOpen", varDisk, "luks-"+uuid)
		cryptsetupCmd.Stdin = os.Stdin
		cryptsetupCmd.Stderr = os.Stderr
		cryptsetupCmd.Stdout = os.Stdout
		err := cryptsetupCmd.Run()
		if err != nil {
			return err
		}
		cmdr.Info.Println("The system mounts have been performed successfully.")
	}

	return nil
}
