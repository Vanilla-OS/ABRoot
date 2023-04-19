package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ChrootExecute runs a command in a chroot environment
func ChrootExecute(root string, cmd string, args []string) error {
	PrintVerbose("ChrootExecute: running...")

	// check if root exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		PrintVerbose("ChrootExecute:error: " + err.Error())
		return err
	}

	// run command
	cmd = filepath.Join(root, cmd)
	cmd = strings.Join(append([]string{cmd}, args...), " ")
	PrintVerbose("ChrootExecute: running command: " + cmd)
	c := exec.Command("chroot", root, "/bin/sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	err := c.Run()
	if err != nil {
		PrintVerbose("ChrootExecute:error: " + err.Error())
		return err
	}

	PrintVerbose("ChrootExecute: successfully ran.")
	return nil
}

// ChrootExecuteCmds runs a list of commands in a chroot environment,
// stops at the first error
func ChrootExecuteCmds(root string, cmds []string) error {
	PrintVerbose("ChrootExecuteCmds: running...")

	// run commands
	for _, cmd := range cmds {
		err := ChrootExecute(root, cmd, []string{})
		if err != nil {
			PrintVerbose("ChrootExecuteCmds:error: " + err.Error())
			return err
		}
	}

	PrintVerbose("ChrootExecuteCmds: successfully ran.")
	return nil
}
