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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"runtime"

	"github.com/vanilla-os/abroot/settings"
)

// A Registry instance exposes functions to interact with the configured
// Docker registry
type Registry struct {
	API string
}

// Manifest is the struct used to parse the manifest response from the registry
// it contains the manifest itself, the digest and the list of layers. This
// should be compatible with most registries, but it's not guaranteed
type Manifest struct {
	Manifest map[string]interface{}
	Digest   string
	Layers   []string
}

var ErrImageNotFound error = errors.New("configured image cannot be found")

// NewRegistry returns a new Registry instance, exposing functions to
// interact with the configured Docker registry
func NewRegistry() *Registry {
	PrintVerboseInfo("NewRegistry", "running...")
	return &Registry{
		API: fmt.Sprintf("https://%s/%s", settings.Cnf.Registry, settings.Cnf.RegistryAPIVersion),
	}
}

// HasUpdate checks if the image/tag from the registry has a different digest
// it returns the new digest and a boolean indicating if an update is available
func (r *Registry) HasUpdate(digest string) (string, bool, error) {
	PrintVerboseInfo("Registry.HasUpdate", "Checking for updates ...")

	token, err := GetToken()
	if err != nil {
		PrintVerboseErr("Registry.HasUpdate", 0, err)
		return "", false, err
	}

	manifest, err := r.GetManifest(token)
	if err != nil {
		PrintVerboseErr("Registry.HasUpdate", 1, err)
		return "", false, err
	}

	if manifest.Digest == digest {
		PrintVerboseInfo("Registry.HasUpdate", "no update available")
		return "", false, nil
	}

	PrintVerboseInfo("Registry.HasUpdate", "update available. Old digest: ", digest, ", new digest: ", manifest.Digest)
	return manifest.Digest, true, nil
}

func getRegistryAuthUrl() (string, string, error) {
	requestUrl := fmt.Sprintf(
		"https://%s/%s/",
		settings.Cnf.Registry,
		settings.Cnf.RegistryAPIVersion,
	)

	resp, err := http.Get(requestUrl)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode == 401 {
		authUrl := resp.Header["www-authenticate"]
		if len(authUrl) == 0 {
			authUrl = resp.Header["Www-Authenticate"]
			if len(authUrl) == 0 {
				return "", "", fmt.Errorf("unable to find authentication url for registry")
			}
		}
		return strings.Split(strings.Split(authUrl[0], "realm=\"")[1], "\",")[0], strings.Split(strings.Split(authUrl[0], "service=\"")[1], "\"")[0], nil
	} else {
		PrintVerboseInfo("Registry.getRegistryAuthUrl", "registry does not require authentication")
		return fmt.Sprintf("https://%s/", settings.Cnf.Registry), settings.Cnf.RegistryService, nil
	}
}

// GetToken generates a token using the provided tokenURL and returns it
func GetToken() (string, error) {
	authUrl, serviceUrl, err := getRegistryAuthUrl()
	if err != nil {
		return "", err
	}
	requestUrl := fmt.Sprintf(
		"%s?scope=repository:%s:pull&service=%s",
		authUrl,
		settings.Cnf.Name,
		serviceUrl,
	)
	PrintVerboseInfo("Registry.GetToken", "call URI is ", requestUrl)

	resp, err := http.Get(requestUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return "", ErrImageNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status code: %d", resp.StatusCode)
	}

	tokenBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse token from response
	var tokenResponse struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(tokenBytes, &tokenResponse)
	if err != nil {
		return "", err
	}

	token := tokenResponse.Token
	return token, nil
}

// getManifestJSON makes an HTTP request to fetch a manifest JSON from the given API URL
// a token is required to perform the request and is generated using GetToken()
func (r *Registry) getManifestJSON(url string, token string) (map[string]interface{}, string, error) {
	PrintVerboseInfo("Registry.GetManifestJSON", "call URI is ", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		PrintVerboseErr("Registry.GetManifestJSON", 0, err)
		return nil, "", err
	}

	req.Header.Set("User-Agent", "abroot")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		PrintVerboseErr("Registry.GetManifestJSON", 1, err)
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		PrintVerboseErr("Registry.GetManifestJSON", 2, err)
		return nil, "", err
	}

	digest := resp.Header.Get("Docker-Content-Digest")

	requestBody := make(map[string]interface{})
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		PrintVerboseErr("Registry.GetManifestJSON", 3, err)
		return nil, "", err
	}

	return requestBody, digest, nil
}

