package core

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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vanilla-os/abroot/settings"
)

// Grub represents a grub instance, it exposes methods to generate a new grub
// config compatible with ABRoot, and to check if the system is booted into
// the present root or the future root
type Grub struct {
	PresentRoot string
	FutureRoot  string
}

// generateABGrubConf generates a new grub config with the given details
// kernel version is automatically detected
func generateABGrubConf(rootPath string, rootUuid string, rootLabel string) error {
	PrintVerboseInfo("generateABGrubConf", "generating grub config for ABRoot")

	kargs, err := KargsRead()
	if err != nil {
		PrintVerboseErr("generateABGrubConf", 0, err)
		return err
	}

	var grubPath, bootPrefix, bootPath, systemRoot string
	if settings.Cnf.ThinProvisioning {
		grubPath = filepath.Join(rootPath, "boot", "init", rootLabel)
		bootPrefix = "/" + rootLabel
		bootPath = grubPath

		diskM := NewDiskManager()
		sysRootPart, err := diskM.GetPartitionByLabel(rootLabel)
		if err != nil {
			PrintVerboseErr("generateABGrubConf", 3, err)
			return err
		}
		systemRoot = "/dev/mapper/" + sysRootPart.Device
	} else {
		grubPath = filepath.Join(rootPath, "boot", "grub")
		bootPrefix = "/.system/boot"
		bootPath = filepath.Join(rootPath, "boot")
		systemRoot = "UUID=" + rootUuid
	}

	confPath := filepath.Join(grubPath, "abroot.cfg")
	template := `insmod gzio
insmod part_gpt
insmod ext2
search --no-floppy --fs-uuid --set=root %s
linux   %s/vmlinuz-%s root=%s %s
initrd  %s/initrd.img-%s`

	kernelVersion := getKernelVersion(bootPath)
	if kernelVersion == "" {
		err := errors.New("could not get kernel version")
		PrintVerboseErr("generateABGrubConf", 1, err)
		return err
	}

	err = os.MkdirAll(grubPath, 0755)
	if err != nil {
		PrintVerboseErr("generateABGrubConf", 2, err)
		return err
	}

	err = os.WriteFile(
		confPath,
		[]byte(fmt.Sprintf(template, rootUuid, bootPrefix, kernelVersion, systemRoot, kargs, bootPrefix, kernelVersion)),
		0644,
	)
	if err != nil {
		PrintVerboseErr("generateABGrubConf", 3, err)
		return err
	}

	PrintVerboseInfo("generateABGrubConf", "done")
	return nil
}

// getKernelVersion returns the latest kernel version available in the root
func getKernelVersion(bootPath string) string {
	PrintVerboseInfo("getKernelVersion", "running...")

	kernelDir := filepath.Join(bootPath, "vmlinuz-*")
	files, err := filepath.Glob(kernelDir)
	if err != nil {
		PrintVerboseErr("getKernelVersion", 0, err)
		return ""
	}

	if len(files) == 0 {
		PrintVerboseErr("getKernelVersion", 1, errors.New("no kernel found"))
		return ""
	}

	var maxVersion string
	for _, file := range files {
		version := filepath.Base(file)
		if version > maxVersion {
			maxVersion = version
		}
	}

	maxVersion = maxVersion[8:]

	PrintVerboseInfo("getKernelVersion", "done")
	return maxVersion
}

// NewGrub creates a new Grub instance
func NewGrub(bootPart Partition) (*Grub, error) {
	PrintVerboseInfo("NewGrub", "running...")

	grubPath := filepath.Join(bootPart.MountPoint, "grub")
	confPath := filepath.Join(grubPath, "grub.cfg")

	cfg, err := os.ReadFile(confPath)
	if err != nil {
		PrintVerboseErr("NewGrub", 0, err)
		return nil, err
	}

	var presentRoot, futureRoot string

	for _, entry := range strings.Split(string(cfg), "\n") {
		if strings.Contains(entry, "abroot-a") {
			if strings.Contains(entry, "Current State") {
				presentRoot = "a"
			} else if strings.Contains(entry, "Previous State") {
				futureRoot = "a"
			}
		} else if strings.Contains(entry, "abroot-b") {
			if strings.Contains(entry, "Current State") {
				presentRoot = "b"
			} else if strings.Contains(entry, "Previous State") {
				futureRoot = "b"
			}
		}
	}

	if presentRoot == "" || futureRoot == "" {
		err := errors.New("could not find root partitions")
		PrintVerboseErr("NewGrub", 1, err)
		return nil, err
	}

	PrintVerboseInfo("NewGrub", "done")
	return &Grub{
		PresentRoot: presentRoot,
		FutureRoot:  futureRoot,
	}, nil
}

func (g *Grub) IsBootedIntoPresentRoot() (bool, error) {
	PrintVerboseInfo("Grub.IsBootedIntoPresentRoot", "running...")

	a := NewABRootManager()
	future, err := a.GetFuture()
	if err != nil {
		return false, err
	}

	if g.FutureRoot == "a" {
		PrintVerboseInfo("Grub.IsBootedIntoPresentRoot", "done")
		return future.Label == settings.Cnf.PartLabelA, nil
	} else {
		PrintVerboseInfo("Grub.IsBootedIntoPresentRoot", "done")
		return future.Label == settings.Cnf.PartLabelB, nil
	}
}
