package core

import (
	"fmt"
	"os"
	"os/exec"
)

func getRootDevice(state string) (string, error) {
	/*
	 * getRootDevice returns the device of requested root partition.
	 * Note that the present root partition is always the current one, while
	 * the future root partition is the next one. So, the future root partition
	 * is detected by checking for the next label, e.g. B if dcurent is A.
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
