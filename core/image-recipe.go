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
	"fmt"
	"os"
)

// An ImageRecipe represents a Dockerfile/Containerfile-like recipe
type ImageRecipe struct {
	From    string
	Labels  map[string]string
	Args    map[string]string
	Content string
}

// NewImageRecipe creates a new ImageRecipe instance and returns a pointer to it
func NewImageRecipe(image string, labels map[string]string, args map[string]string, content string) *ImageRecipe {
	PrintVerboseInfo("NewImageRecipe", "running...")

	return &ImageRecipe{
		From:    image,
		Labels:  labels,
		Args:    args,
		Content: content,
	}
}

// Write writes a ImageRecipe to the given path, returning an error if any
func (c *ImageRecipe) Write(path string) error {
	PrintVerboseInfo("ImageRecipe.Write", "running...")

	// create file
	file, err := os.Create(path)
	if err != nil {
		PrintVerboseErr("ImageRecipe.Write", 0, err)
		return err
	}
	defer file.Close()

	// write from
	_, err = file.WriteString(fmt.Sprintf("FROM %s\n", c.From))
	if err != nil {
		PrintVerboseErr("ImageRecipe.Write", 1, err)
		return err
	}

	// write labels
	for key, value := range c.Labels {
		_, err = file.WriteString(fmt.Sprintf("LABEL %s=%s\n", key, value))
		if err != nil {
			PrintVerboseErr("ImageRecipe.Write", 2, err)
			return err
		}
	}

	// write args
	for key, value := range c.Args {
		_, err = file.WriteString(fmt.Sprintf("ARG %s=%s\n", key, value))
		if err != nil {
			PrintVerboseErr("ImageRecipe.Write", 3, err)
			return err
		}
	}

	// write content
	_, err = file.WriteString(c.Content)
	if err != nil {
		PrintVerboseErr("ImageRecipe.Write", 4, err)
		return err
	}

	PrintVerboseInfo("ImageRecipe.Write", "done")
	return nil
}
