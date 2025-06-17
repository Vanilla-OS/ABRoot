package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/containers/buildah"
	"github.com/containers/image/v5/types"
	"github.com/containers/storage"
	humanize "github.com/dustin/go-humanize"
	"github.com/pterm/pterm"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/prometheus"
)

type NotEnoughSpaceError struct{}

func (v *NotEnoughSpaceError) Error() string {
	return "not enough space in disk"
}

var Progressbar = pterm.ProgressbarPrinter{
	Total:                     100,
	BarCharacter:              "■",
	LastCharacter:             "■",
	ElapsedTimeRoundingFactor: time.Second,
	BarStyle:                  &pterm.Style{pterm.Bold, pterm.FgDefault},
	TitleStyle:                &pterm.Style{pterm.FgDefault},
	ShowTitle:                 true,
	ShowCount:                 false,
	ShowPercentage:            true,
	ShowElapsedTime:           false,
	BarFiller:                 " ",
	MaxWidth:                  60,
	Writer:                    os.Stdout,
}

func padString(str string, size int) string {
	if len(str) < size {
		return str + strings.Repeat(" ", size-len(str))
	} else {
		return str
	}
}

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
	err = os.MkdirAll(dest, 0o755)
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
	err = os.MkdirAll(transDir, 0o755)
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

	pulledImage := false
	// pull image
	if !strings.HasPrefix(imageRecipe.From, "localhost/") {
		err = pullImageWithProgressbar(pt, buildImageName, imageRecipe)
		if err != nil {
			PrintVerboseErr("OciExportRootFs", 6.1, err)
			return err
		}
		pulledImage = true
	}

	// build image
	imageBuild, err := pt.BuildContainerFile(imageRecipePath, buildImageName)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 7, err)
		return err
	}

	if pulledImage {
		// This is safe because BuildContainerFile layers on top of the base image
		// So this won't delete the actual layers, only the image reference
		_, err = pt.Store.DeleteImage(imageRecipe.From, true)
		if err != nil {
			PrintVerboseWarn("OciExportRootFs", 7.5, "could not delete downloaded image", err)
		}
	}

	// mount image
	mountDir, err := pt.MountImage(imageBuild.TopLayer)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 8, err)
		return err
	}

	err = checkImageSize(mountDir, dest)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 8.5, err)
		return err
	}

	// copy mount dir contents to dest
	err = rsyncCmd(mountDir+"/", dest, []string{"--delete", "--checksum"}, false)
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

// returns nil if there's enough space in the filesystem for the image
// returns NotEnoughSpaceError if there is not enough space
// returns other error if the sizes were not calculated correctly
func checkImageSize(imageDir string, filesystemMount string) error {
	imageDirStat, err := os.Stat(imageDir)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 8.1, err)
		return err
	}

	var imageDirSize int64
	if imageDirStat.IsDir() {
		imageDirSize, err = getDirSize(imageDir)
		if err != nil {
			PrintVerboseErr("OciExportRootFs", 8.2, err)
			return err
		}
	} else {
		imageDirSize = imageDirStat.Size()
	}

	var stat syscall.Statfs_t
	err = syscall.Statfs(filesystemMount, &stat)
	if err != nil {
		PrintVerboseErr("OciExportRootFs", 8.3, err)
		return err
	}

	availableSpace := stat.Blocks * uint64(stat.Bsize)
	if settings.Cnf.ThinProvisioning {
		availableSpace /= 2
	}

	if uint64(imageDirSize) > availableSpace {
		err := &NotEnoughSpaceError{}
		PrintVerboseErr("OciExportRootFs", 8.4, err)
		return err
	}

	return nil
}

// pullImageWithProgressbar pulls the image specified in the provided recipe
// and reports the download progress using pterm progressbars. Each blob has
// its own bar, similar to how docker and podman report downloads in their
// respective CLIs
func pullImageWithProgressbar(pt *prometheus.Prometheus, name string, image *ImageRecipe) error {
	PrintVerboseInfo("pullImageWithProgressbar", "running...")

	progressCh := make(chan types.ProgressProperties)
	manifestCh := make(chan prometheus.OciManifest)

	defer close(progressCh)
	defer close(manifestCh)

	err := pt.PullImageAsync(image.From, name, progressCh, manifestCh)
	if err != nil {
		PrintVerboseErr("pullImageWithProgressbar", 0, err)
		return err
	}

	multi := pterm.DefaultMultiPrinter
	bars := map[string]*pterm.ProgressbarPrinter{}

	multi.Start()

	barFmt := "%s [%s/%s]"
	for {
		select {
		case report := <-progressCh:
			digest := report.Artifact.Digest.Encoded()
			if pb, ok := bars[digest]; ok {
				progressBytes := humanize.Bytes(uint64(report.Offset))
				totalBytes := humanize.Bytes(uint64(report.Artifact.Size))

				pb.Add(int(report.Offset) - pb.Current)

				title := fmt.Sprintf(barFmt, digest[:12], progressBytes, totalBytes)
				pb.UpdateTitle(padString(title, 28))
			} else {
				totalBytes := humanize.Bytes(uint64(report.Artifact.Size))

				title := fmt.Sprintf(barFmt, digest[:12], "0", totalBytes)
				newPb, err := Progressbar.WithTotal(int(report.Artifact.Size)).WithWriter(multi.NewWriter()).Start(padString(title, 28))
				if err != nil {
					PrintVerboseErr("pullImageWithProgressbar", 1, err)
					return err
				}

				bars[digest] = newPb
			}
		case <-manifestCh:
			multi.Stop()
			return nil
		}
	}
}

// FindImageWithLabel returns the name of the first image containinig the provided key-value pair
// or an empty string if none was found
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

// DeleteAllButLatestImage deletes all images
func DeleteAllButLatestImage() error {
	pt, err := prometheus.NewPrometheus(
		"/var/lib/abroot/storage",
		"overlay",
		settings.Cnf.MaxParallelDownloads,
	)
	if err != nil {
		PrintVerboseErr("DeleteAllImagesButLatest", 1, err)
		return err
	}

	allImages, err := pt.Store.Images()
	if err != nil {
		PrintVerboseErr("DeleteAllImagesButLatest", 2, err)
		return fmt.Errorf("could not retrieve all images: %w", err)
	}

	if len(allImages) == 0 {
		return nil
	}

	var latestImage *storage.Image
	for _, image := range allImages {
		if latestImage == nil || image.Created.After(latestImage.Created) {
			latestImage = &image
		}
	}

	for _, image := range allImages {
		if image.ID != latestImage.ID {
			_, err := pt.Store.DeleteImage(image.ID, true)
			if err != nil {
				PrintVerboseErr("DeleteAllImagesButLatest", 3, "failed to remove image: ", err)
			}
		}
	}

	return nil
}
