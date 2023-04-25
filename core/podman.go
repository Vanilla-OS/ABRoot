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

// README:
// This file will migrate to a more robust solution, maybe when libpod
// becomes more stable or podman bindings become backwards compatible
// with older versions of podman server.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	abrootStorage = "/var/lib/abroot/storage"
)

type PodmanManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

type PodmanImage struct {
	Digest string
	Image  string
}

type PodInspection struct {
	Id     string    `json:"Id"`
	Digest string    `json:"Digest"`
	Size   int64     `json:"Size"`
	RootFS PodRootFS `json:"RootFS"`
}

type PodRootFS struct {
	Type   string   `json:"Type"`
	Layers []string `json:"Layers"`
}

// Run runs a podman command
func PodRun(args []string) error {
	PrintVerbose("Podman.Run: running %s", strings.Join(args, " "))

	// add root flag to args
	args = append([]string{"--root", abrootStorage}, args...)

	// run podman
	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		PrintVerbose("Podman.Run:err: %s", err)
		return err
	}

	PrintVerbose("Podman.Run: successfully ran.")
	return nil
}

// RunOutput runs a podman command and returns the output
func PodRunOutput(args []string) (string, error) {
	PrintVerbose("Podman.RunOutput: running: %s", strings.Join(args, " "))

	args = append([]string{"--root", abrootStorage}, args...)
	cmd := exec.Command("podman", args...)
	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("Podman.RunOutput:err: %s", err)
		return "", err
	}

	PrintVerbose("Podman.RunOutput: successfully ran.")
	return string(out), nil
}

// Pull pulls an image and returns a PodmanImage struct
func PodPull(image string) (*PodmanImage, error) {
	PrintVerbose("Podman.Pull: running...")

	err := PodRun([]string{"pull", image})
	if err != nil {
		PrintVerbose("Podman.Pull:err: %s", err)
		return nil, err
	}

	digest, err := PodInspect(image, "Digest")
	if err != nil {
		PrintVerbose("Podman.Pull:err(2): %s", err)
		return nil, err
	}

	return &PodmanImage{
		Digest: digest,
		Image:  image,
	}, nil
}

// Inspect returns a value from an image inspection
func PodInspect(image string, key string) (string, error) {
	PrintVerbose("Podman.Inspect: running...")

	out, err := PodRunOutput([]string{"inspect", image, "--format", fmt.Sprintf("{{.%s}}", key)})
	if err != nil {
		PrintVerbose("Podman.Inspect:err: %s", err)
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// GetInspection returns a PodInspection struct from an image
func PodGetInspection(image string) (*PodInspection, error) {
	out, err := PodRunOutput([]string{"inspect", image})
	if err != nil {
		PrintVerbose("Podman.GetInspection:err: %s", err)
		return nil, err
	}

	var inspections []PodInspection
	err = json.Unmarshal([]byte(out), &inspections)
	if err != nil {
		PrintVerbose("Podman.GetInspection:err: %s", err)
		return nil, err
	}

	if len(inspections) == 0 {
		return nil, errors.New("no inspections found")
	}

	targetInspection := &inspections[0]
	for i, layer := range targetInspection.RootFS.Layers {
		targetInspection.RootFS.Layers[i] = strings.Replace(layer, "sha256:", "", 1)
	}

	return targetInspection, nil
}

// Save saves an image to a destination
func PodSave(image string, dest string) error {
	PrintVerbose("Podman.Save: running...")
	return PodRun([]string{"save", image, "-o", dest})
}

// BuildImage builds an image from a container file
func PodBuildImage(buildImageName string, imageRecipe string) (string, error) {
	PrintVerbose("Podman.BuildImage: running...")
	return buildImageName, PodRun([]string{
		"build",
		"--layers=false",
		"--no-cache",
		"-t", buildImageName,
		"-f", imageRecipe, ".",
	})
}

// Create creates a container
func PodCreate(image string, name string, start bool) error {
	PrintVerbose("Podman.Create: running...")

	err := PodRun([]string{"create", "--name", name, image})
	if err != nil {
		PrintVerbose("Podman.Create:err: %s", err)
		return err
	}

	if start {
		err = PodRun([]string{"start", name})
		if err != nil {
			PrintVerbose("Podman.Create:err(2): %s", err)
			return err
		}
	}

	PrintVerbose("Podman.Create: successfully created.")
	return nil
}

// Start starts a container
func PodStart(name string) error {
	PrintVerbose("Podman.Start: running...")
	return PodRun([]string{"start", name})
}

// Remove removes a container
func PodRemove(name string, force bool) error {
	PrintVerbose("Podman.Remove: running...")

	forceFlag := ""
	if force {
		forceFlag = "-f"
	}

	return PodRun([]string{"rm", forceFlag, name})
}

// RemoveImage removes an image (and container if force is true)
func PodRemoveImage(name string, force bool) error {
	PrintVerbose("Podman.RemoveImage: running...")
	forceFlag := ""
	if force {
		forceFlag = "-f"
	}

	return PodRun([]string{"rmi", forceFlag, name})
}

// Export exports a container
func PodExport(name string, dest string) error {
	PrintVerbose("Podman.Export: running...")
	return PodRun([]string{"export", name, "-o", dest})
}

// MountImage mounts an image
func PodMountImage(image string) (string, error) {
	PrintVerbose("Podman.MountImage: running...")
	out, err := PodRunOutput([]string{"image", "mount", image})
	if err != nil {
		PrintVerbose("Podman.MountImage:err: %s", err)
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// UnmountImage unmounts an image
func PodUnmountImage(image string) error {
	PrintVerbose("Podman.UnmountImage: running...")
	return PodRun([]string{"image", "umount", image})
}
