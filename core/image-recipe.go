package core

import (
	"fmt"
	"os"
)

type ImageRecipe struct {
	From    string
	Labels  map[string]string
	Args    map[string]string
	Content string
}

// NewImageRecipe creates a new ImageRecipe struct
func NewImageRecipe(image string, labels map[string]string, args map[string]string, content string) *ImageRecipe {
	PrintVerbose("Podman.NewImageRecipe: running...")

	return &ImageRecipe{
		From:    image,
		Labels:  labels,
		Args:    args,
		Content: content,
	}
}

// Write writes a ImageRecipe to a path
func (c *ImageRecipe) Write(path string) error {
	PrintVerbose("ImageRecipe.Write: running...")

	// create file
	file, err := os.Create(path)
	if err != nil {
		PrintVerbose("ImageRecipe.Write:err: %s", err)
		return err
	}
	defer file.Close()

	// write from
	_, err = file.WriteString(fmt.Sprintf("FROM %s\n", c.From))
	if err != nil {
		PrintVerbose("ImageRecipe.Write:err(2): %s", err)
		return err
	}

	// write labels
	for key, value := range c.Labels {
		_, err = file.WriteString(fmt.Sprintf("LABEL %s=%s\n", key, value))
		if err != nil {
			PrintVerbose("ImageRecipe.Write:err(3): %s", err)
			return err
		}
	}

	// write args
	for key, value := range c.Args {
		_, err = file.WriteString(fmt.Sprintf("ARG %s=%s\n", key, value))
		if err != nil {
			PrintVerbose("ImageRecipe.Write:err(4): %s", err)
			return err
		}
	}

	// write content
	_, err = file.WriteString(c.Content)
	if err != nil {
		PrintVerbose("ImageRecipe.Write:err(5): %s", err)
		return err
	}

	PrintVerbose("ImageRecipe.Write: successfully wrote.")
	return nil
}
