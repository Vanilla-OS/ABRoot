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

type Manifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

func NewPodman() *Podman {
	return &Podman{
		Root: podmanStorageRoot,
	}
}

func (p *Podman) Run(args []string) error {
	// add root flag to args
	args = append([]string{"--root", p.Root}, args...)

	// run podman
	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (p *Podman) Pull(image string) error {
	return p.Run([]string{"pull", image})
}

func (p *Podman) Save(image string, dest string) error {
	return p.Run([]string{"save", image, "-o", dest})
}

func (p *Podman) BuildImage(image string, containerFile string) error {
	return p.Run([]string{"build", "-t", image, "-f", containerFile, "."})
}

func (p *Podman) NewContainerFile(image string, labels map[string]string, args map[string]string, content string) *ContainerFile {
	return &ContainerFile{
		From:    image,
		Labels:  labels,
		Args:    args,
		Content: content,
	}
}

func (c *ContainerFile) Write(path string) error {
	// create file
	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	// write from
	_, err = file.WriteString(fmt.Sprintf("FROM %s\n", c.From))
	if err != nil {
		fmt.Println(err)
		return err
	}

	// write labels
	for key, value := range c.Labels {
		_, err = file.WriteString(fmt.Sprintf("LABEL %s=%s\n", key, value))
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// write args
	for key, value := range c.Args {
		_, err = file.WriteString(fmt.Sprintf("ARG %s=%s\n", key, value))
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// write content
	_, err = file.WriteString(c.Content)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (p *Podman) GenerateRootfs(image string, containerFile *ContainerFile, dest string) error {
	rootfs := filepath.Join(dest, "abroot_trans")

	// create rootfs dir
	err := os.MkdirAll(rootfs, 0755)
	if err != nil {
		fmt.Println("create_rootfs_dir:", err)
		return err
	}

	// create containerfile
	err = containerFile.Write(filepath.Join(rootfs, "Containerfile"))
	if err != nil {
		fmt.Println("create_containerfile:", err)
		return err
	}

	// build image
	err = p.BuildImage(image, filepath.Join(rootfs, "Containerfile"))
	if err != nil {
		fmt.Println("build_image:", err)
		return err
	}

	// save image
	err = p.Save(image, filepath.Join(rootfs, "image.tar"))
	if err != nil {
		fmt.Println("save_image:", err)
		return err
	}

	// extract layers
	err = ExtractLayers(filepath.Join(rootfs, "image.tar"), rootfs)
	if err != nil {
		fmt.Println("extract_layers:", err)
		return err
	}

	// move rootfs
	err = os.Rename(filepath.Join(rootfs, "rootfs"), filepath.Join(dest, "rootfs"))
	if err != nil {
		fmt.Println("move_rootfs:", err)
		return err
	}

	// remove trans dir
	err = os.RemoveAll(rootfs)
	if err != nil {
		fmt.Println("remove_trans_dir:", err)
		return err
	}

	return nil
}

func ExtractLayers(image string, dest string) error {
	imageExDest := filepath.Join(dest, "image")
	rootfsDest := filepath.Join(dest, "rootfs")

	// create image dir
	err := os.MkdirAll(imageExDest, 0755)
	if err != nil {
		fmt.Println("create_image_dir:", err)
		return err
	}

	// create layers dir
	err = os.MkdirAll(rootfsDest, 0755)
	if err != nil {
		fmt.Println("create_rootfs_dir:", err)
		return err
	}

	// extract image
	cmd := exec.Command("tar", "-xvf", image, "-C", imageExDest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("extract_image:", err)
		return err
	}

	// read manifest
	m, err := os.Open(filepath.Join(imageExDest, "manifest.json"))
	if err != nil {
		fmt.Println("open_manifest:", err)
		return err
	}
	defer m.Close()

	var manifest []Manifest
	err = json.NewDecoder(m).Decode(&manifest)
	if err != nil {
		fmt.Println("decode_manifest:", err)
		return err
	}

	// extract layers
	for _, layer := range manifest[0].Layers {
		cmd := exec.Command("tar", "-xvf", filepath.Join(imageExDest, layer), "-C", rootfsDest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("extract_layer:", err)
			return err
		}
	}

	return nil
}
