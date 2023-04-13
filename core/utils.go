package core

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
