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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
	"github.com/vanilla-os/abroot/settings"
)

// Hooks struct
type Hooks struct {
	HooksPath string
	Pre       []Hook
	Post      []Hook
}

// Hook struct
type Hook struct {
	Name string
	Path string
}

// NewHooks returns a new Hooks struct
func NewHooks() *Hooks {
	h := &Hooks{
		HooksPath: settings.Cnf.HooksPath,
	}
	h.getHooks()
	return h
}

// getHooks populates the Hooks struct with the hooks
func (h *Hooks) getHooks() {
	PrintVerbose("Hooks.getHooks: Getting hooks ...")

	hooksPath := filepath.Join(h.HooksPath, "abroot-hooks")
	preHooksPath := filepath.Join(hooksPath, "pre")
	preHooks, err1 := ioutil.ReadDir(preHooksPath)
	if err1 != nil {
		PrintVerbose("Hooks.getHooks:warning: no pre hooks found")
	}

	postHooksPath := filepath.Join(hooksPath, "post")
	postHooks, err2 := ioutil.ReadDir(postHooksPath)
	if err2 != nil {
		PrintVerbose("Hooks.getHooks:warning: no post hooks found")
	}

	if err1 != nil && err2 != nil {
		PrintVerbose("Hooks.getHooks:warn(2): no hooks to load")
		return
	}

	sort.Slice(preHooks, func(i, j int) bool {
		return preHooks[i].Name() < preHooks[j].Name()
	})
	sort.Slice(postHooks, func(i, j int) bool {
		return postHooks[i].Name() < postHooks[j].Name()
	})

	for _, hook := range preHooks {
		h.Pre = append(h.Pre, Hook{
			Name: hook.Name(),
			Path: filepath.Join(preHooksPath, hook.Name()),
		})
	}
	for _, hook := range postHooks {
		h.Post = append(h.Post, Hook{
			Name: hook.Name(),
			Path: filepath.Join(postHooksPath, hook.Name()),
		})
	}

	PrintVerbose("Hooks.getHooks: %d pre hooks loaded", len(h.Pre))
	PrintVerbose("Hooks.getHooks: %d post hooks loaded", len(h.Post))
}

// finalScript creates the final script according to the requested event
func (h *Hooks) FinalScript(event string) (string, error) {
	PrintVerbose("Hooks.FinalScript: Creating final script for %s hooks ...", event)

	var hooks []Hook
	switch event {
	case "pre":
		hooks = h.Pre
	case "post":
		hooks = h.Post
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		PrintVerbose("Hooks.FinalScript:err: %s", err)
		return "", err
	}

	finalScriptPath := filepath.Join(tmpDir, fmt.Sprintf("abh-%s-%s.sh", event, uuid.New()))
	finalScript, err := os.Create(finalScriptPath)
	if err != nil {
		PrintVerbose("Hooks.FinalScript:err: %s", err)
		return "", err
	}
	defer finalScript.Close()

	for _, hook := range hooks {
		finalScript.WriteString(fmt.Sprintf("source %s\n", hook.Path))
	}

	PrintVerbose("Hooks.FinalScript: Final script created at %s", finalScriptPath)

	return finalScriptPath, nil
}
