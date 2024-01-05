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
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Chroot represents a chroot instance, which can be used to run commands
// inside a chroot environment
type Chroot struct {
	root       string
	rootUuid   string
	rootDevice string
}

// ReservedMounts is a list of mount points from host which should be
// mounted inside the chroot environment to ensure it works properly in
// some cases, such as grub-mkconfig
var ReservedMounts = []string{
	"/dev",
	"/dev/pts",
	"/proc",
	"/run",
	"/sys",
}

// NewChroot creates a new chroot environment from the given root path and
// returns its Chroot instance or an error if something went wrong
func NewChroot(root string, rootUuid string, rootDevice string) (*Chroot, error) {
	PrintVerboseInfo("NewChroot", "running...")

	root = strings.ReplaceAll(root, "//", "/")

	if _, err := os.Stat(root); os.IsNotExist(err) {
		PrintVerboseErr("NewChroot", 0, err)
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
		PrintVerboseErr("NewChroot", 1, err)
		return nil, err
	}

	for _, mount := range ReservedMounts {
		PrintVerboseInfo("NewChroot", "mounting", mount)
		err := syscall.Mount(mount, filepath.Join(root, mount), "", syscall.MS_BIND, "")
		if err != nil {
			PrintVerboseErr("NewChroot", 2, err)
			return nil, err
		}
	}

	PrintVerboseInfo("NewChroot", "successfully created.")
	return chroot, nil
}

// Close unmounts all the bind mounts and closes the chroot environment
func (c *Chroot) Close() error {
	PrintVerboseInfo("Chroot.Close", "running...")

	err := syscall.Unmount(filepath.Join(c.root, "/dev/pts"), 0)
	if err != nil {
		PrintVerboseErr("Chroot.Close", 0, err)
		return err
	}

	mountList := ReservedMounts
	mountList = append(mountList, "")

	for _, mount := range mountList {
		if mount == "/dev/pts" {
			continue
		}

		mountDir := filepath.Join(c.root, mount)
		PrintVerboseInfo("Chroot.Close", "unmounting", mountDir)
		err := syscall.Unmount(mountDir, 0)
		if err != nil {
			PrintVerboseErr("Chroot.Close", 1, err)
			return err
		}
	}

	PrintVerboseInfo("Chroot.Close", "successfully closed.")
	return nil
}

// Execute runs a command in the chroot environment, the command is
// a string and the arguments are a list of strings. If an error occurs
// it is returned.
func (c *Chroot) Execute(cmd string, args []string) error {
	PrintVerboseInfo("Chroot.Execute", "running...")

	cmd = strings.Join(append([]string{cmd}, args...), " ")
	PrintVerboseInfo("Chroot.Execute", "running command:", cmd)
	e := exec.Command("chroot", c.root, "/bin/sh", "-c", cmd)
	e.Stdout = os.Stdout
	e.Stderr = os.Stderr
	e.Stdin = os.Stdin
	err := e.Run()
	if err != nil {
		PrintVerboseErr("Chroot.Execute", 0, err)
		return err
	}

	PrintVerboseInfo("Chroot.Execute", "successfully ran.")
	return nil
}

// ExecuteCmds runs a list of commands in the chroot environment,
// stops at the first error
func (c *Chroot) ExecuteCmds(cmds []string) error {
	PrintVerboseInfo("Chroot.ExecuteCmds", "running...")

	for _, cmd := range cmds {
		err := c.Execute(cmd, []string{})
		if err != nil {
			PrintVerboseErr("Chroot.ExecuteCmds", 0, err)
			return err
		}
	}

	PrintVerboseInfo("Chroot.ExecuteCmds", "successfully ran.")
	return nil
}
