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

// OciExportRootFs generates a rootfs from an image recipe file
func OciExportRootFs(buildImageName string, imageRecipe *ImageRecipe, transDir string, dest string) error {
	PrintVerboseInfo("OciExportRootFs", "running...")

	pt, err := prometheus.NewPrometheus(
		"/var/lib/abroot/storage",
		"overlay",
		settings.Cnf.MaxParallelDownloads,
	)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 0, err)
		return err
	}

	imageRecipePath := filepath.Join(transDir, "imageRecipe")

	if transDir == dest {
		err := errors.New("transDir and dest cannot be the same")
		PrintVerboseErr("OciExportRootFs", 1, err)
		return err
	}

	// create dest if it doesn't exist
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 3, err)
		return err
	}

	// cleanup transDir
	err = os.RemoveAll(transDir)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 4, err)
		return err
	}
	err = os.MkdirAll(transDir, 0755)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 5, err)
		return err
	}

	// write imageRecipe
	err = imageRecipe.Write(imageRecipePath)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 6, err)
		return err
	}

	// build image
	imageBuild, err := pt.BuildContainerFile(imageRecipePath, buildImageName)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 7, err)
		return err
	}

	// mount image
	mountDir, err := pt.MountImage(imageBuild.TopLayer)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 8, err)
		return err
	}

	// copy mount dir contents to dest
	err = rsyncCmd(mountDir+"/", dest, []string{"--delete"}, false)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 9, err)
		return err
	}

	// unmount image
	_, err = pt.UnMountImage(imageBuild.TopLayer, true)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 10, err)
		return err
	}

	return nil
}

// FindImageWithLabel returns the name of the first image containing the
// provided key-value pair or an empty string if none was found
func FindImageWithLabel(key, value string) (string, error) {
	PrintVerboseInfo("FindImageWithLabel", "running...")

	pt, err := prometheus.NewPrometheus(
		"/var/lib/abroot/storage",
		"overlay",
		settings.Cnf.MaxParallelDownloads,
	)
	if err != nil {
		PrintVerboseErr("FindImageWithLabel", 0, err)
		return "", err
	}

	images, err := pt.Store.Images()
	if err != nil {
		PrintVerboseErr("FindImageWithLabel", 1, err)
		return "", err
	}

	for _, img := range images {
		// This is the only way I could find to get the labels form an image
		builder, err := buildah.ImportBuilderFromImage(context.Background(), pt.Store, buildah.ImportFromImageOptions{Image: img.ID})
		if err != nil {
			PrintVerboseErr("FindImageWithLabel", 2, err)
			return "", err
		}

		val, ok := builder.Labels()[key]
		if ok && val == value {
			return img.Names[0], nil
		}
	}

	return "", nil
}

// RetrieveImageForRoot retrieves the image created for the provided root
// based on the label. Note for distro maintainers: labels must follow those
// defined in the ABRoot config file
func RetrieveImageForRoot(root string) (string, error) {
	PrintVerboseInfo("RetrieveImageForRoot", "running...")

	image, err := FindImageWithLabel("ABRoot.root", root)
	if err != nil {
		PrintVerboseErr("RetrieveImageForRoot", 0, err)
		return "", err
	}

	return image, nil
}

// DeleteImageForRoot deletes the image created for the provided root
func DeleteImageForRoot(root string) error {
	image, err := RetrieveImageForRoot(root)
	if err != nil {
		PrintVerboseErr("DeleteImageForRoot", 0, err)
		return err
	}

	pt, err := prometheus.NewPrometheus(
		"/var/lib/abroot/storage",
		"overlay",
		settings.Cnf.MaxParallelDownloads,
	)
	if err != nil {
		PrintVerboseErr("DeleteImageForRoot", 1, err)
		return err
	}

	_, err = pt.Store.DeleteImage(image, true)
	if err != nil && err != cstypes.ErrNotAnImage {
		PrintVerboseErr("DeleteImageForRoot", 2, err)
		return err
	}

	return nil
}
