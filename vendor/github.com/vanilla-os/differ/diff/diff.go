package diff

import (
	"cmp"
	"regexp"
	"strconv"
)

type Package map[string]string

type PackageDiff struct {
	Name            string `json:"name"`
	NewVersion      string `json:"new_version,omitempty"`
	PreviousVersion string `json:"previous_version,omitempty"`
}

// This monstruosity is an adaptation of the regex for semver (available in https://semver.org/).
// It SHOULD be able to capture every type of exoteric versioning scheme out there.
var versionRegex = regexp.MustCompile(`^(?:(?P<prefix>\d+):)?(?P<major>\d+[a-zA-Z]?)(?:\.(?P<minor>\d+))?(?:\.(?P<patch>\d+))?(?:[-~](?P<prerelease>(?:\d+|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:\d+|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:[+.](?P<buildmetadata>[0-9a-zA-Z-+.]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// compareVersions has the same behavior as cmp.Compare, but for package versions. It parses
// both version strings and checks for differences in major, minor, patch, pre-release, etc.
func CompareVersions(a, b string) int {
	aMatchStr := versionRegex.FindStringSubmatch(a)
	aMatch := make(map[string]string)
	for i, name := range versionRegex.SubexpNames() {
		if i != 0 && name != "" && aMatchStr[i] != "" {
			aMatch[name] = aMatchStr[i]
		}
	}

	bMatchStr := versionRegex.FindStringSubmatch(b)
	bMatch := make(map[string]string)
	for i, name := range versionRegex.SubexpNames() {
		if i != 0 && name != "" && bMatchStr[i] != "" {
			bMatch[name] = bMatchStr[i]
		}
	}

	compResult := 0

	compOrder := []string{"prefix", "major", "minor", "patch", "prerelease", "buildmetadata"}
	for _, comp := range compOrder {
		aValue, aOk := aMatch[comp]
		bValue, bOk := bMatch[comp]
		// If neither version has component or if they equal
		if !aOk && !bOk {
			continue
		}
		// If a has component but b doesn't, package was upgraded, unless it's prerelease
		if aOk && !bOk {
			if comp == "prerelease" {
				compResult = -1
			} else {
				compResult = 1
			}
			break
		}
		// If b has component but a doesn't, package was downgraded
		if !aOk && bOk {
			compResult = -1
			break
		}

		// If both have, do regular compare
		aValueInt, aErr := strconv.Atoi(aValue)
		bValueInt, bErr := strconv.Atoi(bValue)

		var abComp int
		if aErr == nil && bErr == nil {
			abComp = cmp.Compare(aValueInt, bValueInt)
		} else {
			abComp = cmp.Compare(aValue, bValue)
		}
		if abComp == 0 {
			continue
		}
		compResult = abComp
		break
	}

	return compResult
}

// PackageDiff returns the difference in packages between two images, organized into
// four slices: Added, Upgraded, Downgraded, and Removed packages, respectively.
func DiffPackages(oldPackages, newPackages Package) ([]PackageDiff, []PackageDiff, []PackageDiff, []PackageDiff) {
	c := make(chan struct {
		PackageDiff
		int
	})

	newPkgsCopy := make(Package, len(newPackages))
	for k, v := range newPackages {
		newPkgsCopy[k] = v
	}

	for pkg, oldVersion := range oldPackages {
		if newVersion, ok := newPkgsCopy[pkg]; ok {
			go func(diff PackageDiff) {
				result := CompareVersions(diff.PreviousVersion, diff.NewVersion)
				c <- struct {
					PackageDiff
					int
				}{diff, result}
			}(PackageDiff{pkg, newVersion, oldVersion})
		} else {
			go func(diff PackageDiff) {
				c <- struct {
					PackageDiff
					int
				}{diff, 2}
			}(PackageDiff{pkg, oldVersion, ""})
		}

		// Clear package from copy so we can later check for removed packages
		delete(newPkgsCopy, pkg)
	}

	removed := []PackageDiff{}
	for pkg, version := range newPkgsCopy {
		removed = append(removed, PackageDiff{pkg, "", version})
	}

	added := []PackageDiff{}
	upgraded := []PackageDiff{}
	downgraded := []PackageDiff{}
	for i := 0; i < len(oldPackages); i++ {
		pkg := <-c
		switch pkg.int {
		case -1:
			downgraded = append(downgraded, pkg.PackageDiff)
		case 1:
			upgraded = append(upgraded, pkg.PackageDiff)
		case 2:
			added = append(added, pkg.PackageDiff)
		}
	}

	return added, upgraded, downgraded, removed
}
