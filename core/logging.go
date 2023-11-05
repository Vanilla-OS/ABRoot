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
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/vanilla-os/orchid/cmdr"
)

var logFile *os.File

var printLog = log.New(os.Stdout, "(Verbose) ", 0)

func init() {
	PrintVerboseNoLog("NewLogFile: running...")

	incremental := 0

	logFiles, err := filepath.Glob("/var/log/abroot.log.*")
	if err != nil {
		incremental = 1
	} else {
		allIncrementals := []int{}
		for _, logFile := range logFiles {
			_, err := fmt.Sscanf(logFile, "/var/log/abroot.log.%d", &incremental)
			if err != nil {
				continue
			}
			allIncrementals = append(allIncrementals, incremental)
		}
		if len(allIncrementals) == 0 {
			incremental = 1
		} else {
			incremental = allIncrementals[len(allIncrementals)-1] + 1
		}
	}

	logFile, err = os.OpenFile(
		fmt.Sprintf("/var/log/abroot.log.%d", incremental),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666,
	)
	if err != nil {
		PrintVerboseNoLog("NewLogFile: error: %s", err)
	}
}

func IsVerbose() bool {
	flag := cmdr.FlagValBool("verbose")
	_, arg := os.LookupEnv("ABROOT_VERBOSE")
	return flag || arg
}

func PrintVerbose(msg string, args ...interface{}) {
	PrintVerboseNoLog(msg, args...)

	if logFile != nil {
		LogToFile(msg, args...)
	}
}

func PrintVerboseNoLog(msg string, args ...interface{}) {
	if IsVerbose() {
		printLog.Printf(msg, args...)
		printLog.Println()
	}
}

func LogToFile(msg string, args ...interface{}) error {
	if logFile != nil {
		_, err := fmt.Fprintf(
			logFile,
			"%s: %s\n",
			time.Now().Format("2006-01-02 1 15:04:05"),
			fmt.Sprintf(msg, args...),
		)
		return err
	}
	return nil
}

func GetLogFile() *os.File {
	return logFile
}
