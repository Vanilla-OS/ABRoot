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
	combinerPath  = "/usr/bin/overlayfs-combiner"
)

func UnmountOverlayFS(path string) error {
	/*
	 * UnmountOverlayFS unmounts an overlayfs from the requested path.
	 */
	if err := unix.Unmount(path, 0); err != nil {
		return err
	}

	return nil
}

func NewOverlayFS(lowers []string) error {
	/*
	 * NewOverlayFS creates and mounts an overlayfs.
	 * It does so by creating the overlayfs in /tmp/transactionalOverlay
	 * and mounting it to the combinerPath.
	 */
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

	CleanupOverlayPaths()

	err := os.Mkdir(overlayfsPath, 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir(combinerPath, 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir(overlayfsWork, 0755)
	if err != nil {
		return err
	}

	lower := ""
	for _, l := range lowers {
		lower += l + ":"
	}
	lower = lower[:len(lower)-1]

	if err := unix.Mount(
		"overlay", combinerPath, "overlay", 0,
		fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, overlayfsPath, overlayfsWork)); err != nil {
		return err
	}

	return nil
}

func isMounted(path string) bool {
	/*
	 * isMounted checks if a path is mounted.
	 */
	cmd := exec.Command("mountpoint", path)
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func MergeOverlayFS(path string) error {
	/*
	 * MergeOverlayFS unmounts and merges an overlayfs into the original directory.
	 * Merging is done by rsyncing the overlay into the combiner, unmounting
	 * the overlayfs, and rsyncing the combiner into the original directory.
	 */
	if err := AtomicRsync(combinerPath, path, []string{"home", "dev", "proc", "sys", "media", "mnt", "boot", "tmp"}, false); err != nil {
		return err
	}

	if err := unix.Unmount(combinerPath, 0); err != nil {
		// at this point, the overlayfs is already merged into the original
		// directory, so we can safely ignore the error
		fmt.Printf(`an error occured while unmounting the overlayfs, but it is
already merged into the original directory, so it is safe to ignore it.`)
	}

	return nil
}

func CleanupOverlayPaths() error {
	/*
	 * CleanupOverlayPaths unmounts and removes an overlayfs plus the workdir.
	 */
	if isMounted(overlayfsPath) {
		if err := unix.Unmount(overlayfsPath, 0); err != nil {
			fmt.Printf("failed to unmount overlayfs: %s", err)
			return err
		}
	}

	if isMounted(combinerPath) {
		if err := unix.Unmount(combinerPath, 0); err != nil {
			fmt.Printf("failed to unmount combiner: %s", err)
			return err
		}
	}

	if err := os.RemoveAll(overlayfsPath); err != nil {
		fmt.Printf("failed to remove overlayfs: %s", err)
		return err
	}

	if err := os.RemoveAll(overlayfsWork); err != nil {
		fmt.Printf("failed to remove overlayfs workdir: %s", err)
		return err
	}

	if err := os.RemoveAll(combinerPath); err != nil {
		fmt.Printf("failed to remove combiner: %s", err)
		return err
	}

	return nil
}

func ChrootOverlayFS(path string, mount bool, command string) (out string, err error) {
	/*
	 * ChrootOverlayFS creates a new overlayfs and chroots into it.
	 */
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
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if output, err := cmd.Output(); err != nil {
		return "", err
	} else {
		out = string(output)
	}

	return out, nil
}
