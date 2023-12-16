package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/abroot/settings"
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

func TestBaseImagePackageDiff(t *testing.T) {
	settings.Cnf.Name = "vanilla-os/pico"

	oldDigest := "sha256:a99e4593b23fd07e3761639e9db38c0315e198d6e39dad6070e0e0e88be3de0b"
	newDigest := "sha256:a99e4593b23fd07e3761639e9db38c0315e198d6e39dad6070e0e0e88be3de0c"

	added, upgraded, downgraded, removed, err := core.BaseImagePackageDiff(oldDigest, newDigest)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Added: %v\n", added)
	fmt.Printf("Upgraded: %v\n", upgraded)
	fmt.Printf("Downgraded: %v\n", downgraded)
	fmt.Printf("Removed: %v\n", removed)
}

func TestOverlayPackageDiff(t *testing.T) {
	added, upgraded, downgraded, removed, err := core.OverlayPackageDiff()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Added: %v\n", added)
	fmt.Printf("Upgraded: %v\n", upgraded)
	fmt.Printf("Downgraded: %v\n", downgraded)
	fmt.Printf("Removed: %v\n", removed)
}
