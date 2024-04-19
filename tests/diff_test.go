package tests

import (
	"os"
	"testing"

	"github.com/vanilla-os/abroot/core"
)

// TestMergeDiff tests the MergeDiff function by creating 2 files with
// different content and merging them into a destination file. The destination
// file should reflect the changes.
func TestMergeDiff(t *testing.T) {
	firstFile := t.TempDir() + "/first"
	err := os.WriteFile(firstFile, []byte("first file"), 0644)
	if err != nil {
		t.Error(err)
	}

	secondFile := t.TempDir() + "/second"
	err = os.WriteFile(secondFile, []byte("second file"), 0644)
	if err != nil {
		t.Error(err)
	}

	destination := t.TempDir() + "/destination"

	err = core.MergeDiff(firstFile, secondFile, destination)
	if err != nil {
		t.Error(err)
	}

	firstFileContents, err := os.ReadFile(firstFile)
	if err != nil {
		t.Error(err)
	}

	destinationContents, err := os.ReadFile(destination)
	if err != nil {
		t.Error(err)
	}

	if string(firstFileContents) != string(destinationContents) {
		t.Error("destination file does not have the same content as the second file")
	}

	t.Log("TestMergeDiff: destination file:", string(destinationContents))
	t.Log("TestMergeDiff: done")
}

// TestDiffFiles tests the DiffFiles function in 2 cases:
// 1. when the source and dest files have the same content
// 2. when the source and dest files have different content
func TestDiffFiles(t *testing.T) {
	// same content
	sourceFile := t.TempDir() + "/source"
	err := os.WriteFile(sourceFile, []byte("same file"), 0644)
	if err != nil {
		t.Error(err)
	}

	destFile := t.TempDir() + "/dest"
	err = os.WriteFile(destFile, []byte("same file"), 0644)
	if err != nil {
		t.Error(err)
	}

	diffLines, err := core.DiffFiles(sourceFile, destFile)
	if err != nil {
		t.Error(err)
	}

	if diffLines != nil {
		t.Error("diff lines should be nil")
	}

	t.Log("No diff found. Expected behavior.")

	// different content
	err = os.WriteFile(destFile, []byte("different file"), 0644)
	if err != nil {
		t.Error(err)
	}

	diffLines, err = core.DiffFiles(sourceFile, destFile)
	if err != nil {
		t.Error(err)
	}

	if diffLines == nil {
		t.Error("diff lines should not be nil")
	}

	t.Log("Diff: ", string(diffLines))
	t.Log("Diff found. Expected behavior.")
}
