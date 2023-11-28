package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vanilla-os/abroot/settings"
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
