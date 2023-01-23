package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/vanilla-os/orchid/cmdr"
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

	if presentLabel == "a" {
		device, err := getDeviceByLabel("b")
		if err != nil {
			return "", err
		}
		return device, nil
	}

	if presentLabel == "b" {
		device, err := getDeviceByLabel("a")
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
	device, err := GetDeviceByMountPoint("/")
	if err != nil {
		return "", err
	}

	cmd := exec.Command("lsblk", "-o", "LABEL", "-n", device)

	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("err:getCurrentRootLabel: %s", err)
		return "", err
	}

	label := strings.TrimSpace(string(out))
	label = strings.ToLower(label)
	return label, nil
}

// GetDeviceByMountPoint returns the device of the requested mount point.
func GetDeviceByMountPoint(mountPoint string) (string, error) {
	cmd := exec.Command("lsblk", "-o", "MOUNTPOINT,NAME", "-AnM", "--tree=PATH")

	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("err:GetDeviceByMountPoint: %s", err)
		return "", err
	}

	for _, line := range strings.Split(string(out), "\n") {
		split := strings.Fields(line)
		if len(split) != 2 {
			continue
		}
		if strings.HasSuffix(split[0], mountPoint) {
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
		PrintVerbose("err:getDeviceByLabel: %s", err)
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
		PrintVerbose("err:getRootUUID: %s", err)
		return "", err
	}

	uuid := strings.TrimSpace(string(out))
	return uuid, nil
}

// GetBootUUID returns the UUID of the boot partition.
func GetBootUUID() (string, error) {
	device, err := GetDeviceByMountPoint("/boot")
	if err != nil {
		device, err = GetDeviceByMountPoint("/partFuture/boot")
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
		PrintVerbose("err:GetBootUUID: %s", err)
		return "", err
	}

	uuid := strings.TrimSpace(string(out))
	return uuid, nil
}

// GetEfiUUID returns the UUID of the EFI partition.
func GetEfiUUID() (string, error) {
	device, err := GetDeviceByMountPoint("/boot/efi")
	if err != nil {
		device, err = GetDeviceByMountPoint("/partFuture/boot/efi")
		if err != nil {
			device, err = getDeviceByLabel("efi")
			if err != nil {
				return "", err
			}
		}
	}

	cmd := exec.Command("lsblk", "-o", "UUID", "-n", device)
	out, err := cmd.Output()
	if err != nil {
		PrintVerbose("err:GetEfiUUID: %s", err)
		return "", err
	}

	uuid := strings.TrimSpace(string(out))
	return uuid, nil
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

	if presentLabel == "a" {
		return "b", nil
	}

	if presentLabel == "b" {
		return "a", nil
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
		PrintVerbose("err:getRootFileSystem: %s", err)
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
			PrintVerbose("err:MountFutureRoot: %s", err)
			return err
		}
	}

	if err := os.Mkdir("/partFuture", 0755); err != nil {
		PrintVerbose("err:MountFutureRoot: %s", err)
		return err
	}

	PrintVerbose("step:  Mount device: %s", device)
	// unix.Mount does not work here for some reason.
	cmd := exec.Command("mount", device, "/partFuture")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		PrintVerbose("err:MountFutureRoot: %s", err)
		return err
	}

	return nil
}

// UnmountFutureRoot unmounts the future root partition.
func UnmountFutureRoot() error {
	if err := exec.Command("umount", "-l", "/partFuture").Run(); err != nil {
		PrintVerbose("err:UnmountFutureRoot: %s", err)
		return err
	}

	return nil
}

// GetKargs reads current kernel arguments from GRUB config.
func GetKargs(state string) (string, error) {
	file, err := os.Open("/etc/grub.d/10_vanilla")
	if err != nil {
		return "", err
	}

	var kargs_lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "linux\t/vmlinuz") {
			splits := strings.Split(scanner.Text(), " ")
			kargs_lines = append(kargs_lines, strings.Join(splits[2:len(splits)-1], " "))
		}
	}

	presentLabel, err := GetPresentRootLabel()
	if err != nil {
		return "", err
	}

	switch state {
	case "present":
		if presentLabel == "a" {
			return kargs_lines[0], nil
		}
		return kargs_lines[1], nil
	case "future":
		if presentLabel == "a" {
			return kargs_lines[1], nil
		}
		return kargs_lines[0], nil
	default:
		return "", errors.New(fmt.Sprintf("Invalid state %s", state))
	}
}

