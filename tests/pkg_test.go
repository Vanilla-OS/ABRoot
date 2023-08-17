package tests

import (
	"strings"
	"testing"

	"github.com/vanilla-os/abroot/core"
)

func TestPackageManager(t *testing.T) {
	pm := core.NewPackageManager(true)

	// Add a package
	pkg := "bash htop"
	err := pm.Add(pkg)
	if err != nil {
		t.Error(err)
	}

	// Check if package is in packages.add
	pkgs, err := pm.GetAddPackages()
	if err != nil {
		t.Error(err)
	}

	found := false
	for _, p := range pkgs {
		if p == pkg {
			found = true
			break
		}
	}

	if !found {
		t.Error("package was not added to packages.add")
	}

	// Get final cmd
	cmd := pm.GetFinalCmd(core.APPLY)
	if len(cmd) == 0 {
		t.Error("final cmd is empty")
	}

	// Clear unstaged packages
	err = pm.ClearUnstagedPackages()
	if err != nil {
		t.Error(err)
	}

	// Check if packages.unstaged is empty
	upkgs, err := pm.GetUnstagedPackages()
	if err != nil {
		t.Error(err)
	}

	if len(upkgs) != 0 {
		t.Error("packages.unstaged was not cleared")
	}

	// Check if package exists in repo
	for _, _pkg := range strings.Split(pkg, " ") {
		err = pm.ExistsInRepo(_pkg)
		if err != nil {
			t.Error(err)
		}
	}
}
