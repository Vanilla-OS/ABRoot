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
	"os"
	"time"
)

// ABImage struct
type ABImage struct {
	Digest    string    `json:"digest"`
	Timestamp time.Time `json:"timestamp"`
	Image     string    `json:"image"`
}

// NewABImage returns a new ABImage struct
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

// NewABImageFromRoot returns the current ABImage from /abimage.abr
func NewABImageFromRoot() (*ABImage, error) {
	PrintVerbose("NewABImageFromRoot: running...")

	abimage, err := os.ReadFile("/abimage.abr")
	if err != nil {
		PrintVerbose("NewABImageFromRoot:err: " + err.Error())
		return nil, err
	}

	var a ABImage
	err = json.Unmarshal(abimage, &a)
	if err != nil {
		PrintVerbose("NewABImageFromRoot:err(2): " + err.Error())
		return nil, err
	}

	PrintVerbose("NewABImageFromRoot: found abimage.abr: " + a.Digest)
	return &a, nil
}

// WriteTo writes the json to a dest path
func (a *ABImage) WriteTo(dest string, suffix string) error {
	PrintVerbose("ABImage.WriteTo: running...")

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			PrintVerbose("ABImage.WriteTo:err: " + err.Error())
			return err
		}
	}

	if suffix != "" {
		suffix = "-" + suffix
	}
	imageName := "abimage" + suffix + ".abr"
	imagePath := fmt.Sprintf("%s/%s", dest, imageName)

	abimage, err := json.Marshal(a)
	if err != nil {
		PrintVerbose("ABImage.WriteTo:err(2): " + err.Error())
		return err
	}

	err = os.WriteFile(imagePath, abimage, 0644)
	if err != nil {
		PrintVerbose("ABImage.WriteTo:err(3): " + err.Error())
		return err
	}

	PrintVerbose("ABImage.WriteTo: successfully wrote abimage.abr to " + imagePath)

	return nil
}