// GetCurrentKargs reads current kernel arguments from GRUB config.
func GetCurrentKargs() (string, error) {
	return GetKargs("present")
}

// GetFutureKargs reads future kernel arguments from GRUB config.
func GetFutureKargs() (string, error) {
	return GetKargs("future")
}

// UpdateRootBoot updates the boot entries for the requested root partition.
// It does so by writing the new boot entries to 10_vanilla, setting the
// future root partition as the first entry, and then updating the boot.
// Note that 10_vanilla is written in both the present and future root
// partitions. If transacting is true, the future partition is not mounted
// at /partFuture, since it should already be there. kargs should be a space
// separated list of kernel arguments to be added in both boot entries.
func UpdateRootBoot(transacting bool, kargs string) error {
	unwanted := []string{"10_linux", "20_memtest86+"} // those files cause undesired boot entries

	PrintVerbose("step:  GetPresentRootLabel")
	presentLabel, err := GetPresentRootLabel()
	if err != nil {
		return err
	}

	PrintVerbose("step:  GetFutureRootLabel")
	futureLabel, err := GetFutureRootLabel()
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

	bootHeader := `#!/bin/sh
exec tail -n +3 $0

set menu_color_normal=white/black
set menu_color_highlight=black/light-gray

function gfxmode {
	set gfxpayload="${1}"
	if [ "${1}" = "keep" ]; then
			set vt_handoff=vt.handoff=7
	else
			set vt_handoff=
	fi
}
if [ "${recordfail}" != 1 ]; then
  if [ -e ${prefix}/gfxblacklist.txt ]; then
    if [ ${grub_platform} != pc ]; then
      set linux_gfx_mode=keep
    elif hwmatch ${prefix}/gfxblacklist.txt 3; then
      if [ ${match} = 0 ]; then
        set linux_gfx_mode=keep
      else
        set linux_gfx_mode=text
      fi
    else
      set linux_gfx_mode=text
    fi
  else
    set linux_gfx_mode=keep
  fi
else
  set linux_gfx_mode=text
fi
export linux_gfx_mode
`
	// root is mounted rw since protected paths are managed by sysd and fstab
	bootEntry := `menuentry 'Vanilla OS - Root %s (%s)' --class gnu-linux --class gnu --class os {
	recordfail
	load_video
	gfxmode $linux_gfx_mode
	insmod gzio
	if [ x$grub_platform = xxen ]; then insmod xzio; insmod lzopio; fi
	insmod part_gpt
	insmod ext2
	search --no-floppy --fs-uuid --set=root %s
	linux	/vmlinuz-%s root=UUID=%s %s $vt_handoff
	initrd  /initrd.img-%s
}
`
	PrintVerbose("step:  getKernelVersion present")
	presentKernelVersion, err := getKernelVersion("present")
	if err != nil {
		PrintVerbose("err:UpdateRootBoot: %s", err)
		return err
	}

	PrintVerbose("step:  getKernelVersion future")
	futureKernelVersion, err := getKernelVersion("future")
	if err != nil {
		PrintVerbose("err:UpdateRootBoot: %s", err)
		return err
	}

	old_kargs, err := GetCurrentKargs()
	if err != nil {
		return err
	}

	var boot_a, boot_b string
	if presentLabel == "a" {
		boot_a = fmt.Sprintf(bootEntry, presentLabel, "previous", bootUUID, presentKernelVersion, presentUUID, old_kargs, presentKernelVersion)
		boot_b = fmt.Sprintf(bootEntry, futureLabel, "current", bootUUID, futureKernelVersion, futureUUID, kargs, futureKernelVersion)
	} else {
		boot_a = fmt.Sprintf(bootEntry, futureLabel, "current", bootUUID, futureKernelVersion, futureUUID, kargs, futureKernelVersion)
		boot_b = fmt.Sprintf(bootEntry, presentLabel, "previous", bootUUID, presentKernelVersion, presentUUID, old_kargs, presentKernelVersion)
	}
	bootTemplate := fmt.Sprintf("%s\n%s\n%s", bootHeader, boot_a, boot_b)

	PrintVerbose("step:  WriteFile future")
	if err := os.WriteFile("/partFuture/etc/grub.d/10_vanilla", []byte(bootTemplate), 0755); err != nil {
		PrintVerbose("err:UpdateRootBoot: %s", err)
		return err
	}

	PrintVerbose("step:  WriteFile present")
	if err := os.WriteFile("/etc/grub.d/10_vanilla", []byte(bootTemplate), 0755); err != nil {
		PrintVerbose("err:UpdateRootBoot: %s", err)
		return err
	}

	PrintVerbose("step:  Remove unwanted grub files in future")
	for _, file := range unwanted {
		os.Remove("/partFuture/etc/grub.d/" + file)
	}

	PrintVerbose("step:  Remove unwanted grub files in present")
	for _, file := range unwanted {
		os.Remove("/etc/grub.d/" + file)
	}

	PrintVerbose("step:  switchBootDefault")
	next := "0"
	if next, err = switchBootDefault(presentLabel); err != nil {
		return err
	}

	PrintVerbose("step:  updateGrubConfig")
	if err := updateGrubConfig(); err != nil {
		return err
	}

	PrintVerbose("step:  verifyGrubConfig")
	if err := verifyGrubConfig(next); err != nil {
		return err
	}

	return nil
}

