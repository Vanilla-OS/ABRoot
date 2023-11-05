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
	"bytes"
	"os"
	"os/exec"
)

// MergeDiff merges the diff lines between the first and second files into destination
func MergeDiff(firstFile, secondFile, destination string) error {
	PrintVerbose("MergeDiff: merging %s + %s -> %s", firstFile, secondFile, destination)

	// get the diff lines
	diffLines, err := DiffFiles(firstFile, secondFile)
	if err != nil {
		PrintVerbose("MergeDiff:err: %s", err)
		return err
	}

	// copy second file to destination to apply patch
	secondFileContents, err := os.ReadFile(secondFile)
	if err != nil {
		PrintVerbose("MergeDiff:err(2): %s", err)
		return err
	}
	err = os.WriteFile(destination, secondFileContents, 0644)
	if err != nil {
		PrintVerbose("MergeDiff:err(3): %s", err)
		return err
	}

	// write the diff to the destination
	err = WriteDiff(destination, diffLines)
	if err != nil {
		PrintVerbose("MergeDiff:err(4): %s", err)
		return err
	}

	PrintVerbose("MergeDiff: merge completed")
	return nil
}

// DiffFiles returns the diff lines between source and dest files.
func DiffFiles(sourceFile, destFile string) ([]byte, error) {
	PrintVerbose("DiffFiles: diffing %s -> %s", sourceFile, destFile)

	cmd := exec.Command("diff", "-u", sourceFile, destFile)
	var out bytes.Buffer
	cmd.Stdout = &out
	errCode := 0
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errCode = exitError.ExitCode()
		}
	}

	if errCode == 1 {
		PrintVerbose("DiffFiles: diff found")
		return out.Bytes(), nil
	}

	PrintVerbose("DiffFiles: no diff found")
	return nil, nil
}

// WriteDiff applies the diff lines to the destination file.
func WriteDiff(destFile string, diffLines []byte) error {
	PrintVerbose("WriteDiff: applying diff to %s", destFile)
	if len(diffLines) == 0 {
		PrintVerbose("WriteDiff: no changes to apply")
		return nil // no changes to apply
	}

	cmd := exec.Command("patch", "-R", destFile)
	cmd.Stdin = bytes.NewReader(diffLines)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		PrintVerbose("WriteDiff:err: %s", err)
		return err
	}

	PrintVerbose("WriteDiff: diff applied")
	return nil
}
