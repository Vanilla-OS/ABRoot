package tests

import (
	"testing"

	"github.com/vanilla-os/abroot/core"
)

// TestRegistryHasUpdate tests the HasUpdate function by checking if the
// registry has an update with a digest different than our current one.
func TestRegistryHasUpdate(t *testing.T) {
	t.Log("TestRegistryHasUpdate: running...")

	registry := core.NewRegistry()
	if registry == nil {
		t.Fatal("TestRegistryHasUpdate: registry is nil")
	}

	digest, update, err := registry.HasUpdate("impossible_digest")
	if err != nil {
		t.Fatal(err)
	}
	if digest == "" && update == false {
		t.Fatal("TestRegistryHasUpdate: digest and update are empty")
	}

	t.Logf("TestRegistryHasUpdate: digest: %s, update: %t", digest, update)
	t.Log("TestRegistryHasUpdate: done")
}

// TestRegistryGetToken tests the GetToken function by getting a token from the
// registry to authenticate the requests.
func TestRegistryGetToken(t *testing.T) {
	t.Log("TestRegistryGetToken: running...")

	token, err := core.GetToken()
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Fatal("TestRegistryGetToken: token is empty")
	}

	t.Log("TestRegistryGetToken: done")
}

func TestRegistryGetManifest(t *testing.T) {
	t.Log("TestRegistryGetManifest: running...")

	token, err := core.GetToken()
	if err != nil {
		t.Fatal(err)
	}

	registry := core.NewRegistry()
	if registry == nil {
		t.Fatal("TestRegistryGetManifest: registry is nil")
	}

	manifest, err := registry.GetManifest(token)
	if err != nil {
		t.Fatal(err)
	}

	if manifest.Digest == "" {
		t.Fatal("TestRegistryGetManifest: manifest.Digest is empty")
	}

	t.Logf("TestRegistryGetManifest: manifest.Digest: %s", manifest.Digest)
	t.Log("TestRegistryGetManifest: done")
}