// UpdateFsTab updates the fstab file to reflect the new root partition.
func UpdateFsTab() error {
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

	PrintVerbose("step:  ReadFile present")
	fstab, err := os.ReadFile("/etc/fstab")
	if err != nil {
		PrintVerbose("err:updateFsTab: %s", err)
		return err
	}

	PrintVerbose("step:  ReplaceAll future")
	fstabFuture := fstab
	fstabFuture = bytes.ReplaceAll(fstabFuture, []byte(presentUUID), []byte(futureUUID))

	PrintVerbose("step: Configure Bind Mounts") // those are needed since symlinked directories cause issues with flatpak and snap
	bindMounts := []string{"/var", "/opt"}
	for _, bindMount := range bindMounts {
		if !strings.Contains(string(fstabFuture), bindMount) {
			fstabFuture = append(fstabFuture, []byte("\n"+"/.system"+bindMount+" "+bindMount+" none bind 0 0")...)
		}
	}

	PrintVerbose("step:  WriteFile future")
	if err := os.WriteFile("/partFuture/etc/fstab", fstabFuture, 0644); err != nil {
		PrintVerbose("err:updateFsTab: %s", err)
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
		PrintVerbose("err:getKernelVersion: %s", err)
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
// "a" (0), then the future root partition is labeled "b" (1), then the
// GRUB_DEFAULT variable is set to "1" in both partitions.
func switchBootDefault(presentLabel string) (next string, err error) {
	var newGrubDefault string
	if presentLabel == "a" {
		newGrubDefault = "1"
	} else {
		newGrubDefault = "0"
	}

	grubContent := `GRUB_DEFAULT=%s
GRUB_TIMEOUT=0
GRUB_HIDDEN_TIMEOUT=2
GRUB_TIMEOUT_STYLE=hidden`
	if err := os.WriteFile("/etc/default/grub", []byte(fmt.Sprintf(grubContent, newGrubDefault)), 0644); err != nil {
		PrintVerbose("err:switchBootDefault: %s", err)
		return "", err
	}

	if err := os.WriteFile("/partFuture/etc/default/grub", []byte(fmt.Sprintf(grubContent, newGrubDefault)), 0644); err != nil {
		PrintVerbose("err:switchBootDefault: %s", err)
		return "", err
	}

	return newGrubDefault, nil
}

// verifyGrubConfig verifies that the GRUB_DEFAULT variable is set to the
// expected value, otherwise it replaces the it with the expected one.
// This is more a workaround since mkconfig does not respect the
// GRUB_DEFAULT variable for some reason.
func verifyGrubConfig(expected string) error {
	grubCfg, err := os.ReadFile("/boot/grub/grub.cfg")
	if err != nil {
		PrintVerbose("err:verifyGrubConfig: %s", err)
		return err
	}

	if !bytes.Contains(grubCfg, []byte(fmt.Sprintf("set default=\"%s\"", expected))) {
		if expected == "0" {
			grubCfg = bytes.ReplaceAll(grubCfg, []byte("set default=\"1\""), []byte("set default=\"0\""))
		} else {
			grubCfg = bytes.ReplaceAll(grubCfg, []byte("set default=\"0\""), []byte("set default=\"1\""))
		}

		if err := os.WriteFile("/boot/grub/grub.cfg", grubCfg, 0644); err != nil {
			PrintVerbose("err:verifyGrubConfig: %s", err)
			return err
		}
	}

	return nil
}

// updateGrubConfig updates the grub configuration for both the future
// and present root partitions.
func updateGrubConfig() error {
	// NOTE: in this funcion we are assuming that /boot and /boot/efi are
	// mounted. ABRoot will fail if they are not. Will need to revisit
	// this by moving the presentUUID, futureUUID, bootUUID, efiUUID to
	// the init function, then check remount every time we need to
	// access them. This is not a priority, since the whole transaction
	// will safely fail if the user unmounts them.
	commandList := [][]string{
		{"mount", "-t", "proc", "/proc", "/partFuture/proc"},
		{"mount", "-t", "sysfs", "/sys", "/partFuture/sys"},
		{"mount", "--rbind", "/dev", "/partFuture/dev"},
		{"mount", "--rbind", "/run", "/partFuture/run"},
		{"mount", "--rbind", "/sys/firmware/efi/efivars", "/partFuture/sys/firmware/efi/efivars"},
	}
	for _, command := range commandList {
		cmd := exec.Command(command[0], command[1:]...)
		err := cmd.Run()
		if err != nil {
			if strings.Contains(cmd.String(), "efivars") {
				continue // Ignore error, we are probably not on UEFI
			}
			PrintVerbose("err:updateGrubConfig (BindMount): %s", err)
			PrintVerbose("err:updateGrubConfig (BindMount): command: %s", command)
			return err
		}
	}

	if err := exec.Command("grub-mkconfig", "-o", "/boot/grub/grub.cfg").Run(); err != nil {
		PrintVerbose("err:updateGrubConfig: %s", err)
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
		PrintVerbose("err:DoesSupportAB: %s", err)
		support = false
	}

	_, err = GetFutureRootDevice()
	if err != nil {
		PrintVerbose("err:DoesSupportAB: %s", err)
		support = false
	}

	return support
}

/* SetMutablePath sets the i attribute of the given path to mutable. */
func SetMutablePath(path string) error {
	if err := exec.Command("chattr", "-i", path).Run(); err != nil {
		PrintVerbose("err:SetMutablePath: %s", err)
		return err
	}

	return nil
}

/* SetImmutablePath sets the i attribute of the given path to immutable. */
func SetImmutablePath(path string) error {
	if err := exec.Command("chattr", "+i", path).Run(); err != nil {
		PrintVerbose("err:SetImmutablePath: %s", err)
		return err
	}

	return nil
}

// getGrubMarkedRoot obtains the root name ("a", "b") of the entry marked as state ("current", "previous")
func getGRUBMarkedRoot(state string) (string, error) {
	file, err := os.Open("/etc/grub.d/10_vanilla")
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "("+state+")") {
			splits := strings.Split(scanner.Text(), " ")
			return splits[5], nil
		}
	}

	return "", errors.New(fmt.Sprintf("No partition found for state %s", state))
}

