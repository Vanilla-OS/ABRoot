package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DiskManager represents a disk
type DiskManager struct{}

// Disk represents a disk
type Disk struct {
	Device     string
	Partitions []Partition
}

// Partition represents a standard partition
type Partition struct {
	Label        string
	MountPoint   string
	MountOptions string
	Uuid         string
	FsType       string
}

// NewDiskManager creates a new DiskManager
func NewDiskManager() *DiskManager {
	return &DiskManager{}
}

// GetDisk gets a disk by device
func (d *DiskManager) GetDisk(device string) (Disk, error) {
	partitions, err := d.getPartitions(device)
	if err != nil {
		return Disk{}, err
	}

	return Disk{
		Device:     device,
		Partitions: partitions,
	}, nil
}

// GetDiskByPartition gets a disk by partition
func (d *DiskManager) GetDiskByPartition(partition string) (Disk, error) {
	output, err := exec.Command("lsblk", "-n", "-o", "PKNAME", "/dev/"+partition).Output()
	if err != nil {
		return Disk{}, err
	}

	device := strings.TrimSpace(string(output))
	return d.GetDisk(device)
}

// GetCurrentDisk gets the current disk
func (d *DiskManager) GetCurrentDisk() (Disk, error) {
	root, err := os.Getwd()
	if err != nil {
		return Disk{}, err
	}

	// we need to evaluate symlinks to get the real root path
	// in case of weird setups
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return Disk{}, err
	}

	output, err := exec.Command("df", "-P", root).Output()
	if err != nil {
		return Disk{}, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return Disk{}, fmt.Errorf("could not determine device name for %s", root)
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		return Disk{}, fmt.Errorf("could not determine device name for %s", root)
	}

	device := filepath.Base(fields[0])
	return d.GetDiskByPartition(device)
}

// getPartitions gets a disk's partitions
func (d *DiskManager) getPartitions(device string) ([]Partition, error) {
	output, err := exec.Command("lsblk", "-J", "-o", "NAME,FSTYPE,LABEL,MOUNTPOINT,UUID").Output()
	if err != nil {
		return nil, err
	}

	var partitions struct {
		BlockDevices []struct {
			Name     string `json:"name"`
			Type     string `json:"type"`
			Children []struct {
				MountPoint   string `json:"mountpoint"`
				FsType       string `json:"fstype"`
				Label        string `json:"label"`
				Uuid         string `json:"uuid"`
				LogicalName  string `json:"name"`
				Size         string `json:"size"`
				MountOptions string `json:"mountopts"`
			} `json:"children"`
		} `json:"blockdevices"`
	}

	if err := json.Unmarshal(output, &partitions); err != nil {
		return nil, err
	}

	var result []Partition
	for _, blockDevice := range partitions.BlockDevices {
		if blockDevice.Name != device {
			continue
		}

		for _, child := range blockDevice.Children {
			result = append(result, Partition{
				Label:        child.Label,
				MountPoint:   child.MountPoint,
				MountOptions: child.MountOptions,
				Uuid:         child.Uuid,
				FsType:       child.FsType,
			})
		}
	}

	return result, nil
}
