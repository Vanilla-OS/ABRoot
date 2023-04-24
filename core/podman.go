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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	podmanStorageRoot = "/var/lib/abroot/storage"
)

type Podman struct {
	Root string
}

type ContainerFile struct {
	From    string
	Labels  map[string]string
	Args    map[string]string
	Content string
}

type PodmanManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

type PodmanImage struct {
	Digest string
	Image  string
}

func NewPodman() (*Podman, error) {
	PrintVerbose("NewPodman: running...")

	return &Podman{
		Root: podmanStorageRoot,
	}, nil
}

// Run runs a podman command
func (p *Podman) Run(args []string) error {
	PrintVerbose("Podman.Run: running %s", strings.Join(args, " "))

	// add root flag to args
	args = append([]string{"--root", p.Root}, args...)

	// run podman
	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		PrintVerbose("Podman.Run:error: %s", err)
		return err
	}

	PrintVerbose("Podman.Run: successfully ran.")
	return nil
}

// RunOutput runs a podman command and returns the output
func (p *Podman) RunOutput(args []string) (string, error) {
	PrintVerbose("Podman.RunOutput: running %s", strings.Join(args, " "))

	// add root flag to args
	args = append([]string{"--root", p.Root}, args...)

	// run podman
	cmd := exec.Command("podman", args...)
	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("Podman.RunOutput:error: %s", err)
		return "", err
	}

	PrintVerbose("Podman.RunOutput: successfully ran.")
	return string(out), nil
}

// Pull pulls an image and returns a PodmanImage struct
func (p *Podman) Pull(image string) (*PodmanImage, error) {
	PrintVerbose("Podman.Pull: running...")

	// pull image
	err := p.Run([]string{"pull", image})
	if err != nil {
		PrintVerbose("Podman.Pull:error: %s", err)
		return nil, err
	}

	// get digest
	digest, err := p.Inspect(image, "Digest")
	if err != nil {
		PrintVerbose("Podman.Pull:error(2): %s", err)
		return nil, err
	}

	return &PodmanImage{
		Digest: digest,
		Image:  image,
	}, nil
}