// Rollback switches back to the previous root.
func Rollback() error {
	// Root marked as "current" by ABroot
	current_root, err := getGRUBMarkedRoot("current")
	if err != nil {
		return err
	}

	// Root we're booted into
	currently_booted_root, err := getCurrentRootLabel()
	if err != nil {
		return err
	}

	var previous_root string
	if current_root == "a" {
		previous_root = "b"
	} else {
		previous_root = "a"
	}

	bold := cmdr.Bold.Sprint
	red := cmdr.Red
	green := cmdr.Green

	var currently_booted_root_colored string
	var rollback_kargs string
	if current_root == currently_booted_root {
		rollback_kargs, err = GetFutureKargs()
		if err != nil {
			return err
		}
		currently_booted_root_colored = red(strings.ToUpper(currently_booted_root))
	} else {
		rollback_kargs, err = GetCurrentKargs()
		if err != nil {
			return err
		}
		currently_booted_root_colored = green(strings.ToUpper(currently_booted_root))
	}

	message := fmt.Sprintf(`
You are currently in partition %s
Your "present" partition is %s

This command will make %s the present partition again. Any changes made to %s will be lost.
Continue?`,
		bold(currently_booted_root_colored),
		bold(red(strings.ToUpper(current_root))),
		bold(green(strings.ToUpper(previous_root))),
		bold(red(strings.ToUpper(current_root))),
	)

	confirmation, err := cmdr.Confirm.Show(message)
	if err != nil {
		return err
	}

	if confirmation {
		UpdateRootBoot(false, rollback_kargs)
		cmdr.Success.Println("Rollback complete. Reboot your system to apply changes.")
	}

	return nil
}
