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
	"os"
	"os/exec"
	"runtime"
)

// Checks struct
type Checks struct{}

// NewChecks returns a new Checks struct
func NewChecks() *Checks {
	return &Checks{}
}

// PerformAllChecks performs all checks
func (c *Checks) PerformAllChecks() error {
	err := c.CheckCompatibilityFS()
	if err != nil {
		return err
	}

	err = c.CheckEssentialTools()
	if err != nil {
		return err
	}

	err = c.CheckConnectivity()
	if err != nil {
		return err
	}

	err = c.CheckRoot()
	if err != nil {
		return err
	}

	return nil
}

// CheckCompatibilityFS checks if the filesystem is compatible
func (c *Checks) CheckCompatibilityFS() error {
	PrintVerbose("CheckCompatibilityFS: running...")

	var fs []string
	if runtime.GOOS == "linux" {
		fs = []string{"ext4", "btrfs", "xfs"}
	} else {
		PrintVerbose("CheckCompatibilityFS:error: " + runtime.GOOS + " is not supported")
		return errors.New(`your OS ("` + runtime.GOOS + `") is not supported)`)
	}

	cmd, err := exec.Command("findmnt", "-n", "-o", "source", "/").Output()
	if err != nil {
		PrintVerbose("CheckCompatibilityFS:error(2): " + err.Error())
		return err
	}
	device := string([]byte(cmd[:len(cmd)-1]))

	cmd, err = exec.Command("lsblk", "-o", "fstype", "-n", device).Output()
	if err != nil {
		PrintVerbose("CheckCompatibilityFS:error(3): " + err.Error())
		return err
	}
	fsType := string([]byte(cmd[:len(cmd)-1]))

	for _, f := range fs {
		if f == string(fsType) {
			PrintVerbose("CheckCompatibilityFS: " + fsType + " is supported")
			return nil
		}
	}

	PrintVerbose("CheckCompatibilityFS:error(4): " + fsType + " is not supported")
	return errors.New(`the filesystem ("` + fsType + `") is not supported`)
}

// CheckEssentialTools checks if the essential tools are installed (podman, tar)
func (c *Checks) CheckEssentialTools() error {
	PrintVerbose("CheckEssentialTools: running...")

	var tools []string
	if runtime.GOOS == "linux" {
		tools = []string{"podman", "tar", "ping"}
	} else {
		PrintVerbose("CheckEssentialTools:error: " + runtime.GOOS + " is not supported")
		return errors.New(`your OS ("` + runtime.GOOS + `") is not supported)`)
	}

	for _, tool := range tools {
		_, err := exec.LookPath(tool)
		if err != nil {
			PrintVerbose("CheckEssentialTools:error(2): " + err.Error())
			return err
		}
	}

	PrintVerbose("CheckEssentialTools: all tools are installed")
	return nil
}

// CheckConnectivity checks if the system is connected to the internet
func (c *Checks) CheckConnectivity() error {
	PrintVerbose("CheckConnectivity: running...")

	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command("ping", "-c", "1", "google.com")
	} else {
		PrintVerbose("CheckConnectivity:error: " + runtime.GOOS + " is not supported")
		return errors.New(`your OS ("` + runtime.GOOS + `") is not supported)`)
	}

	err := cmd.Run()
	if err != nil {
		PrintVerbose("CheckConnectivity:error(2): " + err.Error())
		return err
	}

	PrintVerbose("CheckConnectivity: connected to the internet")
	return nil
}

// CheckRoot checks if the user is root
func (c *Checks) CheckRoot() error {
	PrintVerbose("CheckRoot: running...")

	if os.Geteuid() == 0 {
		PrintVerbose("CheckRoot: you are root")
		return nil
	}

	PrintVerbose("CheckRoot:error: you must be root")
	return errors.New("you must be root")
}
