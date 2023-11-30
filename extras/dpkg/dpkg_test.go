package dpkg

import (
	"fmt"
	"testing"
)

func TestGetPackageVersion(t *testing.T) {
	version := DpkgGetPackageVersion("git")
	fmt.Printf("Version: %s\n", version)

	if version != "1:2.40.1-1" {
		t.Fail()
	}
}

func TestBatchGetPackageVersion(t *testing.T) {
	versions := DpkgBatchGetPackageVersion([]string{"git", "golang"})
	fmt.Printf("Versions: %v\n", versions)

	if versions[0] != "1:2.40.1-1" || versions[1] != "2:1.21~2" {
		t.Fail()
	}
}
