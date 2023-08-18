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
	PrintVerbose("kargsCreateIfMissing: running...")

	if _, err := os.Stat(KargsPath); os.IsNotExist(err) {
		PrintVerbose("kargsCreateIfMissing: creating kargs file...")
		err = os.WriteFile(KargsPath, []byte(DefaultKargs), 0644)
		if err != nil {
			PrintVerbose("kargsCreateIfMissing:err: " + err.Error())
			return err
		}
	}

	return nil
}

// KargsWrite makes a backup of the current kargs file and then
// writes the new content to it
func KargsWrite(content string) error {
	PrintVerbose("KargsWrite: running...")

	err := kargsCreateIfMissing()
	if err != nil {
		PrintVerbose("KargsWrite:err: " + err.Error())
		return err
	}

	validated, err := KargsFormat(content)
	if err != nil {
		PrintVerbose("KargsWrite:err(2): " + err.Error())
		return err
	}

	err = KargsBackup()
	if err != nil {
		PrintVerbose("KargsWrite:err(3): " + err.Error())
		return err
	}

	err = os.WriteFile(KargsPath, []byte(validated), 0644)
	if err != nil {
		PrintVerbose("KargsWrite:err(4): " + err.Error())
		return err
	}

	return nil
}

// KargsBackup makes a backup of the current kargs file
func KargsBackup() error {
	PrintVerbose("KargsBackup: running...")

	content, err := KargsRead()
	if err != nil {
		PrintVerbose("KargsBackup:err: " + err.Error())
		return err
	}

	err = os.WriteFile(KargsPath+".bak", []byte(content), 0644)
	if err != nil {
		PrintVerbose("KargsBackup:err: " + err.Error())
		return err
	}

	return nil
}

// KargsRead reads the content of the kargs file
func KargsRead() (string, error) {
	PrintVerbose("KargsRead: running...")

	err := kargsCreateIfMissing()
	if err != nil {
		PrintVerbose("KargsRead:err: " + err.Error())
		return "", err
	}

	content, err := os.ReadFile(KargsPath)
	if err != nil {
		PrintVerbose("KargsRead:err(2): " + err.Error())
		return "", err
	}

	return string(content), nil
}

// KargsFormat formats the contents of the kargs file, ensuring that
// there are no duplicate entries, multiple spaces or trailing newline
func KargsFormat(content string) (string, error) {
	PrintVerbose("KargsValidate: running...")

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

	return strings.Join(kargs, " "), nil
}

// KargsEdit copies the kargs file to a temporary file and opens it in the
// user's preferred editor by querying the $EDITOR environment variable.
// Once closed, its contents are written back to the main kargs file.
// This function returns a boolean parameter indicating whether any changes
// were made to the kargs file.
func KargsEdit() (bool, error) {
	PrintVerbose("KargsEdit: running...")

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	err := kargsCreateIfMissing()
	if err != nil {
		PrintVerbose("KargsEdit:err: " + err.Error())
		return false, err
	}

	// Open a temporary file, so editors installed via apx can also be used
	PrintVerbose("KargsEdit: Copying kargs file to /tmp")
	err = CopyFile(KargsPath, KargsTmpFile)
	if err != nil {
		PrintVerbose("KargsEdit:err(2): " + err.Error())
		return false, err
	}

	// Call $EDITOR on temp file
	PrintVerbose("KargsEdit: Opening %s with %s", KargsTmpFile, editor)
	cmd := exec.Command(editor, KargsTmpFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		PrintVerbose("KargsEdit:err(3): " + err.Error())
		return false, err
	}

	content, err := os.ReadFile(KargsTmpFile)
	if err != nil {
		PrintVerbose("KargsEdit:err(4): " + err.Error())
		return false, err
	}

	// Check whether there were any changes
	ogContent, err := os.ReadFile(KargsPath)
	if err != nil {
		PrintVerbose("KargsEdit:err(5): " + err.Error())
		return false, err
	}
	if string(ogContent) == string(content) {
		PrintVerbose("KargsEdit: No changes were made to kargs, skipping save.")
		return false, nil
	}

	PrintVerbose("KargsEdit: Writing contents of %s to original location", KargsTmpFile)
	err = KargsWrite(string(content))
	if err != nil {
		PrintVerbose("KargsEdit:err(6): " + err.Error())
		return false, err
	}

	return true, nil
}
