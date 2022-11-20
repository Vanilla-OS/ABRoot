package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

func getRootDevice(state string) (string, error) {
	/*
	 * getRootDevice returns the device of requested root partition.
	 * Note that the present root partition is always the current one, while
	 * the future root partition is the next one. So, the future root partition
	 * is detected by checking for the next label, e.g. B if curent is A.
	 */
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

func getCurrentRootLabel() (string, error) {
	/*
	 * getCurrentRootLabel returns the label of the current root partition.
	 * It does so by checking the label of the current root partition.
	 */
	cmd := exec.Command("lsblk", "-o", "LABEL", "-n", "/")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func getRootUUID(state string) (string, error) {
	/*
	 * getRootUUID returns the UUID of requested root partition.
	 */
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

func getRootLabel(state string) (string, error) {
	/*
	 * getRootLabel returns the label of requested root partition.
	 */
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

func getRootFileSystem(state string) (string, error) {
	/*
	 * getRootFileSystem returns the filesystem of requested root partition.
	 */
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

func MountFutureRoot() error {
	/*
	 * MountFutureRoot mounts the future root partition to /partB.
	 */
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
		} else {
			if err := os.RemoveAll("/partFuture"); err != nil {
				return err
			}
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

func UnmountFutureRoot() error {
	/*
	 * UnmountFutureRoot unmounts the future root partition.
	 */
	if err := unix.Unmount("/partB", 0); err != nil {
		return err
	}

	return nil
}

func UpdateRootBoot(transacting bool) error {
	/*
	 * UpdateRootBoot updates the boot entries for the requested root partition.
	 * It does so by writing the new boot entries to 10_vanilla, setting the
	 * future root partition as the first entry, and then updating the boot.
	 * Note that 10_vanilla is written in both the present and future root
	 * partitions. If transacting is true, the future partition is not mounted
	 * at /partFuture, since it should already be there.
	 */
	presentLabel, err := GetPresentRootLabel()
	if err != nil {
		return err
	}

	futureLabel, err := GetFutureRootLabel()
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
		if err := MountFutureRoot(); err != nil {
			return err
		}
	}
	/* layout of /etc/grub.d/10_vanilla:
	#!/bin/sh
	exec tail -n +3 $0

	menuentry 'State A' {
			search --no-floppy --fs-uuid --set=root 437fd101-ef63-41d3-a177-aab12c090b9a
			linux   /vmlinuz-5.19.0-23-generic root=UUID=ead269ee-105c-4f7e-a707-c0600b677ee0 ro  quiet splash $vt_handoff
			initrd  /initrd.img-5.19.0-23-generic
	}

	menuentry 'State B' {
			search --no-floppy --fs-uuid --set=root 437fd101-ef63-41d3-a177-aab12c090b9a
			linux   /vmlinuz-5.19.0-23-generic root=UUID=eff11375-f669-436f-991d-0e3fc1955380 ro  quiet splash $vt_handoff
			initrd  /initrd.img-5.19.0-23-generic
	}
	*/
	bootHeader := "#!/bin/sh\nexec tail -n +3 $0"
	bootEntry := `menuentry 'State %s' {
	search --no-floppy --fs-uuid --set=root %s
	linux   /vmlinuz-5.19.0-23-generic
	initrd  /initrd.img-5.19.0-23-generic
}`
	bootEntryPresent := fmt.Sprintf(bootEntry, presentLabel, presentUUID)
	bootEntryFuture := fmt.Sprintf(bootEntry, futureLabel, futureUUID)
	bootFile := strings.Join([]string{bootHeader, bootEntryPresent, bootEntryFuture}, "\n")

	if err := os.WriteFile("/etc/grub.d/10_vanilla", []byte(bootFile), 0644); err != nil {
		return err
	}

	if err := os.WriteFile("/partFuture/etc/grub.d/10_vanilla", []byte(bootFile), 0644); err != nil {
		return err
	}

	// TODO: finish this
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
