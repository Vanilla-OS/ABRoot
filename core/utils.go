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
	"io"
	"os"
	"os/exec"
)

var abrootDir = "/etc/abroot"

func init() {
	if !RootCheck(false) {
		return
	}

	if _, err := os.Stat(abrootDir); os.IsNotExist(err) {
		err := os.Mkdir(abrootDir, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func RootCheck(display bool) bool {
	if os.Geteuid() != 0 {
		if display {
			fmt.Println("You must be root to run this command")
		}

		return false
	}

	return true
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		PrintVerbose("File exists: " + path)
		return true
	}

	PrintVerbose("File does not exist: " + path)
	return false
}

// isLink checks if a path is a link
func isLink(path string) bool {
	if _, err := os.Lstat(path); err == nil {
		PrintVerbose("Path is a link: " + path)
		return true
	}

	PrintVerbose("Path is not a link: " + path)
	return false
}

// CopyFile copies a file from source to dest
func CopyFile(source, dest string) error {
	PrintVerbose("CopyFile: running...")

	PrintVerbose("CopyFile: Opening source file")
	srcFile, err := os.Open(source)
	if err != nil {
		PrintVerbose("CopyFile:err: " + err.Error())
		return err
	}
	defer srcFile.Close()

	PrintVerbose("CopyFile: Opening destination file")
	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		PrintVerbose("CopyFile:err: " + err.Error())
		return err
	}
	defer destFile.Close()

	PrintVerbose("CopyFile: Performing copy operation")
	if _, err := io.Copy(destFile, srcFile); err != nil {
		PrintVerbose("CopyFile:err: " + err.Error())
		return err
	}

	return nil
}

// isDeviceLUKSEncrypted checks whether a device specified by devicePath is a LUKS-encrypted device
func isDeviceLUKSEncrypted(devicePath string) (bool, error) {
	PrintVerbose("Verifying if %s is encrypted", devicePath)

	isLuksCmd := "cryptsetup isLuks %s"

	cmd := exec.Command("sh", "-c", fmt.Sprintf(isLuksCmd, devicePath))
	err := cmd.Run()
	if err != nil {
		// We expect the command to return exit status 1 if partition isn't
		// LUKS-encrypted
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("Failed to check if %s is LUKS-encrypted: %s", devicePath, err)
	}

	return true, nil
}
