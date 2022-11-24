package core

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"golang.org/x/sys/unix"
)

func CheckABRequirements() {
	if !DoesSupportAB() {
		fmt.Println("Your system does not support A/B root.")
		os.Exit(1)
	}
}

// getRootDevice returns the device of requested root partition.
// Note that the present root partition is always the current one, while
// the future root partition is the next one. So, the future root partition
// is detected by checking for the next label, e.g. B if current is A.
func getRootDevice(state string) (string, error) {
	presentLabel, err := getCurrentRootLabel()
	if err != nil {
		return "", err
	}

	if state == "present" {
		device, err := getDeviceByLabel(presentLabel)
		if err != nil {
			return "", err
		}
		return device, nil
	}

	if presentLabel == "A" {
		device, err := getDeviceByLabel("B")
		if err != nil {
			return "", err
		}
		return device, nil
	}

	if presentLabel == "B" {
		device, err := getDeviceByLabel("A")
		if err != nil {
			return "", err
		}
		return device, nil
	}

	return "", fmt.Errorf("could not detect future root partition")
}

// getCurrentRootLabel returns the label of the current root partition.
// It does so by checking the label of the current root partition.
func getCurrentRootLabel() (string, error) {
	device, err := getDeviceByMountPoint("/")
	if err != nil {
		return "", err
	}

	cmd := exec.Command("lsblk", "-o", "LABEL", "-n", device)

	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("err:getCurrentRootLabel: %s\n", err)
		return "", err
	}

	label := strings.TrimSpace(string(out))
	return label, nil
}

// getDeviceByMountPoint returns the device of the requested mount point.
func getDeviceByMountPoint(mountPoint string) (string, error) {
	cmd := exec.Command("lsblk", "-o", "MOUNTPOINT,NAME", "-AnM", "--tree=PATH")

	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("err:getDeviceByMountPoint: %s\n", err)
		return "", err
	}

	for _, line := range strings.Split(string(out), "\n") {
		split := strings.Fields(line)
		if len(split) != 2 {
			continue
		}
		if split[0] == mountPoint {
			return "/dev/" + split[1], nil
		}
	}

	return "", fmt.Errorf("could not find device for mount point %s", mountPoint)
}

// getDeviceByLabel returns the device of the requested label.
func getDeviceByLabel(label string) (string, error) {
	cmd := exec.Command("lsblk", "-o", "LABEL,NAME", "-AnM", "--tree=PATH")

	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("err:getDeviceByLabel: %s\n", err)
		return "", err
	}

	for _, line := range strings.Split(string(out), "\n") {
		split := strings.Fields(line)
		if len(split) != 2 {
			continue
		}
		if split[0] == label {
			return "/dev/" + split[1], nil
		}
	}

	return "", fmt.Errorf("could not find device for label %s", label)
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
		PrintVerbose("err:getRootUUID: %s\n", err)
		return "", err
	}

	return string(out), nil
}

// GetBootUUID returns the UUID of the boot partition.
func GetBootUUID() (string, error) {
	device, err := getDeviceByMountPoint("/boot")
	if err != nil {
		device, err = getDeviceByMountPoint("/partFuture/boot")
		if err != nil {
			device, err = getDeviceByLabel("boot")
			if err != nil {
				return "", err
			}
		}
	}

	cmd := exec.Command("lsblk", "-o", "UUID", "-n", device)
	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("err:GetBootUUID: %s\n", err)
		return "", err
	}

	return string(out), nil
}

// getRootLabel returns the label of requested root partition.
func getRootLabel(state string) (string, error) {
	presentLabel, err := getCurrentRootLabel()
	if err != nil {
		return "", err
	}

	if state == "present" {
		return presentLabel, nil
	}

	if presentLabel == "A" {
		return "B", nil
	}

	if presentLabel == "B" {
		return "A", nil
	}

	return "", fmt.Errorf("partitions are not labeled correctly")
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
		PrintVerbose("err:getRootFileSystem: %s\n", err)
		return "", err
	}

	return string(out), nil
}

