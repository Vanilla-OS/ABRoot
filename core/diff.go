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

// MergeDiff merges the diff lines between the first and second files into
// the destination file. If any errors occur, they are returned.
func MergeDiff(firstFile, secondFile, destination string) error {
	PrintVerboseInfo("MergeDiff", "merging", firstFile, "+", secondFile, "->", destination)

	// get the diff lines
	diffLines, err := DiffFiles(firstFile, secondFile)
	if err != nil {
		PrintVerboseErr("MergeDiff", 0, err)
		return err
	}

	// copy second file to destination to apply patch
	secondFileContents, err := os.ReadFile(secondFile)
	if err != nil {
		PrintVerboseErr("MergeDiff", 1, err)
		return err
	}
	err = os.WriteFile(destination, secondFileContents, 0644)
	if err != nil {
		PrintVerboseErr("MergeDiff", 2, err)
		return err
	}

	// write the diff to the destination
	err = WriteDiff(destination, diffLines)
	if err != nil {
		PrintVerboseErr("MergeDiff", 3, err)
		return err
	}

	PrintVerboseInfo("MergeDiff", "merge completed")
	return nil
}

// DiffFiles returns the diff lines between source and dest files using the
// diff command (assuming it is installed). If no diff is found, nil is
// returned. If any errors occur, they are returned.
func DiffFiles(sourceFile, destFile string) ([]byte, error) {
	PrintVerboseInfo("DiffFiles", "diffing", sourceFile, "and", destFile)

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

	// diff returns 1 if there are differences
	if errCode == 1 {
		PrintVerboseInfo("DiffFiles", "diff found")
		return out.Bytes(), nil
	}

	PrintVerboseInfo("DiffFiles", "no diff found")
	return nil, nil
}

// WriteDiff applies the diff lines to the destination file using the patch
// command (assuming it is installed). If any errors occur, they are returned.
func WriteDiff(destFile string, diffLines []byte) error {
	PrintVerboseInfo("WriteDiff", "applying diff to", destFile)
	if len(diffLines) == 0 {
		PrintVerboseInfo("WriteDiff", "no changes to apply")
		return nil // no changes to apply
	}

	cmd := exec.Command("patch", "-R", destFile)
	cmd.Stdin = bytes.NewReader(diffLines)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		PrintVerboseErr("WriteDiff", 0, err)
		return err
	}

	PrintVerboseInfo("WriteDiff", "diff applied")
	return nil
}
