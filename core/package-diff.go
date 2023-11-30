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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vanilla-os/abroot/extras/dpkg"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/differ/diff"
)

type PackageDiff struct {
	Added []struct {
		Name       string
		NewVersion string `json:"new_version"`
	}
	Removed []struct {
		Name       string
		OldVersion string `json:"old_version"`
	}
	Upgraded, Downgraded []struct {
		Name       string
		OldVersion string `json:"new_version"`
		NewVersion string `json:"old_version"`
	}
}

// BaseImagePackageDiff retrieves the added, removed, upgraded and downgraded
// base packages (the ones bundled with the image).
func BaseImagePackageDiff(currentDigest, newDigest string) (PackageDiff, error) {
	PrintVerbose("BaseImagePackageDiff: running...")

	imageComponents := strings.Split(settings.Cnf.Name, "/")
	imageName := imageComponents[len(imageComponents)-1]
	reqUrl := fmt.Sprintf("%s/images/%s/diff", settings.Cnf.DifferURL, imageName)
	body := fmt.Sprintf("{\"old_digest\": \"%s\", \"new_digest\": \"%s\"}", currentDigest, newDigest)

	request, err := http.NewRequest(http.MethodGet, reqUrl, strings.NewReader(body))
	if err != nil {
		PrintVerbose("PackageDiff.BaseImagePackageDiff:err: %s", err)
		return PackageDiff{}, err
	}
	defer request.Body.Close()

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		PrintVerbose("PackageDiff.BaseImagePackageDiff(1):err: %s", err)
		return PackageDiff{}, err
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		PrintVerbose("PackageDiff.BaseImagePackageDiff(2):err: %s", err)
		return PackageDiff{}, err
	}

	diff := PackageDiff{}
	err = json.Unmarshal(contents, &diff)
	if err != nil {
		PrintVerbose("PackageDiff.BaseImagePackageDiff(3):err: %s", err)
		return PackageDiff{}, err
	}

	return diff, nil
}

func OverlayPackageDiff() (
	added, upgraded, downgraded, removed []diff.PackageDiff,
	err error,
) {
	PrintVerbose("OverlayPackageDiff: running...")

	pkgM := NewPackageManager(false)
	addedPkgs, err := pkgM.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageDiff.OverlayPackageDiff:err: %s", err)
		return
	}

	localAddedVersions := dpkg.DpkgBatchGetPackageVersion(addedPkgs)
	localAdded := map[string]string{}
	for i := 0; i < len(addedPkgs); i++ {
		localAdded[addedPkgs[i]] = localAddedVersions[i]
	}

	remoteAdded := map[string]string{}
	version := ""
	for i := 0; i < len(addedPkgs); i++ {
		version, err = pkgM.GetRepoContentsForPkg(addedPkgs[i])
		if err != nil {
			PrintVerbose("PackageDiff.OverlayPackageDiff(1):err: %s", err)
			return
		}
		remoteAdded[addedPkgs[i]] = version
	}

	added, upgraded, downgraded, removed = diff.DiffPackages(localAdded, remoteAdded)
	return
}
