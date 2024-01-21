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
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/vanilla-os/orchid/cmdr"
)

// logFile is a file handle for the log file
var logFile *os.File

// printLog is a logger to Stdout for verbose information
var printLog = log.New(os.Stdout, "(Verbose) ", 0)

// init initializes the log file and sets up logging
func init() {
	PrintVerboseInfo("NewLogFile", "running...")

	// Incremental value to append to log file name
	incremental := 0

	// Check for existing log files
	logFiles, err := filepath.Glob("/var/log/abroot.log.*")
	if err != nil {
		// If there are no log files, start with incremental 1
		incremental = 1
	} else {
		allIncrementals := []int{}
		// Extract incremental values from existing log file names
		for _, logFile := range logFiles {
			_, err := fmt.Sscanf(logFile, "/var/log/abroot.log.%d", &incremental)
			if err != nil {
				continue
			}
			allIncrementals = append(allIncrementals, incremental)
		}
		// Set incremental to the next available value
		if len(allIncrementals) == 0 {
			incremental = 1
		} else {
			incremental = allIncrementals[len(allIncrementals)-1] + 1
		}
	}

	// Open or create the log file
	logFile, err = os.OpenFile(
		fmt.Sprintf("/var/log/abroot.log.%d", incremental),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666,
	)
	if err != nil {
		PrintVerboseErrNoLog("NewLogFile", 0, "failed to open log file", err)
	}
}

// IsVerbose checks if verbose mode is enabled
func IsVerbose() bool {
	flag := cmdr.FlagValBool("verbose")
	_, arg := os.LookupEnv("ABROOT_VERBOSE")
	return flag || arg
}

// formatMessage formats log messages based on prefix, level, and depth
func formatMessage(prefix, level string, depth float32, args ...interface{}) string {
	if prefix == "" && level == "" && depth == -1 {
		return fmt.Sprint(args...)
	}

	if depth > -1 {
		level = fmt.Sprintf("%s(%f)", level, depth)
	}
	return fmt.Sprintf("%s:%s:%s", prefix, level, fmt.Sprint(args...))
}

// printFormattedMessage prints formatted log messages to Stdout
func printFormattedMessage(formattedMsg string) {
	printLog.Printf("%s\n", formattedMsg)
}

// logToFileIfEnabled logs messages to the file if logging is enabled
func logToFileIfEnabled(formattedMsg string) {
	if logFile != nil {
		LogToFile(formattedMsg)
	}
}

// PrintVerboseNoLog prints verbose messages without logging to the file
func PrintVerboseNoLog(prefix, level string, depth float32, args ...interface{}) {
	if IsVerbose() {
		formattedMsg := formatMessage(prefix, level, depth, args...)
		printFormattedMessage(formattedMsg)
	}
}

// PrintVerbose prints verbose messages and logs to the file if enabled
func PrintVerbose(prefix, level string, depth float32, args ...interface{}) {
	PrintVerboseNoLog(prefix, level, depth, args...)

	logToFileIfEnabled(formatMessage(prefix, level, depth, args...))
}

// PrintVerboseSimpleNoLog prints simple verbose messages without logging to the file
func PrintVerboseSimpleNoLog(args ...interface{}) {
	PrintVerboseNoLog("", "", -1, args...)
}

// PrintVerboseSimple prints simple verbose messages and logs to the file if enabled
func PrintVerboseSimple(args ...interface{}) {
	PrintVerbose("", "", -1, args...)
}

// PrintVerboseErrNoLog prints verbose error messages without logging to the file
func PrintVerboseErrNoLog(prefix string, depth float32, args ...interface{}) {
	PrintVerboseNoLog(prefix, "err", depth, args...)
}

// PrintVerboseErr prints verbose error messages and logs to the file if enabled
func PrintVerboseErr(prefix string, depth float32, args ...interface{}) {
	PrintVerbose(prefix, "err", depth, args...)
}

// PrintVerboseWarnNoLog prints verbose warning messages without logging to the file
func PrintVerboseWarnNoLog(prefix string, depth float32, args ...interface{}) {
	PrintVerboseNoLog(prefix, "warn", depth, args...)
}

// PrintVerboseWarn prints verbose warning messages and logs to the file if enabled
func PrintVerboseWarn(prefix string, depth float32, args ...interface{}) {
	PrintVerbose(prefix, "warn", depth, args...)
}

// PrintVerboseInfoNoLog prints verbose info messages without logging to the file
func PrintVerboseInfoNoLog(prefix string, args ...interface{}) {
	PrintVerboseNoLog(prefix, "info", -1, args...)
}

// PrintVerboseInfo prints verbose info messages and logs to the file if enabled
func PrintVerboseInfo(prefix string, args ...interface{}) {
	PrintVerbose(prefix, "info", -1, args...)
}

// LogToFile writes messages to the log file
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

// GetLogFile returns the log file handle
func GetLogFile() *os.File {
	return logFile
}
