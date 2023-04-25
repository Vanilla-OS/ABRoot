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

// Represents a Checks struct which contains all the checks which can
// be performed one by one or all at once using PerformAllChecks()
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
	PrintVerbose("Checks.CheckCompatibilityFS: running...")

	var fs []string
	if runtime.GOOS == "linux" {
		fs = []string{"ext4", "btrfs", "xfs"}
	} else {
		PrintVerbose("Checks.CheckCompatibilityFS:err: " + runtime.GOOS + " is not supported")
		return errors.New(`your OS ("` + runtime.GOOS + `") is not supported)`)
	}

	cmd, err := exec.Command("findmnt", "-n", "-o", "source", "/").Output()
	if err != nil {
		PrintVerbose("Checks.CheckCompatibilityFS:err(2): " + err.Error())
		return err
	}
	device := string([]byte(cmd[:len(cmd)-1]))

	cmd, err = exec.Command("lsblk", "-o", "fstype", "-n", device).Output()
	if err != nil {
		PrintVerbose("Checks.CheckCompatibilityFS:err(3): " + err.Error())
		return err
	}
	fsType := string([]byte(cmd[:len(cmd)-1]))

	for _, f := range fs {
		if f == string(fsType) {
			PrintVerbose("CheckCompatibilityFS: " + fsType + " is supported")
			return nil
		}
	}

	err = errors.New(`the filesystem ("` + fsType + `") is not supported`)
	PrintVerbose("Checks.CheckCompatibilityFS:err(4): " + err.Error())
	return err
}

// CheckEssentialTools checks if the essential tools are installed (podman, tar)
func (c *Checks) CheckEssentialTools() error {
	PrintVerbose("Checks.CheckEssentialTools: running...")

	var tools []string
	if runtime.GOOS == "linux" {
		tools = []string{"podman", "tar", "ping"}
	} else {
		err := errors.New(`your OS ("` + runtime.GOOS + `") is not supported)`)
		PrintVerbose("Checks.CheckEssentialTools:err(1): " + err.Error())
		return err
	}

	for _, tool := range tools {
		_, err := exec.LookPath(tool)
		if err != nil {
			PrintVerbose("Checks.CheckEssentialTools:err(2): " + err.Error())
			return err
		}
	}

	PrintVerbose("Checks.CheckEssentialTools: all essential tools are installed")
	return nil
}

// CheckConnectivity checks if the system is connected to the internet
func (c *Checks) CheckConnectivity() error {
	PrintVerbose("Checks.CheckConnectivity: running...")

	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command("ping", "-c", "1", "google.com")
	} else {
		err := errors.New(`your OS ("` + runtime.GOOS + `") is not supported)`)
		PrintVerbose("Checks.CheckConnectivity:err(1): " + err.Error())
		return err
	}

	err := cmd.Run()
	if err != nil {
		PrintVerbose("Checks.CheckConnectivity:err(2): " + err.Error())
		return err
	}

	err = errors.New("no internet connection")
	PrintVerbose("Checks.CheckConnectivity:err(3): " + err.Error())
	return err
}

// CheckRoot checks if the user is root
func (c *Checks) CheckRoot() error {
	PrintVerbose("Checks.CheckRoot: running...")

	if os.Geteuid() == 0 {
		PrintVerbose("Checks.CheckRoot: you are root")
		return nil
	}

	err := errors.New("not root")
	PrintVerbose("Checks.CheckRoot:err(1): " + err.Error())
	return err
}
