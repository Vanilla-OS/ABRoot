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

	var maxVer *version.Version
	for _, file := range files {
		verStr := filepath.Base(file)[8:]
		ver, err := version.NewVersion(verStr)
		if err != nil {
			continue
		}
		if maxVer == nil || ver.GreaterThan(maxVer) {
			maxVer = ver
		}
	}

	if maxVer != nil {
		return maxVer.String()
	}

	return ""
}