// GetManifest returns the manifest of the image, a token is required
// to perform the request and is generated using GetToken()
func (r *Registry) GetManifest(token string) (*Manifest, error) {
	PrintVerboseInfo("Registry.GetManifest", "running...")

	manifestAPIUrl := fmt.Sprintf("%s/%s/manifests/%s", r.API, settings.Cnf.Name, settings.Cnf.Tag)

	requestBody, digest, err := r.getManifestJSON(manifestAPIUrl, token)
	if err != nil {
		PrintVerboseErr("Registry.GetManifest", 0, fmt.Errorf("failed to fetch manifest"))
		return nil, err
	}

	// If the manifest contains an errors property, it means that the
	// request failed. Ref: https://github.com/Vanilla-OS/ABRoot/issues/285
	if requestBody["errors"] != nil {
		errors := requestBody["errors"].([]interface{})
		for _, e := range errors {
			err := e.(map[string]interface{})
			PrintVerboseErr("Registry.GetManifest", 1, err)
			return nil, fmt.Errorf("Registry error: %s", err["code"])
		}
	}

	// Check if this is a multi-architecture manifest list
	if requestBody["manifests"] != nil {
		PrintVerboseInfo("Registry.GetManifest", "detected multi-architecture manifest list")

		manifests := requestBody["manifests"].([]interface{})
		arch := runtime.GOARCH

		// Find the manifest for the current architecture
		var selectedDigest string
		for _, m := range manifests {
			manifest := m.(map[string]interface{})
			platform := manifest["platform"].(map[string]interface{})
			if platform["architecture"] == arch && platform["os"] == "linux" {
				selectedDigest = manifest["digest"].(string)
				PrintVerboseInfo("Registry.GetManifest", "selected manifest for architecture ", arch, ", digest: ", selectedDigest)
				break
			}
		}

		if selectedDigest == "" {
			PrintVerboseErr("Registry.GetManifest", 2, fmt.Errorf("no manifest found for architecture %s", arch))
			return nil, fmt.Errorf("no manifest found for architecture %s", arch)
		}

		// Fetch the actual manifest using the selected digest
		manifestAPIUrl := fmt.Sprintf("%s/%s/manifests/%s", r.API, settings.Cnf.Name, selectedDigest)

		requestBody, digest, err = r.getManifestJSON(manifestAPIUrl, token)
		if err != nil {
			PrintVerboseErr("Registry.GetManifest", 3, fmt.Errorf("failed to fetch manifest"))
			return nil, err
		}
	}

	// we need to parse the layers to get the digests
	var layerDigests []string
	if requestBody["layers"] == nil && requestBody["fsLayers"] == nil {
		PrintVerboseErr("Registry.GetManifest", 4, err)
		return nil, fmt.Errorf("Manifest does not contain layer property")
	} else if requestBody["layers"] == nil && requestBody["fsLayers"] != nil {
		PrintVerboseWarn("Registry.GetManifest", 4, "layers property not found, using fsLayers")
		layers := requestBody["fsLayers"].([]interface{})
		for _, layer := range layers {
			layerDigests = append(layerDigests, layer.(map[string]interface{})["blobSum"].(string))
		}
	} else {
		layers := requestBody["layers"].([]interface{})
		for _, layer := range layers {
			layerDigests = append(layerDigests, layer.(map[string]interface{})["digest"].(string))
		}
	}

	PrintVerboseInfo("Registry.GetManifest", "success")
	manifest := &Manifest{
		Manifest: requestBody,
		Digest:   digest,
		Layers:   layerDigests,
	}

	return manifest, nil
}
