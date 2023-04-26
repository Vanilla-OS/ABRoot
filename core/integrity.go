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
	"fmt"
	"os"
	"path/filepath"

	"github.com/vanilla-os/abroot/settings"
)

type IntegrityCheck struct {
	statesPath          string
	privateOverlaysPath string
	systemPath          string
	rootEtcPath         string
	standardLinks       []string
}

// NewIntegrityCheck creates a new IntegrityCheck instance
func NewIntegrityCheck(root ABRootPartition, repair bool) (*IntegrityCheck, error) {
	systemPath := filepath.Join(root.MountPoint, "/.system")
	rootEtcPath := filepath.Join(
		settings.Cnf.LibPathPrivateOverlays,
		fmt.Sprintf("etc-%s", root.Label),
	)
	ic := &IntegrityCheck{
		statesPath:          settings.Cnf.LibPathStates,
		privateOverlaysPath: settings.Cnf.LibPathPrivateOverlays,
		systemPath:          systemPath,
		rootEtcPath:         rootEtcPath,
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

	// check if states dir exists
	if !fileExists(ic.statesPath) {
		repairPaths = append(repairPaths, ic.statesPath)
	}

	// check if private overlays dir exists
	if !fileExists(ic.privateOverlaysPath) {
		repairPaths = append(repairPaths, ic.privateOverlaysPath)
	}

	// check if root etc dir exists
	if !fileExists(ic.rootEtcPath) {
		repairPaths = append(repairPaths, ic.rootEtcPath)
	}

	// check if system dir exists
	if !fileExists(ic.systemPath) {
		repairPaths = append(repairPaths, ic.systemPath)
	}

	// check if standard links exist and are links
	for _, link := range ic.standardLinks {
		if !isLink(link) {
			repairLinks = append(repairLinks, link)
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
