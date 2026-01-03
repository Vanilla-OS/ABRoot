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

// createABSpecificGrub creates a directory that contains the root specific root information
func createABSpecificGrub(kernelVersion string, rootUuid string, rootLabel string, generatedGrubConfigPath string, bootMountpoint string, filesDir string) error {
	PrintVerboseInfo("createABSpecificGrub", "creating root specific grub info")

	bootPrefix := filepath.Join("/abroot", rootLabel)
	configDir := filepath.Join(bootMountpoint, bootPrefix)

	err := MoveFile(
		filepath.Join(filesDir, "vmlinuz-"+kernelVersion),
		filepath.Join(configDir, "vmlinuz-"+kernelVersion),
	)
	if err != nil {
		PrintVerboseErr("createABSpecificGrub", 1, err)
		return err
	}
	err = MoveFile(
		filepath.Join(filesDir, "initrd.img-"+kernelVersion),
		filepath.Join(configDir, "initrd.img-"+kernelVersion),
	)
	if err != nil {
		PrintVerboseErr("createABSpecificGrub", 2, err)
		return err
	}
	err = MoveFile(
		filepath.Join(filesDir, "config-"+kernelVersion),
		filepath.Join(configDir, "config-"+kernelVersion),
	)
	if err != nil {
		PrintVerboseErr("createABSpecificGrub", 3, err)
		return err
	}
	err = MoveFile(
		filepath.Join(filesDir, "System.map-"+kernelVersion),
		filepath.Join(configDir, "System.map-"+kernelVersion),
	)
	if err != nil {
		PrintVerboseErr("createABSpecificGrub", 4, err)
		return err
	}

	kargs, err := KargsRead()
	if err != nil {
		PrintVerboseErr("createABSpecificGrub", 5, err)
		return err
	}

	var systemRoot string
	if settings.Cnf.ThinProvisioning {
		diskM := NewDiskManager()
		sysRootPart, err := diskM.GetPartitionByLabel(rootLabel)
		if err != nil {
			PrintVerboseErr("createABSpecificGrub", 6, err)
			return err
		}
		systemRoot = "/dev/mapper/" + sysRootPart.Device
	} else {
		systemRoot = "UUID=" + rootUuid
	}

	confPath := filepath.Join(configDir, "abroot.cfg")
	template := `
  linux   %s/vmlinuz-%s root=%s %s
  initrd  %s/initrd.img-%s
`

	_ = os.RemoveAll(confPath)
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		PrintVerboseErr("createABSpecificGrub", 7, err)
		return err
	}

	abrootBootConfig := fmt.Sprintf(template, bootPrefix, kernelVersion, systemRoot, kargs, bootPrefix, kernelVersion)

	generatedGrubConfigContents, err := os.ReadFile(generatedGrubConfigPath)
	if err != nil {
		PrintVerboseErr("createABSpecificGrub", 8, "could not read grub config", err)
		return err
	}

	generatedGrubConfig := string(generatedGrubConfigContents)

	replacementString := "REPLACED_BY_ABROOT"
	if !strings.Contains(generatedGrubConfig, replacementString) {
		err := errors.New("could not find replacement string \"" + replacementString + "\", check /etc/grub.d configuration")
		PrintVerboseErr("createABSpecificGrub", 9, err)
		return err
	}
	grubConfigWithBootEntry := strings.Replace(generatedGrubConfig, "REPLACED_BY_ABROOT", abrootBootConfig, 1)

	err = os.WriteFile(confPath, []byte(grubConfigWithBootEntry), 0644)
	if err != nil {
		PrintVerboseErr("createABSpecificGrub", 10, "could not read grub config", err)
		return err
	}

	PrintVerboseInfo("createABSpecificGrub", "done")
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
