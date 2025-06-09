package core

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/viper"
	"github.com/vanilla-os/abroot/settings"
)

// Supported results for the ConfEditResult type
const (
	CONF_CHANGED = iota
	CONF_UNCHANGED
	CONF_FAILED
)

// ConfEditResult is the result of the ConfEdit function
type ConfEditResult int

// ConfEdit opens the configuration file in the default editor
func ConfEdit() (ConfEditResult, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		nanoBin, err := exec.LookPath("nano")
		if err == nil {
			editor = nanoBin
		}

		viBin, err := exec.LookPath("vi")
		if err == nil {
			editor = viBin
		}

		editorBin, err := exec.LookPath("editor")
		if err == nil {
			editor = editorBin
		}

		if editor == "" {
			return CONF_FAILED, fmt.Errorf("no editor found in $EDITOR, nano or vi")
		}
	}

	// getting the configuration content so as we can compare it later
	// to see if it has been changed
	cnfContent, err := os.ReadFile(settings.CnfPathAdmin)
	if err != nil {
		return CONF_FAILED, err
	}

	tmpFilePath := settings.CnfPathAdmin + ".tmp.json"

	err = os.WriteFile(tmpFilePath, cnfContent, 0o755)
	if err != nil {
		return CONF_FAILED, err
	}
	defer os.Remove(tmpFilePath)

	// open the editor
	cmd := exec.Command(editor, tmpFilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return CONF_FAILED, err
	}

	// getting the new content
	newCnfContent, err := os.ReadFile(tmpFilePath)
	if err != nil {
		return CONF_FAILED, err
	}

	// we compare the old and new content to return the proper result
	if string(cnfContent) == string(newCnfContent) {
		return CONF_UNCHANGED, nil
	}

	viper.SetConfigFile(tmpFilePath)
	err = viper.ReadInConfig()
	if err != nil {
		return CONF_FAILED, err
	}

	err = AtomicSwap(settings.CnfPathAdmin, tmpFilePath)
	if err != nil {
		return CONF_FAILED, err
	}

	return CONF_CHANGED, nil
}
