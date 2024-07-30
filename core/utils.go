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
	"fmt"
	"io"
	"io/fs"
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
		PrintVerboseInfo("fileExists", "File exists:", path)
		return true
	}

	PrintVerboseInfo("fileExists", "File does not exist:", path)
	return false
}

// isLink checks if a path is a link
func isLink(path string) bool {
	if _, err := os.Lstat(path); err == nil {
		PrintVerboseInfo("isLink", "Path is a link:", path)
		return true
	}

	PrintVerboseInfo("isLink", "Path is not a link:", path)
	return false
}

// CopyFile copies a file from source to dest
func CopyFile(source, dest string) error {
	PrintVerboseInfo("CopyFile", "Running...")

	PrintVerboseInfo("CopyFile", "Opening source file")
	srcFile, err := os.Open(source)
	if err != nil {
		PrintVerboseErr("CopyFile", 0, err)
		return err
	}
	defer srcFile.Close()

	PrintVerboseInfo("CopyFile", "Opening destination file")
	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		PrintVerboseErr("CopyFile", 1, err)
		return err
	}
	defer destFile.Close()

	PrintVerboseInfo("CopyFile", "Performing copy operation")
	if _, err := io.Copy(destFile, srcFile); err != nil {
		PrintVerboseErr("CopyFile", 2, err)
		return err
	}

	return nil
}

// isDeviceLUKSEncrypted checks whether a device specified by devicePath is a LUKS-encrypted device
func isDeviceLUKSEncrypted(devicePath string) (bool, error) {
	PrintVerboseInfo("isDeviceLUKSEncrypted", "Verifying if", devicePath, "is encrypted")

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
		err = fmt.Errorf("failed to check if %s is LUKS-encrypted: %s", devicePath, err)
		PrintVerboseErr("isDeviceLUKSEncrypted", 0, err)
		return false, err
	}

	return true, nil
}

// getDirSize calculates the total size of a directory recursively.
func getDirSize(path string) (int64, error) {
	ds, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if !ds.IsDir() {
		return 0, fmt.Errorf("%s is not a directory", path)
	}

	var totalSize int64 = 0

	dfs := os.DirFS(path)
	err = fs.WalkDir(dfs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			fileInfo, err := d.Info()
			if err != nil {
				return err
			}
			totalSize += fileInfo.Size()
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return totalSize, nil
}
