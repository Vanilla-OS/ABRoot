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
	"os"
	"path/filepath"

	"github.com/vanilla-os/abroot/settings"
)

type IntegrityCheck struct {
	rootPath      string
	systemPath    string
	etcA          string
	etcB          string
	standardLinks []string
	rootPaths     []string
}

// NewIntegrityCheck creates a new IntegrityCheck instance
func NewIntegrityCheck(root ABRootPartition, repair bool) (*IntegrityCheck, error) {
	systemPath := filepath.Join(root.MountPoint, "/.system")
	ic := &IntegrityCheck{
		rootPath:   root.MountPoint,
		systemPath: systemPath,
		etcA:       filepath.Join("/var/lib/abroot/etc/", settings.Cnf.PartLabelA),
		etcB:       filepath.Join("/var/lib/abroot/etc/", settings.Cnf.PartLabelB),
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
			"/etc",
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
	}

	if err := ic.check(repair); err != nil {
		return nil, err
	}

	return ic, nil
}

// check performs an integrity check on the system
func (ic *IntegrityCheck) check(repair bool) error {
	repairPaths := []string{}
	repairLinks := []string{}

	// check if system dir exists
	if !fileExists(ic.systemPath) {
		repairPaths = append(repairPaths, ic.systemPath)
	}

	// check if etc dirs exist
	if !fileExists(ic.etcA) {
		repairPaths = append(repairPaths, ic.etcA)
	}
	if !fileExists(ic.etcB) {
		repairPaths = append(repairPaths, ic.etcB)
	}

	// check if standard links exist and are links
	for _, link := range ic.standardLinks {
		if !isLink(link) {
			repairLinks = append(repairLinks, filepath.Join(ic.rootPath, link))
		}
	}

	// check if root paths exist
	for _, path := range ic.rootPaths {
		if !fileExists(path) {
			repairPaths = append(repairPaths, filepath.Join(ic.rootPath, path))
		}
	}

	if repair {
		for _, path := range repairPaths {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		}

		for _, link := range repairLinks {
			sysPath := filepath.Join(ic.systemPath, link)
			if err := os.Symlink(sysPath, link); err != nil {
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
