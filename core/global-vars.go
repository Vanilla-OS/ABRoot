package core

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
