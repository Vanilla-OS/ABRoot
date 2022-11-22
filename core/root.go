package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

// getRootDevice returns the device of requested root partition.
// Note that the present root partition is always the current one, while
// the future root partition is the next one. So, the future root partition
// is detected by checking for the next label, e.g. B if current is A.
func getRootDevice(state string) (string, error) {
	presentLabel, err := getCurrentRootLabel()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if state == "present" {
		return presentLabel, nil
	}

	if presentLabel == "B" {
		return "A", nil
	}

	if presentLabel == "B" {
		return "B", nil
	}

	return "", fmt.Errorf("could not detect future root partition")
}

// getCurrentRootLabel returns the label of the current root partition.
// It does so by checking the label of the current root partition.
func getCurrentRootLabel() (string, error) {
	cmd := exec.Command("lsblk", "-o", "LABEL", "-n", "/")

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// getRootUUID returns the UUID of requested root partition.
func getRootUUID(state string) (string, error) {
	device, err := getRootDevice(state)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("lsblk", "-o", "UUID", "-n", device)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// getRootLabel returns the label of requested root partition.
func getRootLabel(state string) (string, error) {
	device, err := getRootDevice(state)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("lsblk", "-o", "LABEL", "-n", device)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// getRootFileSystem returns the filesystem of requested root partition.
func getRootFileSystem(state string) (string, error) {
	device, err := getRootDevice(state)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("lsblk", "-o", "FSTYPE", "-n", device)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// MountFutureRoot mounts the future root partition to /partB.
func MountFutureRoot() error {
	device, err := GetFutureRootDevice()
	if err != nil {
		return err
	}

	deviceFs, err := getRootFileSystem("future")
	if err != nil {
		return err
	}

	if _, err := os.Stat("/partFuture"); !os.IsNotExist(err) {
		if isMounted("/partFuture") {
			return fmt.Errorf("future root partition is busy. Another transaction?")
		}

		if err := os.RemoveAll("/partFuture"); err != nil {
			return err
		}
	}

	if err := os.Mkdir("/partFuture", 0755); err != nil {
		return err
	}

	if err := unix.Mount(device, "/partFuture", deviceFs, 0, ""); err != nil {
		return err
	}

	return nil
}

// UnmountFutureRoot unmounts the future root partition.
func UnmountFutureRoot() error {
	if err := unix.Unmount("/partB", 0); err != nil {
		return err
	}

	return nil
}

// UpdateRootBoot updates the boot entries for the requested root partition.
// It does so by writing the new boot entries to 10_vanilla, setting the
// future root partition as the first entry, and then updating the boot.
// Note that 10_vanilla is written in both the present and future root
// partitions. If transacting is true, the future partition is not mounted
// at /partFuture, since it should already be there.
func UpdateRootBoot(transacting bool) error {
	presentLabel, err := GetPresentRootLabel()
	if err != nil {
		return err
	}

	presentUUID, err := GetPresentRootUUID()
	if err != nil {
		return err
	}

	futureUUID, err := GetFutureRootUUID()
	if err != nil {
		return err
	}

	if !transacting {
		if AreTransactionsLocked() {
			return fmt.Errorf("another transaction is in progress or a reboot is required")
		}
		if err := MountFutureRoot(); err != nil {
			return err
		}
	}

	bootHeader := "#!/bin/sh\nexec tail -n +3 $0"
	bootEntry := `menuentry 'State %s' {
	search --no-floppy --fs-uuid --set=root %s
	linux   /vmlinuz-%s-generic
	initrd  /initrd.img-%s-generic
}`

	presentKernelVersion, err := getKernelVersion("present")
	if err != nil {
		return err
	}

	futureKernelVersion, err := getKernelVersion("future")
	if err != nil {
		return err
	}

	bootPresent := fmt.Sprintf(bootEntry, "A", presentUUID, presentKernelVersion, presentKernelVersion)
	bootFuture := fmt.Sprintf(bootEntry, "B", futureUUID, futureKernelVersion, futureKernelVersion)
	bootTemplate := fmt.Sprintf("%s\n%s\n%s", bootHeader, bootPresent, bootFuture)

	if err := os.WriteFile("/partFuture/etc/grub.d/10_vanilla", []byte(bootTemplate), 0755); err != nil {
		return err
	}

	if err := os.WriteFile("/etc/grub.d/10_vanilla", []byte(bootTemplate), 0755); err != nil {
		return err
	}

	if err := switchBootDefault(presentLabel); err != nil {
		return err
	}

	if err := updateGrubConfig(); err != nil {
		return err
	}

	return nil
}

// getKernelVersion returns the highest kernel version installed on the
// requested partition.
func getKernelVersion(state string) (string, error) {
	command := []string{"dpkg", "--list", "|", "grep", "linux-image", "|", "awk", "'{print $3}'", "|", "sort", "-V", "|", "tail", "-n", "1"}
	if state == "future" {
		command = append([]string{"chroot", "/partFuture"}, command...)
	}

	cmd := exec.Command(command[0], command[1:]...)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.Split(string(out), ".")[0], nil
}

// switchBootDefault updates the GRUB_DEFAULT variable in both the present
// and future root partitions. It does so by comparing the current present
// root partition label. E.g. if the present root partition is labeled
// "A" (0), then the future root partition is labeled "B" (1), then the
// GRUB_DEFAULT variable is set to "1" in both partitions.
func switchBootDefault(presentLabel string) error {
	var newGrubDefault string
	if presentLabel == "A" {
		newGrubDefault = "1"
	} else {
		newGrubDefault = "0"
	}

	if err := os.WriteFile("/etc/default/grub", []byte(fmt.Sprintf("GRUB_DEFAULT=%s", newGrubDefault)), 0644); err != nil {
		return err
	}

	if err := os.WriteFile("/partFuture/etc/default/grub", []byte(fmt.Sprintf("GRUB_DEFAULT=%s", newGrubDefault)), 0644); err != nil {
		return err
	}

	return nil
}

// updateGrubConfig updates the grub configuration for both the future
// and present root partitions.
func updateGrubConfig() error {
	if err := exec.Command("chroot", "/partFuture", "grub-mkconfig", "-o", "/boot/grub/grub.cfg").Run(); err != nil {
		return err
	}

	if err := exec.Command("grub-mkconfig", "-o", "/boot/grub/grub.cfg").Run(); err != nil {
		return err
	}

	return nil
}

func GetPresentRootDevice() (string, error) {
	return getRootDevice("present")
}

func GetFutureRootDevice() (string, error) {
	return getRootDevice("future")
}

func GetPresentRootLabel() (string, error) {
	return getRootLabel("present")
}

func GetFutureRootLabel() (string, error) {
	return getRootLabel("future")
}

func GetPresentRootUUID() (string, error) {
	return getRootUUID("present")
}

func GetFutureRootUUID() (string, error) {
	return getRootUUID("future")
}
