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
	"os"
	"path/filepath"
	"time"
)

// The ABImage is the representation of an OCI image used by ABRoot, it
// contains the digest, the timestamp and the image name. If you need to
// investigate the current ABImage on an ABRoot system, you can find it
// at /abimage.abr
type ABImage struct {
	Digest    string    `json:"digest"`
	Timestamp time.Time `json:"timestamp"`
	Image     string    `json:"image"`
}

// NewABImage creates a new ABImage instance and returns a pointer to it,
// if the digest is empty, it returns an error
func NewABImage(digest string, image string) (*ABImage, error) {
	if digest == "" {
		return nil, fmt.Errorf("NewABImage: digest is empty")
	}

	return &ABImage{
		Digest:    digest,
		Timestamp: time.Now(),
		Image:     image,
	}, nil
}

// NewABImageFromRoot returns the current ABImage by parsing /abimage.abr, if
// it fails, it returns an error (e.g. if the file doesn't exist).
// Note for distro maintainers: if the /abimage.abr is not present, it could
// mean that the user is running an older version of ABRoot (pre v2) or the
// root state is corrupted. In the latter case, generating a new ABImage should
// fix the issue, Digest and Timestamp can be random, but Image should reflect
// an existing image on the configured Docker registry. Anyway, support on this
// is not guaranteed, so please don't open issues about this.
func NewABImageFromRoot() (*ABImage, error) {
	PrintVerboseInfo("NewABImageFromRoot", "running...")

	abimage, err := os.ReadFile("/abimage.abr")
	if err != nil {
		PrintVerboseErr("NewABImageFromRoot", 0, err)
		return nil, err
	}

	var a ABImage
	err = json.Unmarshal(abimage, &a)
	if err != nil {
		PrintVerboseErr("NewABImageFromRoot", 1, err)
		return nil, err
	}

	PrintVerboseInfo("NewABImageFromRoot", "found abimage.abr: "+a.Digest)
	return &a, nil
}

// WriteTo writes the json to a destination path, if the suffix is not empty,
// it will be appended to the filename
func (a *ABImage) WriteTo(dest string, suffix string) error {
	PrintVerboseInfo("ABImage.WriteTo", "running...")

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			PrintVerboseErr("ABImage.WriteTo", 0, err)
			return err
		}
	}

	if suffix != "" {
		suffix = "-" + suffix
	}
	imageName := "abimage" + suffix + ".abr"
	imagePath := filepath.Join(dest, imageName)

	abimage, err := json.Marshal(a)
	if err != nil {
		PrintVerboseErr("ABImage.WriteTo", 1, err)
		return err
	}

	err = os.WriteFile(imagePath, abimage, 0644)
	if err != nil {
		PrintVerboseErr("ABImage.WriteTo", 2, err)
		return err
	}

	PrintVerboseInfo("ABImage.WriteTo", "successfully wrote abimage.abr to "+imagePath)

	return nil
}