// MountFutureRoot mounts the future root partition to /partB.
func MountFutureRoot() error {
	PrintVerbose("step:  GetFutureRootDevice")
	device, err := GetFutureRootDevice()
	if err != nil {
		return err
	}

	PrintVerbose("step:  getRootFileSystem")
	if err != nil {
		return err
	}

	if _, err := os.Stat("/partFuture"); !os.IsNotExist(err) {
		PrintVerbose("step:  IsMounted")
		if IsMounted("/partFuture") {
			return fmt.Errorf("future root partition is busy. Another transaction?")
		}

		if err := os.RemoveAll("/partFuture"); err != nil {
			PrintVerbose("err:MountFutureRoot: %s\n", err)
			return err
		}
	}

	if err := os.Mkdir("/partFuture", 0755); err != nil {
		PrintVerbose("err:MountFutureRoot: %s\n", err)
		return err
	}

	PrintVerbose("step:  Mount device: %s", device)
	// unix.Mount does not work here for some reason.
	cmd := exec.Command("mount", device, "/partFuture")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		PrintVerbose("err:MountFutureRoot: %s\n", err)
		return err
	}

	return nil
}

// UnmountFutureRoot unmounts the future root partition.
func UnmountFutureRoot() error {
	if err := unix.Unmount("/partFuture", 0); err != nil {
		PrintVerbose("err:UnmountFutureRoot: %s\n", err)
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
	PrintVerbose("step:  GetPresentRootLabel")
	presentLabel, err := GetPresentRootLabel()
	if err != nil {
		return err
	}

	PrintVerbose("step:  GetPresentRootUUID")
	presentUUID, err := GetPresentRootUUID()
	if err != nil {
		return err
	}

	PrintVerbose("step:  GetFutureRootUUID")
	futureUUID, err := GetFutureRootUUID()
	if err != nil {
		return err
	}

	PrintVerbose("step:  GetBootUUID")
	bootUUID, err := GetBootUUID()
	if err != nil {
		return err
	}

	if !transacting {
		PrintVerbose("step:  MountFutureRoot")
		if err := MountFutureRoot(); err != nil {
			return err
		}
	}

	bootHeader := "#!/bin/sh\nexec tail -n +3 $0"
	bootEntry := `menuentry 'State %s' {
	search --no-floppy --fs-uuid --set=root %s
	linux   /vmlinuz-%s  root=UUID=%s ro quiet loglevel=7 splash $vt_handoff
	initrd  /initrd.img-%s
}`

	PrintVerbose("step:  getKernelVersion present")
	presentKernelVersion, err := getKernelVersion("present")
	if err != nil {
		PrintVerbose("err:UpdateRootBoot: %s\n", err)
		return err
	}

	PrintVerbose("step:  getKernelVersion future")
	futureKernelVersion, err := getKernelVersion("future")
	if err != nil {
		PrintVerbose("err:UpdateRootBoot: %s\n", err)
		return err
	}

	bootPresent := fmt.Sprintf(bootEntry, "A", bootUUID, presentKernelVersion, presentUUID, presentKernelVersion)
	bootFuture := fmt.Sprintf(bootEntry, "B", bootUUID, futureKernelVersion, futureUUID, futureKernelVersion)
	bootTemplate := fmt.Sprintf("%s\n%s\n%s", bootHeader, bootPresent, bootFuture)

	PrintVerbose("step:  WriteFile future")
	if err := os.WriteFile("/partFuture/etc/grub.d/10_vanilla", []byte(bootTemplate), 0755); err != nil {
		PrintVerbose("err:UpdateRootBoot: %s\n", err)
		return err
	}

	PrintVerbose("step:  WriteFile present")
	if err := os.WriteFile("/etc/grub.d/10_vanilla", []byte(bootTemplate), 0755); err != nil {
		PrintVerbose("err:UpdateRootBoot: %s\n", err)
		return err
	}

	PrintVerbose("step:  switchBootDefault")
	if err := switchBootDefault(presentLabel); err != nil {
		return err
	}

	PrintVerbose("step:  updateGrubConfig")
	if err := updateGrubConfig(); err != nil {
		return err
	}

	return nil
}

// getKernelVersion returns the highest kernel version installed on the
// requested partition.
func getKernelVersion(state string) (string, error) {
	var dir string
	if state == "present" {
		dir = "/usr/lib/modules"
	} else if state == "future" {
		dir = "/partFuture/.system/usr/lib/modules"
	} else {
		return "", fmt.Errorf("invalid state: %s", state)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		PrintVerbose("err:getKernelVersion: %s\n", err)
		return "", err
	}

	var versions []string
	for _, file := range files {
		if file.IsDir() {
			versions = append(versions, file.Name())
		}
	}

	sort.Strings(versions)
	res := versions[len(versions)-1]
	return res, nil
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
		PrintVerbose("err:switchBootDefault: %s\n", err)
		return err
	}

	if err := os.WriteFile("/partFuture/etc/default/grub", []byte(fmt.Sprintf("GRUB_DEFAULT=%s", newGrubDefault)), 0644); err != nil {
		PrintVerbose("err:switchBootDefault: %s\n", err)
		return err
	}

	return nil
}

