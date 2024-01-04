package diff

import (
	"cmp"
	"regexp"
	"strconv"
	"sync"
)

type Package map[string]string

type PackageDiff struct {
	Name            string `json:"name"`
	NewVersion      string `json:"new_version,omitempty"`
	PreviousVersion string `json:"previous_version,omitempty"`
}

// This monstruosity is an adaptation of the regex for semver (available in https://semver.org/).
// It SHOULD be able to capture every type of exoteric versioning scheme out there.
var versionRegex = regexp.MustCompile(`^(?:(?P<prefix>\d+):)?(?P<major>\d+[a-zA-Z]?)(?:\.(?P<minor>\d+))?(?:\.(?P<patch>\d+))?(?:[-~](?P<prerelease>(?:\d+|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:\d+|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:[+.](?P<buildmetadata>[0-9a-zA-Z-+.~]+(?:\.[0-9a-zA-Z-]+)*))?$`)

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

func processPackage(pkg, oldVersion, newVersion string, c chan<- struct {
	PackageDiff
	int
}, wg *sync.WaitGroup) {
	defer wg.Done()

	result := CompareVersions(oldVersion, newVersion)
	c <- struct {
		PackageDiff
		int
	}{PackageDiff{pkg, newVersion, oldVersion}, result}
}

// PackageDiff returns the difference in packages between two images, organized into
// four slices: Added, Upgraded, Downgraded, and Removed packages, respectively.
func DiffPackages(oldPackages, newPackages Package) ([]PackageDiff, []PackageDiff, []PackageDiff, []PackageDiff) {
	var wg sync.WaitGroup
	c := make(chan struct {
		PackageDiff
		int
	}, len(oldPackages))

	for pkg, oldVersion := range oldPackages {
		wg.Add(1)

		if newVersion, ok := newPackages[pkg]; ok {
			go processPackage(pkg, oldVersion, newVersion, c, &wg)
		} else {
			c <- struct {
				PackageDiff
				int
			}{PackageDiff{pkg, oldVersion, ""}, 2}
			wg.Done()
		}
	}

	removed := []PackageDiff{}
	wg.Add(1)
	go func() {
		for pkg, version := range newPackages {
			if _, ok := oldPackages[pkg]; !ok {
				removed = append(removed, PackageDiff{pkg, "", version})
			}
		}
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(c)
	}()

	added := []PackageDiff{}
	upgraded := []PackageDiff{}
	downgraded := []PackageDiff{}
	for pkgResult := range c {
		switch pkgResult.int {
		case -1:
			downgraded = append(downgraded, pkgResult.PackageDiff)
		case 1:
			upgraded = append(upgraded, pkgResult.PackageDiff)
		case 2:
			added = append(added, pkgResult.PackageDiff)
		}
	}

	return added, upgraded, downgraded, removed
}
