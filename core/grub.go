package core

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
func generateABGrubConf(kernelName string, rootPath string, rootUuid string, rootLabel string, generatedGrubConfigPath string) error {
	PrintVerboseInfo("generateABGrubConf", "generating grub config for ABRoot")

	kargs, err := KargsRead()
	if err != nil {
		PrintVerboseErr("generateABGrubConf", 0, err)
		return err
	}

	var grubPath, bootPrefix, systemRoot string
	if settings.Cnf.ThinProvisioning {
		grubPath = filepath.Join(rootPath, "boot", "init", rootLabel)
		bootPrefix = "/" + rootLabel

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
		systemRoot = "UUID=" + rootUuid
	}

	confPath := filepath.Join(grubPath, "abroot.cfg")
	template := `  search --no-floppy --fs-uuid --set=root %s
  linux   %s/vmlinuz-%s root=%s %s
  initrd  %s/%s
`

	err = os.MkdirAll(grubPath, 0755)
	if err != nil {
		PrintVerboseErr("generateABGrubConf", 2, err)
		return err
	}

	initramfsName := fmt.Sprintf(settings.Cnf.InitramfsFormat, kernelName)

	abrootBootConfig := fmt.Sprintf(template, rootUuid, bootPrefix, kernelName, systemRoot, kargs, bootPrefix, initramfsName)

	generatedGrubConfigContents, err := os.ReadFile(filepath.Join(rootPath, generatedGrubConfigPath))
	if err != nil {
		PrintVerboseErr("generateABGrubConf", 3, "could not read grub config", err)
		return err
	}

	generatedGrubConfig := string(generatedGrubConfigContents)

	replacementString := "REPLACED_BY_ABROOT"
	if !strings.Contains(generatedGrubConfig, replacementString) {
		err := errors.New("could not find replacement string \"" + replacementString + "\", check /etc/grub.d configuration")
		PrintVerboseErr("generateABGrubConf", 3.1, err)
		return err
	}
	grubConfigWithBootEntry := strings.Replace(generatedGrubConfig, "REPLACED_BY_ABROOT", abrootBootConfig, 1)

	err = os.WriteFile(confPath, []byte(grubConfigWithBootEntry), 0644)
	if err != nil {
		PrintVerboseErr("generateABGrubConf", 4, "could not read grub config", err)
		return err
	}

	PrintVerboseInfo("generateABGrubConf", "done")
	return nil
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
