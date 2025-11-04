package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
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

	digest "github.com/opencontainers/go-digest"
	"github.com/vanilla-os/abroot/extras/dpkg"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/differ/diff"
)

// BaseImagePackageDiff retrieves the added, removed, upgraded and downgraded
// base packages (the ones bundled with the image).
func BaseImagePackageDiff(currentDigest, newDigest digest.Digest) (
	added, upgraded, downgraded, removed []diff.PackageDiff,
	err error,
) {
	PrintVerboseInfo("PackageDiff.BaseImagePackageDiff", "running...")

	imageComponents := strings.Split(settings.Cnf.Name, "/")
	imageName := imageComponents[len(imageComponents)-1]
	reqUrl := fmt.Sprintf("%s/images/%s/diff", settings.Cnf.DifferURL, imageName)
	body := fmt.Sprintf("{\"old_digest\": \"%s\", \"new_digest\": \"%s\"}", currentDigest, newDigest)

	PrintVerboseInfo("PackageDiff.BaseImagePackageDiff", "Requesting base image diff to", reqUrl, "with body", body)

	request, err := http.NewRequest(http.MethodGet, reqUrl, strings.NewReader(body))
	if err != nil {
		PrintVerboseErr("PackageDiff.BaseImagePackageDiff", 0, err)
		return
	}
	defer request.Body.Close()

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		PrintVerboseErr("PackageDiff.BaseImagePackageDiff", 1, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		PrintVerboseErr("PackageDiff.BaseImagePackageDiff", 2, fmt.Errorf("received non-OK status %s", resp.Status))
		return
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		PrintVerboseErr("PackageDiff.BaseImagePackageDiff", 3, err)
		return
	}

	pkgDiff := struct {
		Added, Upgraded, Downgraded, Removed []diff.PackageDiff
	}{}
	err = json.Unmarshal(contents, &pkgDiff)
	if err != nil {
		PrintVerboseErr("PackageDiff.BaseImagePackageDiff", 4, err)
		return
	}

	added = pkgDiff.Added
	upgraded = pkgDiff.Upgraded
	downgraded = pkgDiff.Downgraded
	removed = pkgDiff.Removed

	return
}

// OverlayPackageDiff retrieves the added, removed, upgraded and downgraded
// overlay packages (the ones added manually via `abroot pkg add`).
func OverlayPackageDiff() (
	added, upgraded, downgraded, removed []diff.PackageDiff,
	err error,
) {
	PrintVerboseInfo("OverlayPackageDiff", "running...")

	pkgM, err := NewPackageManager(false)
	if err != nil {
		PrintVerboseErr("OverlayPackageDiff", 0, err)
		return
	}

	addedPkgs, err := pkgM.GetAddPackages()
	if err != nil {
		PrintVerboseErr("PackageDiff.OverlayPackageDiff", 0, err)
		return
	}

	localAddedVersions := dpkg.DpkgBatchGetPackageVersion(addedPkgs)
	localAdded := map[string]string{}
	for i := 0; i < len(addedPkgs); i++ {
		if localAddedVersions[i] != "" {
			localAdded[addedPkgs[i]] = localAddedVersions[i]
		}
	}

	remoteAdded := map[string]string{}
	var pkgInfo map[string]interface{}
	for pkgName := range localAdded {
		pkgInfo, err = GetRepoContentsForPkg(pkgName)
		if err != nil {
			PrintVerboseErr("PackageDiff.OverlayPackageDiff", 1, err)
			return
		}
		version, ok := pkgInfo["version"].(string)
		if !ok {
			err = fmt.Errorf("unexpected value when retrieving upstream version of '%s'", pkgName)
			return
		}
		remoteAdded[pkgName] = version
	}

	added, upgraded, downgraded, removed = diff.DiffPackages(localAdded, remoteAdded)
	return
}
