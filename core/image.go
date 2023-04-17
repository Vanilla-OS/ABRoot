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
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// ABImage struct
type ABImage struct {
	Digest    string    `json:"digest"`
	Timestamp time.Time `json:"timestamp"`
	Image     string    `json:"image"`
}

// NewABImage returns a new ABImage struct
func NewABImage(digest string, image string) *ABImage {
	return &ABImage{
		Digest:    digest,
		Timestamp: time.Now(),
		Image:     image,
	}
}

// NewABImageFromRoot returns the current ABImage from /abimage.abr
func NewABImageFromRoot() (*ABImage, error) {
	PrintVerbose("NewABImageFromRoot: running...")

	abimage, err := ioutil.ReadFile("/abimage.abr")
	if err != nil {
		PrintVerbose("NewABImageFromRoot:error: " + err.Error())
		return nil, err
	}

	var a ABImage
	err = json.Unmarshal(abimage, &a)
	if err != nil {
		PrintVerbose("NewABImageFromRoot:error(2): " + err.Error())
		return nil, err
	}

	PrintVerbose("NewABImageFromRoot: found abimage.abr: " + a.Digest)
	return &a, nil
}

// WriteTo writes the json to a dest path
func (a *ABImage) WriteTo(dest string) error {
	PrintVerbose("ABImage.WriteTo: running...")

	dir := filepath.Dir(dest)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		PrintVerbose("ABImage.WriteTo:error: " + err.Error())
		return err
	}

	imageName := "abimage.abr"
	imagePath := filepath.Join(dir, imageName)

	abimage, err := json.Marshal(a)
	if err != nil {
		PrintVerbose("ABImage.WriteTo:error(2): " + err.Error())
		return err
	}

	err = ioutil.WriteFile(imagePath, abimage, 0644)
	if err != nil {
		PrintVerbose("ABImage.WriteTo:error(3): " + err.Error())
		return err
	}

	return nil
}
