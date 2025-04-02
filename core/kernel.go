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
	"path/filepath"

	"github.com/hashicorp/go-version"
)

// getKernelName returns the name of the latest kernel found in the boot directory
func getKernelName(bootPath string) string {
	PrintVerboseInfo("getKernelName", "running...")

	kernelDir := filepath.Join(bootPath, "vmlinuz-*")
	files, err := filepath.Glob(kernelDir)
	if err != nil {
		PrintVerboseErr("getKernelName", 0, err)
		return ""
	}

	if len(files) == 0 {
		PrintVerboseErr("getKernelName", 1, errors.New("no kernel found"))
		return ""
	}

	var maxVer *version.Version
	var latestKernel string

	for _, file := range files {
		verStr := filepath.Base(file)[8:]
		ver, err := version.NewVersion(verStr)
		if err == nil {
			if maxVer == nil || ver.GreaterThan(maxVer) {
				maxVer = ver
				latestKernel = verStr
			}
		} else {
			latestKernel = verStr
		}
	}

	if maxVer != nil {
		return maxVer.String()
	}

	return latestKernel
}