// Inspect returns a value from an image
func (p *Podman) Inspect(image string, key string) (string, error) {
	PrintVerbose("Podman.Inspect: running...")

	// inspect image
	out, err := p.RunOutput([]string{"inspect", image, "--format", fmt.Sprintf("{{.%s}}", key)})
	if err != nil {
		PrintVerbose("Podman.Inspect:error: %s", err)
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// Save saves an image to a destination
func (p *Podman) Save(image string, dest string) error {
	PrintVerbose("Podman.Save: running...")
	return p.Run([]string{"save", image, "-o", dest})
}

// BuildImage builds an image from a container file
func (p *Podman) BuildImage(buildImageName string, containerFile string) (string, error) {
	PrintVerbose("Podman.BuildImage: running...")
	return buildImageName, p.Run([]string{"build", "--layers=false", "--no-cache", "-t", buildImageName, "-f", containerFile, "."})
}

// NewContainerFile creates a new ContainerFile struct
func (p *Podman) NewContainerFile(image string, labels map[string]string, args map[string]string, content string) *ContainerFile {
	PrintVerbose("Podman.NewContainerFile: running...")
	return &ContainerFile{
		From:    image,
		Labels:  labels,
		Args:    args,
		Content: content,
	}
}

// Write writes a ContainerFile to a path
func (c *ContainerFile) Write(path string) error {
	PrintVerbose("ContainerFile.Write: running...")

	// create file
	file, err := os.Create(path)
	if err != nil {
		PrintVerbose("ContainerFile.Write:error: %s", err)
		return err
	}
	defer file.Close()

	// write from
	_, err = file.WriteString(fmt.Sprintf("FROM %s\n", c.From))
	if err != nil {
		PrintVerbose("ContainerFile.Write:error(2): %s", err)
		return err
	}

	// write labels
	for key, value := range c.Labels {
		_, err = file.WriteString(fmt.Sprintf("LABEL %s=%s\n", key, value))
		if err != nil {
			PrintVerbose("ContainerFile.Write:error(3): %s", err)
			return err
		}
	}

	// write args
	for key, value := range c.Args {
		_, err = file.WriteString(fmt.Sprintf("ARG %s=%s\n", key, value))
		if err != nil {
			PrintVerbose("ContainerFile.Write:error(4): %s", err)
			return err
		}
	}

	// write content
	_, err = file.WriteString(c.Content)
	if err != nil {
		PrintVerbose("ContainerFile.Write:error(5): %s", err)
		return err
	}

	PrintVerbose("ContainerFile.Write: successfully wrote.")
	return nil
}

// Create creates a container
func (p *Podman) Create(image string, name string, start bool) error {
	PrintVerbose("Podman.Create: running...")

	err := p.Run([]string{"create", "--name", name, image})
	if err != nil {
		PrintVerbose("Podman.Create:error: %s", err)
		return err
	}

	if start {
		err = p.Run([]string{"start", name})
		if err != nil {
			PrintVerbose("Podman.Create:error(2): %s", err)
			return err
		}
	}

	PrintVerbose("Podman.Create: successfully created.")
	return nil
}

// Start starts a container
func (p *Podman) Start(name string) error {
	PrintVerbose("Podman.Start: running...")
	return p.Run([]string{"start", name})
}

// Remove removes a container
func (p *Podman) Remove(name string, force bool) error {
	PrintVerbose("Podman.Remove: running...")

	forceFlag := ""
	if force {
		forceFlag = "-f"
	}

	return p.Run([]string{"rm", forceFlag, name})
}

// RemoveImage removes an image (and container if force is true)
func (p *Podman) RemoveImage(name string, force bool) error {
	PrintVerbose("Podman.RemoveImage: running...")
	forceFlag := ""
	if force {
		forceFlag = "-f"
	}

	return p.Run([]string{"rmi", forceFlag, name})
}

// Export exports a container
func (p *Podman) Export(name string, dest string) error {
	PrintVerbose("Podman.Export: running...")
	return p.Run([]string{"export", name, "-o", dest})
}

// GenerateRootfs generates a rootfs from a container file
func (p *Podman) GenerateRootfs(buildImageName string, containerFile *ContainerFile, transDir string, dest string) error {
	PrintVerbose("Podman.GenerateRootfs: running...")

	if transDir == dest {
		err := errors.New("transDir and dest cannot be the same")
		PrintVerbose("Podman.GenerateRootfs:error: %s", err)
		return err
	}

	// cleanup dest
	err := os.RemoveAll(dest)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error: %s", err)
		return err
	}

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(3): %s", err)
		return err
	}

	// cleanup transDir
	err = os.RemoveAll(transDir)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(2): %s", err)
		return err
	}

	err = os.MkdirAll(transDir, 0755)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(4): %s", err)
		return err
	}

	// create containerfile
	err = containerFile.Write(filepath.Join(transDir, "Containerfile"))
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(4): %s", err)
		return err
	}

	// build image
	imageBuild, err := p.BuildImage(buildImageName, filepath.Join(transDir, "Containerfile"))
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(5): %s", err)
		return err
	}

	// // get image layers
	// layers := []string{}
	// inspect, err := p.GetInspection(imageBuild)
	// if err != nil {
	// 	PrintVerbose("Podman.GenerateRootfs:error: %s", err)
	// 	return err
	// }

	// layers = append(layers, inspect.RootFS.Layers...)

	// // extract layers
	// for _, layer := range layers {
	// 	err = p.ExtractLayer(layer, dest)
	// 	if err != nil {
	// 		PrintVerbose("Podman.GenerateRootfs:error: %s", err)
	// 		return err
	// 	}
	// }

	// mount image
	mountDir, err := p.MountImage(imageBuild)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(6): %s", err)
		return err
	}

	// copy mount dir contents to dest
	PrintVerbose("Podman.GenerateRootfs: copying %s to %s", mountDir+"/", dest)
	err = exec.Command("rsync", "-avxHAX", mountDir+"/", dest).Run()
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(7): %s", err)
		return err
	}

	// unmount image
	err = p.UnmountImage(imageBuild)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(8): %s", err)
		return err
	}

	return nil
}

// MountImage mounts an image
func (p *Podman) MountImage(image string) (string, error) {
	PrintVerbose("Podman.MountImage: running...")
	out, err := p.RunOutput([]string{"image", "mount", image})
	if err != nil {
		PrintVerbose("Podman.MountImage:error: %s", err)
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// UnmountImage unmounts an image
func (p *Podman) UnmountImage(image string) error {
	PrintVerbose("Podman.UnmountImage: running...")
	return p.Run([]string{"image", "umount", image})
}

type Inspection struct {
	Id     string `json:"Id"`
	Digest string `json:"Digest"`
	Size   int64  `json:"Size"`
	RootFS RootFS `json:"RootFS"`
}

type RootFS struct {
	Type   string   `json:"Type"`
	Layers []string `json:"Layers"`
}

func (p *Podman) GetInspection(image string) (*Inspection, error) {
	out, err := p.RunOutput([]string{"inspect", image})
	if err != nil {
		PrintVerbose("Podman.GetInspection:error: %s", err)
		return nil, err
	}

	var inspections []Inspection
	err = json.Unmarshal([]byte(out), &inspections)
	if err != nil {
		PrintVerbose("Podman.GetInspection:error: %s", err)
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

func (p *Podman) ExtractLayer(layer string, dest string) error {
	PrintVerbose("Podman.ExtractLayer: running...")

	layerPath := filepath.Join(podmanStorageRoot, "overlay", layer, "diff")

	cmd := exec.Command("rsync", "-avxHAX", layerPath, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		PrintVerbose("Podman.ExtractLayer:error: %s", err)
		return err
	}

	return nil
}
