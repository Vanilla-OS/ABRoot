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
	"io/ioutil"
	"os"
	"path/filepath"
)

// generateABGrubConf generates a new grub config with the given details
// kernel version is automatically detected
func generateABGrubConf(rootPath string, rootUuid string) error {
	PrintVerbose("generateABGrubConf: generating grub config for ABRoot")

	grubPath := filepath.Join(rootPath, "boot", "grub")
	confPath := filepath.Join(grubPath, "abroot.cfg")
	template := `insmod gzio
insmod part_gpt
insmod ext2
search --no-floppy --fs-uuid --set=root %s
linux   /.system/boot/vmlinuz-%s root=UUID=%s quiet splash bgrt_disable $vt_handoff
initrd  /.system/boot/initrd.img-%s`

	kernelVersion := getKernelVersion(rootPath)
	if kernelVersion == "" {
		err := errors.New("could not get kernel version")
		PrintVerbose("generateABGrubConf:err: %s", err)
		return err
	}

	err := os.MkdirAll(grubPath, 0755)
	if err != nil {
		PrintVerbose("generateABGrubConf:err(2): %s", err)
		return err
	}

	err = ioutil.WriteFile(
		confPath,
		[]byte(fmt.Sprintf(template, rootUuid, kernelVersion, rootUuid, kernelVersion)),
		0644,
	)
	if err != nil {
		PrintVerbose("generateABGrubConf:err(3): %s", err)
		return err
	}

	return nil
}

// getKernelVersion returns the latest kernel version available in the root
func getKernelVersion(rootPath string) string {
	PrintVerbose("getKernelVersion: getting kernel version")

	kernelDir := filepath.Join(rootPath, "boot", "vmlinuz-*")
	files, err := filepath.Glob(kernelDir)
	if err != nil {
		PrintVerbose("getKernelVersion:err: %s", err)
		return ""
	}

	if len(files) == 0 {
		PrintVerbose("getKernelVersion:err: no kernel found")
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

	return maxVersion
}
