package core

// NOTE: replaced by podman.go, here as reference

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
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/vanilla-os/abroot/settings"
)

// OldRegistry is the main struct for the registry
type OldRegistry struct {
	OldRegistry string
	Name        string
	Tag         string
	Image       *Image
}

// NewOldRegistry creates a new registry
func NewOldRegistry() *OldRegistry {
	r := &OldRegistry{
		OldRegistry: settings.Cnf.Registry,
		Name:        settings.Cnf.Name,
		Tag:         settings.Cnf.Tag,
		Image:       nil,
	}

	r.Image, _ = r.GetManifest()
	return r
}

type Image struct {
	Manifest []byte
	Digest   string
	Layers   []string
}

// GetManifest returns the manifest of the image
func (r *OldRegistry) GetManifest() (*Image, error) {
	fmt.Printf("Getting manifest for %s/%s:%s ...\n", r.OldRegistry, r.Name, r.Tag)

	ref := fmt.Sprintf("%s/%s:%s", r.OldRegistry, r.Name, r.Tag)
	manifest, err := crane.Manifest(ref)
	if err != nil {
		return nil, err
	}

	// convert manifest from []byte to json
	m := make(map[string]interface{})
	err = json.Unmarshal(manifest, &m)
	if err != nil {
		return nil, err
	}

	digest, err := crane.Digest(ref)
	if err != nil {
		return nil, err
	}

	layers := []string{}
	for _, layer := range m["layers"].([]interface{}) {
		layers = append(layers, layer.(map[string]interface{})["digest"].(string))
	}

	return &Image{
		Manifest: manifest,
		Digest:   digest,
		Layers:   layers,
	}, nil
}

// MakeRootfs creates a new rootfs from the image
// by extracting the layers to the destination
func (r *OldRegistry) MakeRootfs(dest string) error {
	options := []crane.Option{}
	layers := r.Image.Layers

	dest = fmt.Sprintf("%s/%s", dest, r.Image.Digest)

	// Create destination directory
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return fmt.Errorf("creating directory %s: %w", dest, err)
	}

	// Download layers
	for n, layer := range layers {
		src := fmt.Sprintf("%s/%s@%s", r.OldRegistry, r.Name, layer)

		layer, err := crane.PullLayer(src, options...)
		if err != nil {
			return fmt.Errorf("pulling layer %s: %w", src, err)
		}

		blob, err := layer.Uncompressed()
		if err != nil {
			return fmt.Errorf("fetching blob %s: %w", src, err)
		}

		tarPath := filepath.Join(dest, fmt.Sprintf("%d.tar", n))
		blobTar, err := os.Create(tarPath)
		if err != nil {
			return fmt.Errorf("creating tar %s: %w", tarPath, err)
		}

		_, err = io.Copy(blobTar, blob)
		if err != nil {
			return fmt.Errorf("writing tar %s: %w", tarPath, err)
		}

		err = blobTar.Close()
		if err != nil {
			return fmt.Errorf("closing tar %s: %w", tarPath, err)
		}

		err = blob.Close()
		if err != nil {
			return fmt.Errorf("closing blob %s: %w", src, err)
		}

		fmt.Printf("Fetched layer: %s\n", layer)
	}

	// Extract layers in dest merging them
	for n := len(layers) - 1; n >= 0; n-- {
		tarPath := filepath.Join(dest, fmt.Sprintf("%d.tar", n))
		err := extractTar(tarPath, dest)
		if err != nil {
			return fmt.Errorf("extracting tar %s: %w", tarPath, err)
		}

		fmt.Printf("Extracted layer: %d\n", n)
	}

	return nil
}

// extractTar extracts a tar file to the destination
func extractTar(src, dest string) error {
	cmd := exec.Command("tar", "-xvf", src, "-C", dest)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("extracting tar %s: %w", src, err)
	}

	return nil
}
