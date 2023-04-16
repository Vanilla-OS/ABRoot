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

func IsVerbose() bool {
	_, res := os.LookupEnv("ABROOT_VERBOSE")
	return res
}

func PrintVerbose(msg string, args ...interface{}) {
	if IsVerbose() {
		fmt.Printf(msg, args...)
		fmt.Println()
	}
}
