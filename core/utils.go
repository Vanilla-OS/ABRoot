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
