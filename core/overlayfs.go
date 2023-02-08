package core

import (
	"bytes"
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
	err := unix.Unmount(path, 0)
	if err != nil {
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

	err = unix.Mount(
		"overlay", combinerPath, "overlay", 0,
		fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, overlayfsPath, overlayfsWork))
	if err != nil {
		e := CleanupOverlayPaths()
		if e != nil {
			PrintVerbose("err:NewOverlayFS: %s", e)
		}

		PrintVerbose("err:NewOverlayFS: %s", err)

		return err
	}

	bindPaths := []string{"/dev", "/dev/pts", "/proc", "/sys", "/run"}
	for _, path := range bindPaths {
		err := os.MkdirAll(overlayfsPath+path, 0755)
		if err != nil {
			PrintVerbose("err:NewOverlayFS (MkdirAll): %s", err)
			return err
		}

		err = exec.Command("mount", "--bind", path, combinerPath+path).Run()
		if err != nil {
			PrintVerbose("err:NewOverlayFS (BindMount): %s", err)
			return err
		}
	}

	bindFromBootPaths := []string{"boot", "boot/efi"}
	for _, path := range bindFromBootPaths {
		err := os.MkdirAll(overlayfsPath+"/"+path, 0755)
		if err != nil {
			PrintVerbose("err:NewOverlayFS (MkdirAll (boot)): %s", err)
			return err
		}

		err = exec.Command("mount", "--bind", "/"+path, combinerPath+"/"+path).Run()
		if err != nil {
			PrintVerbose("err:NewOverlayFS (BindMount (.boot)): %s", err)
			return err
		}
	}

	bindFromSysPaths := []string{"var", "opt"}
	for _, path := range bindFromSysPaths {
		err := os.MkdirAll(overlayfsPath+"/"+path, 0755)
		if err != nil {
			PrintVerbose("warn:NewOverlayFS (MkdirAll (.system)): %s", err)
		}

		err = exec.Command("mount", "--bind", combinerPath+"/.system/"+path, combinerPath+"/"+path).Run()
		if err != nil {
			PrintVerbose("warn:NewOverlayFS (BindMount (.system)): %s", err)
		}
	}

	err = PatchMkConfig()
	if err != nil {
		PrintVerbose("err:NewOverlayFS: %s", err)
		return err
	}

	return nil
}

// PatchMkConfig patches the grub-mkconfig script to support variable overrides
// for the GRUB_DEVICE and grub_cfg variables. This is needed since the boot
// partition is not discoverable in the overlayfs and during the transaction
// the grub.cfg is not required and must not be touched at this point, as it
// will be updated later if the whole transaction succeeds.
func PatchMkConfig() error {
	mkConfigPath := combinerPath + "/usr/sbin/grub-mkconfig"

	data, err := os.ReadFile(mkConfigPath)
	if err != nil {
		PrintVerbose("err:PatchMkConfig: %s", err)
		return err
	}

	data = bytes.ReplaceAll(
		data,
		[]byte("GRUB_DEVICE=\"`${grub_probe} --target=device /`\""),
		[]byte("GRUB_DEVICE=${GRUB_DEVICE-\"`${grub_probe} --target=device /`\"}"),
	)

	data = bytes.ReplaceAll(
		data,
		[]byte("if [ \"$EUID\" != 0 ] ; then"),
		[]byte("if [ -n \"$GRUB_CFG\" ]; then\ngrub_cfg=\"$GRUB_CFG\"\nfi\nif [ \"$EUID\" != 0 ] ; then"),
	)

	err = os.WriteFile(mkConfigPath, data, 0755)
	if err != nil {
		PrintVerbose("err:PatchMkConfig: %s", err)
		return err
	}

	return nil
}

// IsMounted checks if a path is mounted.
func IsMounted(path string) bool {
	cmd := exec.Command("mountpoint", path)

	return cmd.Run() == nil
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

	err := AtomicRsync(combinerPath+"/.system/", path+"/.system/", path+"/.system_new/", path+"/.system/",
		[]string{"home", "partFuture", "partFuture_new", ".*/", "tmp", "var/log", "var/tmp", "var/spool", "var/mail"},
		false)
	if err != nil {
		return err
	}

	return nil
}

// CleanupOverlayPaths unmounts and removes an overlayfs plus the workdir.
func CleanupOverlayPaths() error {
	if IsMounted(overlayfsPath) {
		err := exec.Command("umount", "-l", overlayfsPath).Run()
		if err != nil {
			PrintVerbose("err:CleanupOverlayPaths: %s", err)
			return err
		}
	}

	if IsMounted(combinerPath) {
		err := exec.Command("umount", "-l", combinerPath).Run()
		if err != nil {
			PrintVerbose("err:CleanupOverlayPaths: %s", err)
			return err
		}
	}

	err := os.RemoveAll(overlayfsPath)
	if err != nil {
		PrintVerbose("err:CleanupOverlayPaths: %s", err)
		return err
	}

	err = os.RemoveAll(overlayfsWork)
	if err != nil {
		PrintVerbose("err:CleanupOverlayPaths: %s", err)
		return err
	}

	err = os.RemoveAll(combinerPath)
	if err != nil {
		PrintVerbose("err:CleanupOverlayPaths: %s", err)
		return err
	}

	return nil
}

// ChrootOverlayFS creates a new overlayfs and chroots into it.
func ChrootOverlayFS(path string, mount bool, command string, catchOut bool) (out string, err error) {
	bootDevice, err := GetDeviceByMountPoint("/boot")
	if err != nil {
		PrintVerbose("err:ChrootOverlayFS: %s", err)
		return "", err
	}

	if mount {
		err = NewOverlayFS([]string{path})
		if err != nil {
			return "", err
		}
	}

	args := []string{"chroot", combinerPath, "/bin/bash"}
	if command != "" {
		args = append(args, "-c", command)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GRUB_DEVICE="+bootDevice)
	cmd.Env = append(cmd.Env, "GRUB_CFG=/tmp/_abroot_trashed_grub.cfg")
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
