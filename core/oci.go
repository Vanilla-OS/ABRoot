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
	"errors"
	"os"
	"path/filepath"

	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/prometheus"
)

// GenerateRootfs generates a rootfs from a image recipe file
func OciExportRootFs(buildImageName string, imageRecipe *ImageRecipe, transDir string, dest string) error {
	PrintVerbose("OciExportRootFs: running...")

	pt, err := prometheus.NewPrometheus(
		"/var/lib/abroot/storage",
		"overlay",
		settings.Cnf.MaxParallelDownloads,
	)
	if err != nil {
		PrintVerbose("OciExportRootFs:err: %s", err)
		return err
	}

	imageRecipePath := filepath.Join(transDir, "imageRecipe")

	if transDir == dest {
		err := errors.New("transDir and dest cannot be the same")
		PrintVerbose("OciExportRootFs:err(2): %s", err)
		return err
	}

	// cleanup dest
	err = os.RemoveAll(dest)
	if err != nil {
		PrintVerbose("OciExportRootFs:err(3): %s", err)
		return err
	}
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		PrintVerbose("OciExportRootFs:err(4): %s", err)
		return err
	}

	// cleanup transDir
	err = os.RemoveAll(transDir)
	if err != nil {
		PrintVerbose("OciExportRootFs:err(5): %s", err)
		return err
	}
	err = os.MkdirAll(transDir, 0755)
	if err != nil {
		PrintVerbose("OciExportRootFs:err(6): %s", err)
		return err
	}

	// write imageRecipe
	err = imageRecipe.Write(imageRecipePath)
	if err != nil {
		PrintVerbose("OciExportRootFs:err(7): %s", err)
		return err
	}

	// build image
	imageBuild, err := pt.BuildContainerFile(imageRecipePath, buildImageName)
	if err != nil {
		PrintVerbose("OciExportRootFs:err(8): %s", err)
		return err
	}

	// mount image
	mountDir, err := pt.MountImage(imageBuild.TopLayer)
	if err != nil {
		PrintVerbose("OciExportRootFs:err(9): %s", err)
		return err
	}

	// copy mount dir contents to dest
	err = rsyncCmd(mountDir+"/", dest, []string{}, false)
	if err != nil {
		PrintVerbose("OciExportRootFs:err(10): %s", err)
		return err
	}

	// unmount image
	_, err = pt.UnMountImage(imageBuild.TopLayer, true)
	if err != nil {
		PrintVerbose("OciExportRootFs:err(11): %s", err)
		return err
	}

	return nil
}
