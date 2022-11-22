package core

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

var (
	overlayfsPath = "/tmp/transactionalOverlay"
	overlayfsWork = "/tmp/transactionalOverlayWork"
	combinerPath  = "/tmp/overlayfs-combiner"
)

// UnmountOverlayFS unmounts an overlayfs from the requested path.
func UnmountOverlayFS(path string) error {
	if err := unix.Unmount(path, 0); err != nil {
		if Verbose {
			fmt.Printf("err:UnmountOverlayFS: %s\n", err)
		}
		return err
	}

	return nil
}

// NewOverlayFS creates and mounts an overlayfs.
// It does so by creating the overlayfs in /tmp/transactionalOverlay
// and mounting it to the combinerPath.
func NewOverlayFS(lowers []string) error {
	if AreTransactionsLocked() {
		return fmt.Errorf("transactions are locked")
	}

	if isMounted(combinerPath) {
		return fmt.Errorf("combiner path is busy: %s", combinerPath)
	}

	if isMounted(overlayfsPath) {
		return fmt.Errorf("overlayfs path is busy: %s", overlayfsPath)
	}

	// it is safe to cleanup at this point, as we know that the overlayfs
	// is not mounted
	err := CleanupOverlayPaths()
	if err != nil {
		if Verbose {
			fmt.Printf("err:NewOverlayFS: %s\n", err)
		}
	}

	err = os.Mkdir(overlayfsPath, 0755)
	if err != nil {
		if Verbose {
			fmt.Printf("err:NewOverlayFS: %s\n", err)
		}
	}

	err = os.Mkdir(combinerPath, 0755)
	if err != nil {
		if Verbose {
			fmt.Printf("err:NewOverlayFS: %s\n", err)
		}
	}

	err = os.Mkdir(overlayfsWork, 0755)
	if err != nil {
		if Verbose {
			fmt.Printf("err:NewOverlayFS: %s\n", err)
		}
	}

	lower := ""
	for _, l := range lowers {
		lower += l + ":"
	}

	lower = lower[:len(lower)-1]

	if err := unix.Mount(
		"overlay", combinerPath, "overlay", 0,
		fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, overlayfsPath, overlayfsWork)); err != nil {
		CleanupOverlayPaths()
		if Verbose {
			fmt.Printf("err:NewOverlayFS: %s\n", err)
		}
		return err
	}

	return nil
}

// isMounted checks if a path is mounted.
func isMounted(path string) bool {
	cmd := exec.Command("mountpoint", path)
	if err := cmd.Run(); err != nil {
		if Verbose {
			fmt.Printf("err:isMounted: %s\n", err)
		}
		return false
	}

	return true
}

// MergeOverlayFS unmounts and merges an overlayfs into the original directory.
// Merging is done by rsyncing the overlay into the combiner, unmounting
// the overlayfs, and rsyncing the combiner into the original directory.
func MergeOverlayFS(path string) error {
	if err := AtomicRsync(combinerPath, path, []string{"home", "dev", "proc", "sys", "media", "mnt", "boot", "tmp"}, false); err != nil {
		return err
	}

	if err := unix.Unmount(combinerPath, 0); err != nil {
		// at this point, the overlayfs is already merged into the original
		// directory, so we can safely ignore the error
		fmt.Printf(`an error occurred while unmounting the overlayfs, but it is
already merged into the original directory, so it is safe to ignore it.`)
	}

	return nil
}

// CleanupOverlayPaths unmounts and removes an overlayfs plus the workdir.
func CleanupOverlayPaths() error {
	if isMounted(overlayfsPath) {
		if err := unix.Unmount(overlayfsPath, 0); err != nil {
			if Verbose {
				fmt.Printf("err:CleanupOverlayPaths: %s\n", err)
			}
			return err
		}
	}

	if isMounted(combinerPath) {
		if err := unix.Unmount(combinerPath, 0); err != nil {
			if Verbose {
				fmt.Printf("err:CleanupOverlayPaths: %s\n", err)
			}
			return err
		}
	}

	if err := os.RemoveAll(overlayfsPath); err != nil {
		if Verbose {
			fmt.Printf("err:CleanupOverlayPaths: %s\n", err)
		}
		return err
	}

	if err := os.RemoveAll(overlayfsWork); err != nil {
		if Verbose {
			fmt.Printf("err:CleanupOverlayPaths: %s\n", err)
		}
		return err
	}

	if err := os.RemoveAll(combinerPath); err != nil {
		if Verbose {
			fmt.Printf("err:CleanupOverlayPaths: %s\n", err)
		}
		return err
	}

	return nil
}

// ChrootOverlayFS creates a new overlayfs and chroots into it.
func ChrootOverlayFS(path string, mount bool, command string, catchOut bool) (out string, err error) {
	if mount {
		if err := NewOverlayFS([]string{path}); err != nil {
			return "", err
		}
	}

	if command != "" {
		command = "/bin/bash -c '" + command + "'"
	} else {
		command = "/bin/bash"
	}

	cmd := exec.Command("chroot", combinerPath, command)
	cmd.Env = os.Environ()
	output := []byte{}

	if catchOut {
		output, err = cmd.Output()
	} else {
		if cmd.Stdin == nil {
			cmd.Stdin = os.Stdin
		}
		if cmd.Stdout == nil {
			cmd.Stdout = os.Stdout
		}
		if cmd.Stderr == nil {
			cmd.Stderr = os.Stderr
		}
		err = cmd.Run()
	}

	if err != nil {
		if Verbose {
			fmt.Printf("err:ChrootOverlayFS: %s\n", err)
		}
		return "", err
	}

	out = string(output)

	return out, nil
}
