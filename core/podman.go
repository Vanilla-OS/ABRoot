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

func NewPodman() *Podman {
	PrintVerbose("NewPodman: running...")
	return &Podman{
		Root: podmanStorageRoot,
	}
}

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

func (p *Podman) Pull(image string) error {
	PrintVerbose("Podman.Pull: running...")
	return p.Run([]string{"pull", image})
}

func (p *Podman) Save(image string, dest string) error {
	PrintVerbose("Podman.Save: running...")
	return p.Run([]string{"save", image, "-o", dest})
}

func (p *Podman) BuildImage(image string, containerFile string) error {
	PrintVerbose("Podman.BuildImage: running...")
	return p.Run([]string{"build", "-t", image, "-f", containerFile, "."})
}

func (p *Podman) NewContainerFile(image string, labels map[string]string, args map[string]string, content string) *ContainerFile {
	PrintVerbose("Podman.NewContainerFile: running...")
	return &ContainerFile{
		From:    image,
		Labels:  labels,
		Args:    args,
		Content: content,
	}
}

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

func (p *Podman) GenerateRootfs(image string, containerFile *ContainerFile, dest string) error {
	PrintVerbose("Podman.GenerateRootfs: running...")

	rootfs := filepath.Join(dest, "abroot_trans")

	// create rootfs dir
	err := os.MkdirAll(rootfs, 0755)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error: %s", err)
		return err
	}

	// create containerfile
	err = containerFile.Write(filepath.Join(rootfs, "Containerfile"))
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(2): %s", err)
		return err
	}

	// build image
	err = p.BuildImage(image, filepath.Join(rootfs, "Containerfile"))
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(3): %s", err)
		return err
	}

	// save image
	err = p.Save(image, filepath.Join(rootfs, "image.tar"))
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(4): %s", err)
		return err
	}

	// extract layers
	err = ExtractLayers(filepath.Join(rootfs, "image.tar"), rootfs)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(5): %s", err)
		return err
	}

	// move rootfs
	err = os.Rename(filepath.Join(rootfs, "rootfs"), filepath.Join(dest, "rootfs"))
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(6): %s", err)
		return err
	}

	// remove trans dir
	err = os.RemoveAll(rootfs)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:error(7): %s", err)
		return err
	}

	PrintVerbose("Podman.GenerateRootfs: successfully generated.")

	return nil
}

func ExtractLayers(image string, dest string) error {
	PrintVerbose("ExtractLayers: running...")

	imageExDest := filepath.Join(dest, "image")
	rootfsDest := filepath.Join(dest, "rootfs")

	// create image dir
	err := os.MkdirAll(imageExDest, 0755)
	if err != nil {
		PrintVerbose("ExtractLayers:error: %s", err)
		return err
	}

	// create layers dir
	err = os.MkdirAll(rootfsDest, 0755)
	if err != nil {
		PrintVerbose("ExtractLayers:error(2): %s", err)
		return err
	}

	// extract image
	cmd := exec.Command("tar", "-xvf", image, "-C", imageExDest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		PrintVerbose("ExtractLayers:error(3): %s", err)
		return err
	}

	// read manifest
	m, err := os.Open(filepath.Join(imageExDest, "manifest.json"))
	if err != nil {
		PrintVerbose("ExtractLayers:error(4): %s", err)
		return err
	}
	defer m.Close()

	var manifest []PodmanManifest
	err = json.NewDecoder(m).Decode(&manifest)
	if err != nil {
		PrintVerbose("ExtractLayers:error(5): %s", err)
		return err
	}

	// extract layers
	for _, layer := range manifest[0].Layers {
		cmd := exec.Command("tar", "-xvf", filepath.Join(imageExDest, layer), "-C", rootfsDest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			PrintVerbose("ExtractLayers:error(6): %s", err)
			return err
		}
	}

	PrintVerbose("ExtractLayers: successfully extracted.")
	return nil
}
