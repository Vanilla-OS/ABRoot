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
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"
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

// CheckCompatibilityFS checks if the filesystem is compatible with ABRoot v2
// if not, it returns an error. Note that currently only ext4, btrfs and xfs
// are supported/tested. Here we assume some utilities are installed, such as
// findmnt and lsblk
func (c *Checks) CheckCompatibilityFS() error {
	PrintVerboseInfo("Checks.CheckCompatibilityFS", "running...")

	var fs []string
	if runtime.GOOS == "linux" {
		fs = []string{"ext4", "btrfs", "xfs"}
	} else {
		err := fmt.Errorf("your OS (%s) is not supported", runtime.GOOS)
		PrintVerboseErr("Checks.CheckCompatibilityFS", 0, err)
		return err
	}

	cmd, err := exec.Command("findmnt", "-n", "-o", "source", "/").Output()
	if err != nil {
		PrintVerboseErr("Checks.CheckCompatibilityFS", 1, err)
		return err
	}
	device := string([]byte(cmd[:len(cmd)-1]))

	cmd, err = exec.Command("lsblk", "-o", "fstype", "-n", device).Output()
	if err != nil {
		PrintVerboseErr("Checks.CheckCompatibilityFS", 2, err)
		return err
	}
	fsType := string([]byte(cmd[:len(cmd)-1]))

	for _, f := range fs {
		if f == string(fsType) {
			PrintVerboseInfo("CheckCompatibilityFS", fsType, "is supported")
			return nil
		}
	}

	err = fmt.Errorf("the filesystem (%s) is not supported", fsType)
	PrintVerboseErr("Checks.CheckCompatibilityFS", 3, err)
	return err
}

// CheckConnectivity checks if the system is connected to the internet
func (c *Checks) CheckConnectivity() error {
	PrintVerboseInfo("Checks.CheckConnectivity", "running...")

	timeout := 5 * time.Second
	_, err := net.DialTimeout("tcp", "vanillaos.org:80", timeout)
	if err != nil {
		PrintVerboseErr("Checks.CheckConnectivity", 1, err)
		return err
	}

	return nil
}

// CheckRoot checks if the user is root and returns an error if not
func (c *Checks) CheckRoot() error {
	PrintVerboseInfo("Checks.CheckRoot", "running...")

	if os.Geteuid() == 0 {
		PrintVerboseInfo("Checks.CheckRoot", "you are root")
		return nil
	}

	err := errors.New("not root")
	PrintVerboseErr("Checks.CheckRoot", 1, err)
	return err
}
