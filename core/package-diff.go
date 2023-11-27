package core

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/vanilla-os/abroot/settings"
)

type PackageDiff struct {
	Added, Removed struct {
		Name, Version string
	}
	Updated, Downgraded struct {
		Name, OldVersion, NewVersion string
	}
}

func BaseImagePackageDiff(currentDigest, newDigest string) (PackageDiff, error) {
	PrintVerbose("BaseImagePackageDiff: running...")

	params := url.Values{}
	params.Add("old_digest", currentDigest)
	params.Add("new_digest", newDigest)
	resp, err := http.Get(settings.Cnf.DifferURL + "?" + params.Encode())
	if err != nil {
		PrintVerbose("PackageDiff.BaseImagePackageDiff:err: %s", err)
		return PackageDiff{}, err
	}
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		PrintVerbose("PackageDiff.BaseImagePackageDiff(1):err: %s", err)
		return PackageDiff{}, err
	}

	diff := PackageDiff{}
	err = json.Unmarshal(contents, &diff)
	if err != nil {
		PrintVerbose("PackageDiff.BaseImagePackageDiff(2):err: %s", err)
		return PackageDiff{}, err
	}

	return diff, nil
}
