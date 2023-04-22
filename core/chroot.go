package core

import (
	"os"
	"os/exec"
	"strings"
)

// Chroot is a struct which represents a chroot environment
type Chroot struct {
	root     string
	rootUuid string
}

var mounts = []string{
	"/dev",
	"/dev/pts",
	"/proc",
	"/run",
	"/sys",
}

// NewChroot creates a new chroot environment
func NewChroot(root string, rootUuid string) (*Chroot, error) {
	PrintVerbose("NewChroot: running...")

	if _, err := os.Stat(root); os.IsNotExist(err) {
		PrintVerbose("NewChroot:error: " + err.Error())
		return nil, err
	}

	for _, mount := range mounts {
		err := exec.Command("mount", "--bind", mount, root+mount).Run()
		if err != nil {
			PrintVerbose("NewChroot:error(2): " + err.Error())
			return nil, err
		}
	}

	// Workaround for a bug in os grub-probe, not detecting the
	// root partition
	err := exec.Command("mount", "-U", rootUuid, "/").Run()
	if err != nil {
		PrintVerbose("NewChroot:error(3): " + err.Error())
		return nil, err
	}

	return &Chroot{
		root:     root,
		rootUuid: rootUuid,
	}, nil
}

// Close unmounts all the bind mounts
func (c *Chroot) Close() error {
	PrintVerbose("Close: running...")

	for _, mount := range mounts {
		err := exec.Command("umount", c.root+mount).Run()
		if err != nil {
			PrintVerbose("Close:error: " + err.Error())
			return err
		}
	}

	return nil
}

// Execute runs a command in the chroot environment
func (c *Chroot) Execute(cmd string, args []string) error {
	PrintVerbose("Execute: running...")

	cmd = strings.Join(append([]string{cmd}, args...), " ")
	PrintVerbose("Execute: running command: " + cmd)
	e := exec.Command("chroot", c.root, "/bin/sh", "-c", cmd)
	e.Stdout = os.Stdout
	e.Stderr = os.Stderr
	e.Stdin = os.Stdin
	err := e.Run()
	if err != nil {
		PrintVerbose("Execute:error: " + err.Error())
		return err
	}

	PrintVerbose("Execute: successfully ran.")
	return nil
}

// ExecuteCmds runs a list of commands in the chroot environment,
// stops at the first error
func (c *Chroot) ExecuteCmds(cmds []string) error {
	PrintVerbose("ExecuteCmds: running...")

	for _, cmd := range cmds {
		err := c.Execute(cmd, []string{})
		if err != nil {
			PrintVerbose("ExecuteCmds:error: " + err.Error())
			return err
		}
	}

	PrintVerbose("ExecuteCmds: successfully ran.")
	return nil
}
