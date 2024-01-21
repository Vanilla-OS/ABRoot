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
	"os"
	"os/exec"
	"strings"
)

var KargsPath = "/etc/abroot/kargs"

const (
	DefaultKargs = "quiet splash bgrt_disable $vt_handoff"
	KargsTmpFile = "/tmp/kargs-temp"
)

func init() {
	if os.Getenv("ABROOT_KARGS_PATH") != "" {
		KargsPath = os.Getenv("ABROOT_KARGS_PATH")
	}
}

// kargsCreateIfMissing creates the kargs file if it doesn't exist
func kargsCreateIfMissing() error {
	PrintVerboseInfo("kargsCreateIfMissing", "running...")

	if _, err := os.Stat(KargsPath); os.IsNotExist(err) {
		PrintVerboseInfo("kargsCreateIfMissing", "creating kargs file...")
		err = os.WriteFile(KargsPath, []byte(DefaultKargs), 0644)
		if err != nil {
			PrintVerboseErr("kargsCreateIfMissing", 0, err)
			return err
		}
	}

	PrintVerboseInfo("kargsCreateIfMissing", "done")
	return nil
}

// KargsWrite makes a backup of the current kargs file and then
// writes the new content to it
func KargsWrite(content string) error {
	PrintVerboseInfo("KargsWrite", "running...")

	err := kargsCreateIfMissing()
	if err != nil {
		PrintVerboseErr("KargsWrite", 0, err)
		return err
	}

	validated, err := KargsFormat(content)
	if err != nil {
		PrintVerboseErr("KargsWrite", 1, err)
		return err
	}

	err = KargsBackup()
	if err != nil {
		PrintVerboseErr("KargsWrite", 2, err)
		return err
	}

	err = os.WriteFile(KargsPath, []byte(validated), 0644)
	if err != nil {
		PrintVerboseErr("KargsWrite", 3, err)
		return err
	}

	PrintVerboseInfo("KargsWrite", "done")
	return nil
}

// KargsBackup makes a backup of the current kargs file
func KargsBackup() error {
	PrintVerboseInfo("KargsBackup", "running...")

	content, err := KargsRead()
	if err != nil {
		PrintVerboseErr("KargsBackup", 0, err)
		return err
	}

	err = os.WriteFile(KargsPath+".bak", []byte(content), 0644)
	if err != nil {
		PrintVerboseErr("KargsBackup", 1, err)
		return err
	}

	PrintVerboseInfo("KargsBackup", "done")
	return nil
}

// KargsRead reads the content of the kargs file
func KargsRead() (string, error) {
	PrintVerboseInfo("KargsRead", "running...")

	err := kargsCreateIfMissing()
	if err != nil {
		PrintVerboseErr("KargsRead", 0, err)
		return "", err
	}

	content, err := os.ReadFile(KargsPath)
	if err != nil {
		PrintVerboseErr("KargsRead", 1, err)
		return "", err
	}

	PrintVerboseInfo("KargsRead", "done")
	return string(content), nil
}

// KargsFormat formats the contents of the kargs file, ensuring that
// there are no duplicate entries, multiple spaces or trailing newline
func KargsFormat(content string) (string, error) {
	PrintVerboseInfo("KargsValidate", "running...")

	kargs := []string{}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		lineArgs := strings.Split(line, " ")
		for _, larg := range lineArgs {
			// Check for duplicates
			isDuplicate := false
			for _, ka := range kargs {
				if ka == larg {
					isDuplicate = true
					break
				}
			}

			if !isDuplicate {
				kargs = append(kargs, larg)
			}
		}
	}

	PrintVerboseInfo("KargsValidate", "done")
	return strings.Join(kargs, " "), nil
}

// KargsEdit copies the kargs file to a temporary file and opens it in the
// user's preferred editor by querying the $EDITOR environment variable.
// Once closed, its contents are written back to the main kargs file.
// This function returns a boolean parameter indicating whether any changes
// were made to the kargs file.
func KargsEdit() (bool, error) {
	PrintVerboseInfo("KargsEdit", "running...")

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	err := kargsCreateIfMissing()
	if err != nil {
		PrintVerboseErr("KargsEdit", 0, err)
		return false, err
	}

	// Open a temporary file, so editors installed via apx can also be used
	PrintVerboseInfo("KargsEdit", "Copying kargs file to /tmp")
	err = CopyFile(KargsPath, KargsTmpFile)
	if err != nil {
		PrintVerboseErr("KargsEdit", 1, err)
		return false, err
	}

	// Call $EDITOR on temp file
	PrintVerboseInfo("KargsEdit", "Opening", KargsTmpFile, "in", editor)
	cmd := exec.Command(editor, KargsTmpFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		PrintVerboseErr("KargsEdit", 2, err)
		return false, err
	}

	content, err := os.ReadFile(KargsTmpFile)
	if err != nil {
		PrintVerboseErr("KargsEdit", 3, err)
		return false, err
	}

	// Check whether there were any changes
	ogContent, err := os.ReadFile(KargsPath)
	if err != nil {
		PrintVerboseErr("KargsEdit", 4, err)
		return false, err
	}
	if string(ogContent) == string(content) {
		PrintVerboseInfo("KargsEdit", "No changes were made to kargs, skipping save.")
		return false, nil
	}

	PrintVerboseInfo("KargsEdit", "Writing contents of", KargsTmpFile, "to the original location")
	err = KargsWrite(string(content))
	if err != nil {
		PrintVerboseErr("KargsEdit", 5, err)
		return false, err
	}

	PrintVerboseInfo("KargsEdit", "Done")
	return true, nil
}
