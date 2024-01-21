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
	"fmt"
	"os"
	"path/filepath"

	"github.com/vanilla-os/abroot/settings"
)

type IntegrityCheck struct {
	rootPath      string
	systemPath    string
	standardLinks []string
	rootPaths     []string
	etcPaths      []string
}

// NewIntegrityCheck creates a new IntegrityCheck instance for the given root
// partition, and returns a pointer to it or an error if something went wrong
func NewIntegrityCheck(root ABRootPartition, repair bool) (*IntegrityCheck, error) {
	systemPath := filepath.Join(root.Partition.MountPoint, "/.system")
	etcPath := filepath.Join("/var/lib/abroot/etc", root.Label)
	etcWorkPath := filepath.Join(
		"/var/lib/abroot/etc",
		fmt.Sprintf("%s-work", root.Label),
	)
	etcLocalePath := filepath.Join("/var/lib/abroot/etc", root.Label, "locales")
	ic := &IntegrityCheck{
		rootPath:   root.Partition.MountPoint,
		systemPath: systemPath,
		standardLinks: []string{
			"/bin",
			"/etc",
			"/lib",
			"/lib32",
			"/lib64",
			"/libx32",
			"/sbin",
			"/usr",
		},
		rootPaths: []string{ // those paths must be present in the root partition
			"/boot",
			"/dev",
			"/home",
			"/media",
			"/mnt",
			"/opt",
			"/part-future",
			"/proc",
			"/root",
			"/run",
			"/srv",
			"/sys",
			"/tmp",
			"/var",
			settings.Cnf.LibPathStates,
		},
		etcPaths: []string{
			etcPath,
			etcWorkPath,
			etcLocalePath,
		},
	}

	if err := ic.check(repair); err != nil {
		return nil, err
	}

	return ic, nil
}

// check performs an integrity check on the system by checking if all the
// required paths and links are present. If repair is true, it will also
// try to repair the system by creating missing resources
func (ic *IntegrityCheck) check(repair bool) error {
	PrintVerboseInfo("IntegrityCheck.check", "Running...")
	repairPaths := []string{}
	repairLinks := []string{}

	// check if system dir exists
	if !fileExists(ic.systemPath) {
		repairPaths = append(repairPaths, ic.systemPath)
	}

	// check if standard links exist and are links
	for _, link := range ic.standardLinks {
		testPath := filepath.Join(ic.rootPath, link)
		if !isLink(testPath) {
			repairLinks = append(repairLinks, link)
		}
	}

	// check if root paths exist
	for _, path := range ic.rootPaths {
		finalPath := filepath.Join(ic.rootPath, path)
		if !fileExists(finalPath) {
			repairPaths = append(repairPaths, finalPath)
		}
	}

	// check if etc paths exist
	for _, path := range ic.etcPaths {
		if !fileExists(path) {
			repairPaths = append(repairPaths, path)
		}
	}

	if repair {
		for _, path := range repairPaths {
			PrintVerboseInfo("IntegrityCheck", "Repairing path", path)
			err := os.MkdirAll(path, 0755)
			if err != nil {
				PrintVerboseErr("IntegrityCheck", 0, err)
				return err
			}
		}

		for _, link := range repairLinks {
			srcPath := filepath.Join(ic.systemPath, link)
			dstPath := filepath.Join(ic.rootPath, link)
			relSrcPath, err := filepath.Rel(filepath.Dir(dstPath), srcPath)
			if err != nil {
				PrintVerboseErr("IntegrityCheck", 1, err)
				return err
			}

			PrintVerboseInfo("IntegrityCheck", "Repairing link", relSrcPath, "->", dstPath)
			err = os.Symlink(relSrcPath, dstPath)
			if err != nil {
				PrintVerboseErr("IntegrityCheck", 2, err)
				return err
			}
		}
	}

	return nil
}

// Repair repairs the system
func (ic *IntegrityCheck) Repair() error {
	return ic.check(true)
}
