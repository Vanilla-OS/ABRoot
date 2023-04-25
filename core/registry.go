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
	"io/ioutil"
	"net/http"

	"github.com/vanilla-os/abroot/settings"
)

// Registry struct
type Registry struct {
	API string
}

// Manifest struct
type Manifest struct {
	Manifest []byte
	Digest   string
	Layers   []string
}

// NewRegistry returns a new Registry struct
func NewRegistry() *Registry {
	PrintVerbose("NewRegistry: running...")
	return &Registry{
		API: fmt.Sprintf("https://%s/%s", settings.Cnf.Registry, settings.Cnf.RegistryAPIVersion),
	}
}

// HasUpdate checks if the image/tag from the registry has a different digest
func (r *Registry) HasUpdate(digest string) bool {
	PrintVerbose("Registry.HasUpdate: Checking for updates ...")

	manifest, err := r.GetManifest()
	if err != nil {
		PrintVerbose("Registry.HasUpdate:err: %s", err)
		return false
	}

	if manifest.Digest == digest {
		PrintVerbose("Registry.HasUpdate: no update available")
		return false
	}

	PrintVerbose("Registry.HasUpdate: update available. Old digest: %s, new digest: %s", digest, manifest.Digest)
	return true
}

// GetManifest returns the manifest of the image
func (r *Registry) GetManifest() (*Manifest, error) {
	PrintVerbose("Registry.GetManifest: running...")

	manifestAPIUrl := fmt.Sprintf("%s/%s/manifests/%s", r.API, settings.Cnf.Name, settings.Cnf.Tag)
	PrintVerbose("Registry.GetManifest: call uri is %s", manifestAPIUrl)

	req, err := http.NewRequest("GET", manifestAPIUrl, nil)
	if err != nil {
		PrintVerbose("Registry.GetManifest:err: %s", err)
		return nil, err
	}

	req.Header.Set("User-Agent", "abroot")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		PrintVerbose("Registry.GetManifest:err(2): %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		PrintVerbose("Registry.GetManifest:err(3): %s", err)
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(body, &m)
	if err != nil {
		PrintVerbose("Registry.GetManifest(:err4): %s", err)
		return nil, err
	}

	digest := resp.Header.Get("Docker-Content-Digest")
	layers := m["layers"].([]interface{})
	var layerDigests []string
	for _, layer := range layers {
		layerDigests = append(layerDigests, layer.(map[string]interface{})["digest"].(string))
	}

	PrintVerbose("Registry.GetManifest: success")
	return &Manifest{
		Manifest: body,
		Digest:   digest,
		Layers:   layerDigests,
	}, nil
}
