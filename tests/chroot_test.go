package tests

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/google/uuid"
	"github.com/vanilla-os/abroot/core"
)

func TestChroot(t *testing.T) {
	if os.Getenv("ABROOT_TEST_ROOTFS_URL") == "" {
		t.Skip("ABROOT_TEST_ROOTFS_URL is not set")
	}

	chrootPath := fmt.Sprintf("%s/chroot-%s", os.TempDir(), uuid.New().String())
	chrootUuid := uuid.New().String() // fake
	chrootDevice := "/dev/fk1"        // fake

	err := os.MkdirAll(chrootPath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// rootfs setup
	_, err = os.Stat(fmt.Sprintf("%s/rootfs.tar.gz", os.TempDir()))
	if err != nil {
		resp, err := http.Get(os.Getenv("ABROOT_TEST_ROOTFS_URL"))
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Create(fmt.Sprintf("%s/rootfs.tar.gz", os.TempDir()))
		if err != nil {
			t.Fatal(err)
		}

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	err = exec.Command("tar", "-xzf", fmt.Sprintf("%s/rootfs.tar.gz", os.TempDir()), "-C", chrootPath).Run()
	if err != nil {
		t.Fatal(err)
	}

	// chroot setup
	chroot, err := core.NewChroot(chrootPath, chrootUuid, chrootDevice)
	if err != nil {
		t.Fatal(err)
	}

	// chroot test
	err = chroot.Execute("touch", []string{"/test"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(fmt.Sprintf("%s/test", chrootPath))
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(fmt.Sprintf("%s/test", chrootPath))
	if err != nil {
		t.Fatal(err)
	}

	// chroot teardown
	err = chroot.Close()
	if err != nil {
		t.Fatal(err)
	}
}
