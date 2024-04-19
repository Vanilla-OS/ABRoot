package tests

import (
	"testing"

	"github.com/vanilla-os/abroot/core"
)

// TestGetLogFile tests the GetLogFile function by getting the log file and
// checking if it is not nil.
func TestGetLogFile(t *testing.T) {
	t.Log("TestGetLogFile: running...")

	logFile := core.GetLogFile()
	if logFile == nil {
		t.Fatal("TestGetLogFile: logFile is nil")
	}

	t.Log("TestGetLogFile: done")
}

// TestWriteToLog tests the LogToFile function by writing a bunch of messages
// to the log file.
func TestWriteToLog(t *testing.T) {
	t.Log("TestWriteToLog: running...")
	err := core.LogToFile("TestWriteToLog: running...")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		t.Logf("TestWriteToLog: writing %d to the log file", i)
		err := core.LogToFile("TestWriteToLog: writing %d to the log file", i)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Log("TestWriteToLog: done")
	err = core.LogToFile("TestWriteToLog: done")
	if err != nil {
		t.Fatal(err)
	}
}
