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
	etcMounted bool
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
func NewChroot(root string, rootUuid string, rootDevice string, mountUserEtc bool, userEtcPath string) (*Chroot, error) {
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
		etcMounted: mountUserEtc,
	}

	// workaround for grub-mkconfig, not able to find the device
	// inside a chroot environment
	err := chroot.Execute("mount --bind / /")
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

	if mountUserEtc {
		err = syscall.Mount("overlay", filepath.Join(root, "etc"), "overlay", syscall.MS_RDONLY, "lowerdir="+userEtcPath+":"+filepath.Join(root, "/etc"))
		if err != nil {
			PrintVerboseErr("NewChroot", 3, "failed to mount user etc:", err)
			return nil, err
		}
	}

	PrintVerboseInfo("NewChroot", "successfully created.")
	return chroot, nil
}

// Close unmounts all the bind mounts and closes the chroot environment
func (c *Chroot) Close() error {
	PrintVerboseInfo("Chroot.Close", "running...")

	err := UnmountRecursive(c.root, 0)
	if err != nil {
		PrintVerboseErr("Chroot.Close", 0, err)
		return err
	}

	PrintVerboseInfo("Chroot.Close", "successfully closed.")
	return nil
}

// Execute runs a command in the chroot environment, the command is
// a string and the arguments are a list of strings. If an error occurs
// it is returned.
func (c *Chroot) Execute(cmd string) error {
	PrintVerboseInfo("Chroot.Execute", "running...")

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
		err := c.Execute(cmd)
		if err != nil {
			PrintVerboseErr("Chroot.ExecuteCmds", 0, err)
			return err
		}
	}

	PrintVerboseInfo("Chroot.ExecuteCmds", "successfully ran.")
	return nil
}
