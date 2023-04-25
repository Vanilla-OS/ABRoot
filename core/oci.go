package core

import (
	"errors"
	"os"
	"path/filepath"
)

// GenerateRootfs generates a rootfs from a image recipe file
func OciExportRootFs(buildImageName string, imageRecipe *ImageRecipe, transDir string, dest string) error {
	PrintVerbose("Podman.GenerateRootfs: running...")

	imageRecipePath := filepath.Join(transDir, "imageRecipe")

	if transDir == dest {
		err := errors.New("transDir and dest cannot be the same")
		PrintVerbose("Podman.GenerateRootfs:err: %s", err)
		return err
	}

	// cleanup dest
	err := os.RemoveAll(dest)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:err: %s", err)
		return err
	}
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:err(2): %s", err)
		return err
	}

	// cleanup transDir
	err = os.RemoveAll(transDir)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:err(3): %s", err)
		return err
	}
	err = os.MkdirAll(transDir, 0755)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:err(4): %s", err)
		return err
	}

	// write imageRecipe
	err = imageRecipe.Write(imageRecipePath)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:err(5): %s", err)
		return err
	}

	// build image
	imageBuild, err := PodBuildImage(buildImageName, imageRecipePath)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:err(6): %s", err)
		return err
	}

	// mount image
	mountDir, err := PodMountImage(imageBuild)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:err(7): %s", err)
		return err
	}

	// copy mount dir contents to dest
	err = rsyncCmd(mountDir+"/", dest, []string{}, false)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:err(8): %s", err)
		return err
	}

	// unmount image
	err = PodUnmountImage(imageBuild)
	if err != nil {
		PrintVerbose("Podman.GenerateRootfs:err(9): %s", err)
		return err
	}

	return nil
}
