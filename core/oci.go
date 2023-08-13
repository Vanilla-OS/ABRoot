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
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/containers/buildah"
	cstypes "github.com/containers/storage/types"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/prometheus"
)

// OciExportRootFs generates a rootfs from a image recipe file
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

// FindImageWithLabel returns the name of the first image containinig the provided key-value pair
// or an empty string if none was found
func FindImageWithLabel(key, value string) (string, error) {
	PrintVerbose("FindImageWithLabel: running...")

	pt, err := prometheus.NewPrometheus(
		"/var/lib/abroot/storage",
		"overlay",
		settings.Cnf.MaxParallelDownloads,
	)
	if err != nil {
		PrintVerbose("FindImageWithLabel:err: %s", err)
		return "", err
	}

	images, err := pt.Store.Images()
	if err != nil {
		PrintVerbose("FindImageWithLabel:err(2): %s", err)
		return "", err
	}

	for _, img := range images {
		// This is the only way I could find to get the labels form an image
		builder, err := buildah.ImportBuilderFromImage(context.Background(), pt.Store, buildah.ImportFromImageOptions{Image: img.ID})
		if err != nil {
			PrintVerbose("FindImageWithLabel:err(3): %s", err)
			return "", err
		}

		val, ok := builder.Labels()[key]
		if ok && val == value {
			return img.Names[0], nil
		}
	}

	return "", nil
}

// RetrieveImageForRoot retrieves the image created for the provided root ("vos-a"|"vos-b")
func RetrieveImageForRoot(root string) (string, error) {
	PrintVerbose("ApplyInImageForRoot: running...")

	image, err := FindImageWithLabel("ABRoot.root", root)
	if err != nil {
		PrintVerbose("ApplyInImageForRoot:err: %s", err)
		return "", err
	}

	return image, nil
}

// DeleteImageForRoot deletes the image created for the provided root ("vos-a"|"vos-b")
func DeleteImageForRoot(root string) error {
	image, err := RetrieveImageForRoot(root)
	if err != nil {
		PrintVerbose("DeleteImageForRoot:err: %s", err)
		return err
	}

	pt, err := prometheus.NewPrometheus(
		"/var/lib/abroot/storage",
		"overlay",
		settings.Cnf.MaxParallelDownloads,
	)
	if err != nil {
		PrintVerbose("DeleteImageForRoot:err(2): %s", err)
		return err
	}

	_, err = pt.Store.DeleteImage(image, true)
	if err != nil && err != cstypes.ErrNotAnImage {
		PrintVerbose("DeleteImageForRoot:err(3): %s", err)
		return err
	}

	return nil
}
