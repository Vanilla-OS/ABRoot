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
	"os/exec"
	"path/filepath"
	"strings"
)

// Chroot is a struct which represents a chroot environment
type Chroot struct {
	root       string
	rootUuid   string
	rootDevice string
}

var ReservedMounts = []string{
	"/dev",
	"/dev/pts",
	"/proc",
	"/run",
	"/sys",
}

// NewChroot creates a new chroot environment
func NewChroot(root string, rootUuid string, rootDevice string) (*Chroot, error) {
	PrintVerbose("NewChroot: running...")

	root = strings.ReplaceAll(root, "//", "/")

	if _, err := os.Stat(root); os.IsNotExist(err) {
		PrintVerbose("NewChroot:err: " + err.Error())
		return nil, err
	}

	chroot := &Chroot{
		root:       root,
		rootUuid:   rootUuid,
		rootDevice: rootDevice,
	}

	// workaround for grub-mkconfig, not able to find the device
	// inside a chroot environment
	err := chroot.Execute("mount", []string{"--bind", "/", "/"})
	if err != nil {
		PrintVerbose("NewChroot:err(2): " + err.Error())
		return nil, err
	}

	for _, mount := range ReservedMounts {
		err := exec.Command("mount", "--bind", mount, root+mount).Run()
		fmt.Println("mounting", mount, "to", root+mount)
		if err != nil {
			PrintVerbose("NewChroot:err(3): " + err.Error())
			return nil, err
		}
	}

	PrintVerbose("NewChroot: successfully created.")
	return chroot, nil
}

// Close unmounts all the bind mounts
func (c *Chroot) Close() error {
	PrintVerbose("Chroot.Close: running...")

	err := exec.Command("umount", filepath.Join(c.root, "/dev/pts")).Run()
	if err != nil {
		PrintVerbose("Chroot.Close:err: " + err.Error())
		return err
	}

	mountList := ReservedMounts
	mountList = append(mountList, "")

	for _, mount := range mountList {
		if mount == "/dev/pts" {
			continue
		}

		mountDir := filepath.Join(c.root, mount)
		PrintVerbose("Chroot.Close: unmounting " + mountDir)
		err := exec.Command("umount", mountDir).Run()
		if err != nil {
			PrintVerbose("Chroot.Close:err: " + err.Error())
			return err
		}
	}

	PrintVerbose("Chroot.Close: successfully closed.")
	return nil
}

// Execute runs a command in the chroot environment
func (c *Chroot) Execute(cmd string, args []string) error {
	PrintVerbose("Chroot.Execute: running...")

	cmd = strings.Join(append([]string{cmd}, args...), " ")
	PrintVerbose("Chroot.Execute: running command: " + cmd)
	e := exec.Command("chroot", c.root, "/bin/sh", "-c", cmd)
	e.Stdout = os.Stdout
	e.Stderr = os.Stderr
	e.Stdin = os.Stdin
	err := e.Run()
	if err != nil {
		PrintVerbose("Chroot.Execute:err: " + err.Error())
		return err
	}

	PrintVerbose("Chroot.Execute: successfully ran.")
	return nil
}

// ExecuteCmds runs a list of commands in the chroot environment,
// stops at the first error
func (c *Chroot) ExecuteCmds(cmds []string) error {
	PrintVerbose("Chroot.ExecuteCmds: running...")

	for _, cmd := range cmds {
		err := c.Execute(cmd, []string{})
		if err != nil {
			PrintVerbose("Chroot.ExecuteCmds:err: " + err.Error())
			return err
		}
	}

	PrintVerbose("Chroot.ExecuteCmds: successfully ran.")
	return nil
}
