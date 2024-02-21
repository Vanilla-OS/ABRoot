package core

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"os"
)

func MergeSpecialFile(user string, old string, new string, out string) error {
	// Merges special files
	// Files get merged by first forming a diff between the old file and the user file
	// Then applying the generated patch to the new file
	// The new file then gets written to the given destination
	userData, err := os.ReadFile(user)
	if err != nil {
		return err
	}
	oldData, err := os.ReadFile(old)
	if err != nil {
		return err
	}
	newData, err := os.ReadFile(new)
	if err != nil {
		return err
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(oldData), string(userData), false)
	patches := dmp.PatchMake(string(oldData), diffs)

	result, _ := dmp.PatchApply(patches, string(newData))
	filePerms, err := os.Stat(new)
	if err != nil {
		return err
	}
	err = os.WriteFile(out, []byte(result), filePerms.Mode())
	if err != nil {
		return err
	}
	return nil
}