// updateGrubConfig updates the grub configuration for both the future
// and present root partitions.
func updateGrubConfig() error {
	bootPart, err := getDeviceByMountPoint("/boot")
	if err != nil {
		return err
	}

	if IsDeviceMounted(bootPart) {
		if err := exec.Command("umount", "-l", bootPart).Run(); err != nil {
			PrintVerbose("err:updateGrubConfig (Unmount): %s\n", err)
			return err
		}
	}

	if _, err := os.Stat("/partFuture/boot"); os.IsNotExist(err) {
		if err := os.Mkdir("/partFuture/boot", 0755); err != nil {
			PrintVerbose("err:updateGrubConfig (Mkdir): %s\n", err)
			return err
		}
	}

	if err := exec.Command("mount", bootPart, "/partFuture/boot").Run(); err != nil {
		PrintVerbose("err:updateGrubConfig (Mount): %s\n", err)
		return err
	}

	bindPaths := []string{"/dev", "/dev/pts", "/proc", "/sys"}
	for _, path := range bindPaths {
		if err := exec.Command("mount", "--bind", path, "/partFuture"+path).Run(); err != nil {
			PrintVerbose("err:updateGrubConfig (BindMount): %s\n", err)
			return err
		}
	}

	if err := exec.Command("chroot", "/partFuture", "grub-mkconfig", "-o", "/boot/grub/grub.cfg").Run(); err != nil {
		PrintVerbose("err:updateGrubConfig (chroot): %s\n", err)
		return err
	}

	if err := exec.Command("umount", "-l", bootPart).Run(); err != nil {
		PrintVerbose("err:updateGrubConfig (Unmount-2): %s\n", err)
		return err
	}

	if err := exec.Command("mount", bootPart, "/boot").Run(); err != nil {
		PrintVerbose("err:updateGrubConfig (Mount-2): %s\n", err)
		return err
	}

	if err := exec.Command("grub-mkconfig", "-o", "/boot/grub/grub.cfg").Run(); err != nil {
		PrintVerbose("err:updateGrubConfig: %s\n", err)
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

/* DoesSupportAB check if the current system supports A/B partitioning */
func DoesSupportAB() bool {
	var support bool = true

	_, err := GetPresentRootDevice()
	if err != nil {
		PrintVerbose("err:DoesSupportAB: %s\n", err)
		support = false
	}

	_, err = GetFutureRootDevice()
	if err != nil {
		PrintVerbose("err:DoesSupportAB: %s\n", err)
		support = false
	}

	return support
}
