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
	"os"
	"os/exec"
	"path/filepath"
)

// links that must exist in the root partition
var linksToRepair = [...][2]string{
	{".system/bin", "bin"},
	{".system/etc", "etc"},
	{".system/lib", "lib"},
	{".system/lib32", "lib32"},
	{".system/lib64", "lib64"},
	{".system/libx32", "libx32"},
	{".system/sbin", "sbin"},
	{".system/usr", "usr"},
}

// paths that must exist in the root partition
var pathsToRepair = [...]string{
	".system",
	"boot",
	"dev",
	"opt",
	"part-future",
	"proc",
	"run",
	"srv",
	"sys",
	"tmp",
	"home",
	"media",
	"mnt",
	"root",
	"var",
	"var/home",
	"var/media",
	"var/mnt",
	"var/opt",
	"var/root",
}

func RepairRootIntegrity(rootPath string) (err error) {
	fixupOlderSystems(rootPath)

	err = repairLinks(rootPath)
	if err != nil {
		return err
	}

	err = repairPaths(rootPath)
	if err != nil {
		return err
	}

	return nil
}

func repairLinks(rootPath string) (err error) {
	for _, link := range linksToRepair {
		sourceAbs := filepath.Join(rootPath, link[0])
		targetAbs := filepath.Join(rootPath, link[1])

		err = repairLink(sourceAbs, targetAbs)
		if err != nil {
			return err
		}
	}
	return nil
}

func repairLink(sourceAbs, targetAbs string) (err error) {
	target := targetAbs
	source, err := filepath.Rel(filepath.Dir(target), sourceAbs)
	if err != nil {
		PrintVerboseErr("repairLink", 1, "Can't make ", source, " relative to ", target, " : ", err)
		return err
	}

	if !isLink(target) {
		err = os.RemoveAll(target)
		if err != nil && !os.IsNotExist(err) {
			PrintVerboseErr("repairLink", 2, "Can't remove ", target, " : ", err)
			return err
		}

		PrintVerboseInfo("repairLink", "Repairing ", target, " -> ", source)
		err = os.Symlink(source, target)
		if err != nil {
			return err
		}
	}

	return nil
}

func repairPaths(rootPath string) (err error) {
	for _, path := range pathsToRepair {
		err = repairPath(filepath.Join(rootPath, path))
		if err != nil {
			return err
		}
	}
	return nil
}

func repairPath(path string) (err error) {
	if info, err := os.Lstat(path); err == nil && info.IsDir() {
		return nil
	}

	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		PrintVerboseErr("repairPath", 1, "Can't remove ", path, " : ", err)
		return err
	}

	PrintVerboseInfo("repairPath", "Repairing ", path)
	err = os.MkdirAll(path, 0o755)
	if err != nil {
		PrintVerboseErr("repairPath", 2, "Can't create ", path, " : ", err)
		return err
	}

	return nil
}

// this is here to keep compatibility with older systems
// e.g. /media was a folder instead of a mountpoint for /var/media
func fixupOlderSystems(rootPath string) {
	paths := []string{
		"media",
		"mnt",
		"root",
	}

	for _, path := range paths {
		legacyPath := filepath.Join(rootPath, path)
		newPath := filepath.Join("/var", path)

		if _, err := os.Lstat(newPath); errors.Is(err, os.ErrNotExist) {
			err = exec.Command("mv", legacyPath, newPath).Run()
			if err != nil {
				PrintVerboseErr("fixupOlderSystems", 1, "could not move ", legacyPath, " to ", newPath, " : ", err)
				// if moving failed it probably means that it migrated successfully in the past
				// so it's safe to ignore errors
			}
		}
	}
}
