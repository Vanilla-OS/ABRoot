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
	"io/ioutil"
	"os"
	"strings"
)

var KargsPath = "/etc/abroot/kargs"

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
		err = ioutil.WriteFile(KargsPath, []byte(""), 0644)
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

	validated, err := KargsValidate(content)
	if err != nil {
		PrintVerbose("KargsWrite:err(2): " + err.Error())
		return err
	}

	err = KargsBackup()
	if err != nil {
		PrintVerbose("KargsWrite:err(3): " + err.Error())
		return err
	}

	err = ioutil.WriteFile(KargsPath, []byte(validated), 0644)
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

	err = ioutil.WriteFile(KargsPath+".bak", []byte(content), 0644)
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
		PrintVerbose("KargsWrite:err: " + err.Error())
		return "", err
	}

	content, err := ioutil.ReadFile(KargsPath)
	if err != nil {
		PrintVerbose("KargsRead:err(2): " + err.Error())
		return "", err
	}

	return string(content), nil
}

// KargsValidate validates the content of the kargs file ensuring that
// there are no duplicate entries and multiple newlines
func KargsValidate(content string) (string, error) {
	PrintVerbose("KargsValidate: running...")

	lines := strings.Split(content, "\n")
	validated := ""

	for _, line := range lines {
		if line == "" {
			continue
		}

		if strings.Contains(validated, line) {
			continue
		}

		validated += line + "\n"
	}

	return validated, nil
}
