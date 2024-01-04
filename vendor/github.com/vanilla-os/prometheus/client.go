package prometheus

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description:
		Prometheus is a simple and accessible library for pulling and mounting
		container images. It is designed to be used as a dependency in ABRoot
		and Albius.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/storage"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	cstorage "github.com/containers/storage"
)

/* NewPrometheus creates a new Prometheus instance, note that currently
 * Prometheus only works with custom stores, so you need to pass the
 * root graphDriverName to create a new one.
 */
func NewPrometheus(root, graphDriverName string, maxParallelDownloads uint) (*Prometheus, error) {
	var err error

	root = filepath.Clean(root)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		err = os.MkdirAll(root, 0755)
		if err != nil {
			return nil, err
		}
	}

	runRoot := filepath.Join(root, "run")
	graphRoot := filepath.Join(root, "graph")

	store, err := cstorage.GetStore(cstorage.StoreOptions{
		RunRoot:         runRoot,
		GraphRoot:       graphRoot,
		GraphDriverName: graphDriverName,
	})
	if err != nil {
		return nil, err
	}

	return &Prometheus{
		Store: store,
		Config: PrometheusConfig{
			Root:                 root,
			GraphDriverName:      graphDriverName,
			MaxParallelDownloads: maxParallelDownloads,
		},
	}, nil
}

// PullImage pulls an image from a remote registry and stores it in the
// Prometheus store. It returns the manifest of the pulled image and an
// error if any. Note that the 'docker://' prefix is automatically added
// to the imageName to make it compatible with the alltransports.ParseImageName
// method.
func (p *Prometheus) PullImage(imageName, dstName string) (*OciManifest, error) {
	progressCh := make(chan types.ProgressProperties)
	manifestCh := make(chan OciManifest)

	defer close(progressCh)
	defer close(manifestCh)

	err := p.pullImage(imageName, dstName, progressCh, manifestCh)
	if err != nil {
		return nil, err
	}
	for {
		select {
		case report := <-progressCh:
			fmt.Printf("%s: %v/%v\n", report.Artifact.Digest.Encoded()[:12], report.Offset, report.Artifact.Size)
		case manifest := <-manifestCh:
			return &manifest, nil
		}
	}
}

// PullImageAsync does the same thing as PullImage, but returns right
// after starting the pull process. The user can track progress in the
// background by reading from the `progressCh` channel, which contains
// information about the current blob and its progress. When the pull
// process is done, the image's manifest will be sent via the `manifestCh`
// channel, which indicates the process is done.
//
// NOTE: The user is responsible for closing both channels once the operation
// completes.
func (p *Prometheus) PullImageAsync(imageName, dstName string, progressCh chan types.ProgressProperties, manifestCh chan OciManifest) error {
	err := p.pullImage(imageName, dstName, progressCh, manifestCh)
	return err
}

func (p *Prometheus) pullImage(imageName, dstName string, progressCh chan types.ProgressProperties, manifestCh chan OciManifest) error {
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", imageName))
	if err != nil {
		return err
	}

	destRef, err := storage.Transport.ParseStoreReference(p.Store, dstName)
	if err != nil {
		return err
	}

	systemCtx := &types.SystemContext{}
	policy, err := signature.DefaultPolicy(systemCtx)
	if err != nil {
		return err
	}

	policyCtx, err := signature.NewPolicyContext(policy)
	if err != nil {
		return err
	}

	duration, err := time.ParseDuration("100ms")
	if err != nil {
		return err
	}

	go func() {
		pulledManifestBytes, err := copy.Image(
			context.Background(),
			policyCtx,
			destRef,
			srcRef,
			&copy.Options{
				MaxParallelDownloads: p.Config.MaxParallelDownloads,
				ProgressInterval:     duration,
				Progress:             progressCh,
			},
		)
		if err != nil {
			return
		}

		var manifest OciManifest
		err = json.Unmarshal(pulledManifestBytes, &manifest)
		if err != nil {
			return
		}

		// here we remove the 'sha256:' prefix from the digest, so we don't have
		// to deal with it later
		manifest.Config.Digest = manifest.Config.Digest[7:]
		for i := range manifest.Layers {
			manifest.Layers[i].Digest = manifest.Layers[i].Digest[7:]
		}

		manifestCh <- manifest
	}()

	return nil
}

/* GetImageByDigest returns an image from the Prometheus store by its digest. */
func (p *Prometheus) GetImageByDigest(digest string) (cstorage.Image, error) {
	images, err := p.Store.Images()
	if err != nil {
		return cstorage.Image{}, err
	}

	for _, img := range images {
		if img.ID == digest {
			return img, nil
		}
	}

	err = cstorage.ErrImageUnknown
	return cstorage.Image{}, err
}

/* DoesImageExist checks if an image exists in the Prometheus store by its
 * digest. It returns a boolean indicating if the image exists and an error
 * if any. */
func (p *Prometheus) DoesImageExist(digest string) (bool, error) {
	image, err := p.GetImageByDigest(digest)
	if err != nil {
		return false, err
	}

	if image.ID == digest {
		return true, nil
	}

	return false, nil
}

/* MountImage mounts an image from the Prometheus store by its main layer
 * digest. It returns the mount path and an error if any. */
func (p *Prometheus) MountImage(layerId string) (string, error) {
	mountPath, err := p.Store.Mount(layerId, "")
	if err != nil {
		return "", err
	}

	return mountPath, nil
}

/* UnMountImage unmounts an image from the Prometheus store by its main layer
 * digest. It returns a boolean indicating if the unmount was successful and
 * an error if any. */
func (p *Prometheus) UnMountImage(layerId string, force bool) (bool, error) {
	res, err := p.Store.Unmount(layerId, force)
	if err != nil {
		return res, err
	}

	return res, nil
}

/* BuildContainerFile builds a dockerfile and returns the manifest of the built
 * image and an error if any. */
func (p *Prometheus) BuildContainerFile(dockerfilePath string, imageName string) (cstorage.Image, error) {
	id, _, err := imagebuildah.BuildDockerfiles(
		context.Background(),
		p.Store,
		define.BuildOptions{
			Output: imageName,
		},
		dockerfilePath,
	)
	if err != nil {
		return cstorage.Image{}, err
	}

	image, err := p.GetImageByDigest(id)
	if err != nil {
		return cstorage.Image{}, err
	}

	return image, nil
}
