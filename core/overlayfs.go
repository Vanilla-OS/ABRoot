package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
		PrintVerbose("err:UnmountOverlayFS: %s", err)
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

	if IsMounted(combinerPath) {
		return fmt.Errorf("combiner path is busy: %s", combinerPath)
	}

	if IsMounted(overlayfsPath) {
		return fmt.Errorf("overlayfs path is busy: %s", overlayfsPath)
	}

	// it is safe to cleanup at this point, as we know that the overlayfs
	// is not mounted
	err := CleanupOverlayPaths()
	if err != nil {
		PrintVerbose("err:NewOverlayFS: %s", err)
	}

	err = os.Mkdir(overlayfsPath, 0755)
	if err != nil {
		PrintVerbose("err:NewOverlayFS: %s", err)
	}

	err = os.Mkdir(combinerPath, 0755)
	if err != nil {
		PrintVerbose("err:NewOverlayFS: %s", err)
	}

	err = os.Mkdir(overlayfsWork, 0755)
	if err != nil {
		PrintVerbose("err:NewOverlayFS: %s", err)
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
		PrintVerbose("err:NewOverlayFS: %s", err)
		return err
	}

	bindPaths := []string{"/dev", "/dev/pts", "/proc", "/sys", "/run"}
	for _, path := range bindPaths {
		if err := exec.Command("mount", "--bind", path, combinerPath+path).Run(); err != nil {
			PrintVerbose("err:NewOverlayFS (BindMount): %s", err)
			return err
		}
	}

	bindFromSysPaths := []string{"var", "opt"}
	for _, path := range bindFromSysPaths {
		if err := exec.Command("mount", "--bind", combinerPath+"/.system/"+path, combinerPath+"/"+path).Run(); err != nil {
			PrintVerbose("err:NewOverlayFS (BindMount (.system)): %s", err)
			return err
		}
	}

	return nil
}

// IsMounted checks if a path is mounted.
func IsMounted(path string) bool {
	cmd := exec.Command("mountpoint", path)
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

// IsDeviceMounted checks if a device is mounted.
func IsDeviceMounted(device string) bool {
	cmd := exec.Command("mount")
	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("err:IsDeviceMounted: %s", err)
		return false
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, device) {
			return true
		}
	}

	return false
}

// MergeOverlayFS unmounts and merges an overlayfs into the original directory.
// Merging is done by rsyncing the overlay into the combiner, unmounting
// the overlayfs, and rsyncing the combiner into the original directory.
func MergeOverlayFS(path string) error {
	PrintVerbose("step:  AtomicRsync")
	if err := AtomicRsync(combinerPath+"/.system/", path+"/.system/", path+"/.system_new/", path+"/.system/",
		[]string{"home", "partFuture", "partFuture_new", ".*/"},
		false); err != nil {
		return err
	}

	return nil
}

// CleanupOverlayPaths unmounts and removes an overlayfs plus the workdir.
func CleanupOverlayPaths() error {
	if IsMounted(overlayfsPath) {
		if err := exec.Command("umount", "-l", overlayfsPath).Run(); err != nil {
			PrintVerbose("err:CleanupOverlayPaths: %s", err)
			return err
		}
	}

	if IsMounted(combinerPath) {
		if err := exec.Command("umount", "-l", combinerPath).Run(); err != nil {
			PrintVerbose("err:CleanupOverlayPaths: %s", err)
			return err
		}
	}

	if err := os.RemoveAll(overlayfsPath); err != nil {
		PrintVerbose("err:CleanupOverlayPaths: %s", err)
		return err
	}

	if err := os.RemoveAll(overlayfsWork); err != nil {
		PrintVerbose("err:CleanupOverlayPaths: %s", err)
		return err
	}

	if err := os.RemoveAll(combinerPath); err != nil {
		PrintVerbose("err:CleanupOverlayPaths: %s", err)
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

	args := []string{"chroot", combinerPath, "/bin/bash"}
	if command != "" {
		args = append(args, "-c", command)
	}

	cmd := exec.Command(args[0], args[1:]...)
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
		PrintVerbose("err:ChrootOverlayFS: %s", err)
		return "", err
	}

	out = string(output)

	return out, nil
}
