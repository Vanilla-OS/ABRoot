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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
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
	Device       string // e.g. sda1, nvme0n1p1

	// If the partition is LUKS-encrypted, the logical volume opened in /dev/mapper/
	// will be a child of the physical partiiton in /dev/.
	// Otherwise, the partition will be a direct child of the block device, and
	// therefore Parent will be nil.
	Parent *Partition
}

// The children a block device or partition may have
type Children struct {
	MountPoint   string     `json:"mountpoint"`
	FsType       string     `json:"fstype"`
	Label        string     `json:"label"`
	Uuid         string     `json:"uuid"`
	LogicalName  string     `json:"name"`
	Size         string     `json:"size"`
	MountOptions string     `json:"mountopts"`
	Children     []Children `json:"children"`
}

// NewDiskManager creates a new DiskManager
func NewDiskManager() *DiskManager {
	return &DiskManager{}
}

// GetDisk gets a disk by device
func (d *DiskManager) GetDisk(device string) (Disk, error) {
	PrintVerbose("DiskManager.GetDisk: running...")
	partitions, err := d.getPartitions(device)
	if err != nil {
		PrintVerbose("DiskManager.GetDisk:err: %s", err)
		return Disk{}, err
	}

	PrintVerbose("DiskManager.GetDisk: successfully got disk %s", device)

	return Disk{
		Device:     device,
		Partitions: partitions,
	}, nil
}

// GetDiskByPartition gets a disk by partition
func (d *DiskManager) GetDiskByPartition(partition string) (Disk, error) {
	PrintVerbose("DiskManager.GetDiskByPartition: running...")

	output, err := exec.Command("lsblk", "-n", "-o", "PKNAME", "/dev/"+partition).Output()
	if err != nil {
		PrintVerbose("DiskManager.GetDiskByPartition:err: %s", err)
		return Disk{}, err
	}

	device := strings.TrimSpace(string(output))

	PrintVerbose("DiskManager.GetDiskByPartition: successfully got disk %s", device)

	return d.GetDisk(device)
}

// GetCurrentDisk gets the current disk
func (d *DiskManager) GetCurrentDisk() (Disk, error) {
	PrintVerbose("DiskManager.GetCurrentDisk: running...")

	// we need to evaluate symlinks to get the real root path
	// in case of weird setups
	root, err := filepath.EvalSymlinks("/")
	if err != nil {
		PrintVerbose("DiskManager.GetCurrentDisk:err(2): %s", err)
		return Disk{}, err
	}

	output, err := exec.Command("df", "-P", root).Output()
	if err != nil {
		PrintVerbose("DiskManager.GetCurrentDisk:err(3): %s", err)
		return Disk{}, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		err := errors.New("could not determine device name for " + root)
		PrintVerbose("DiskManager.GetCurrentDisk:err(4): %s", err)
		return Disk{}, err
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		err := errors.New("could not determine device name for " + root)
		PrintVerbose("DiskManager.GetCurrentDisk:err(5): %s", err)
		return Disk{}, err
	}

	device := filepath.Base(fields[0])

	PrintVerbose("DiskManager.GetCurrentDisk: successfully got disk %s", device)

	return d.GetDiskByPartition(device)
}

// iterChildren iterates through the children of a device or partition recursively
func iterChildren(childs *[]Children, result *[]Partition) {
	for _, child := range *childs {
		*result = append(*result, Partition{
			Label:        child.Label,
			MountPoint:   child.MountPoint,
			MountOptions: child.MountOptions,
			Uuid:         child.Uuid,
			FsType:       child.FsType,
			Device:       child.LogicalName,
		})

		detectedPartitions := len(*result)

		iterChildren(&child.Children, result)

		if detectedPartitions == 0 {
			return
		}

		// Populate children's reference to parent
		for i := 0; i < len(*result)-detectedPartitions; i++ {
			(*result)[detectedPartitions+i].Parent = &(*result)[detectedPartitions-1]
		}
	}
}

// getPartitions gets a disk's partitions
func (d *DiskManager) getPartitions(device string) ([]Partition, error) {
	PrintVerbose("DiskManager.getPartitions: running...")

	output, err := exec.Command("lsblk", "-J", "-o", "NAME,FSTYPE,LABEL,MOUNTPOINT,UUID").Output()
	if err != nil {
		PrintVerbose("DiskManager.getPartitions:err: %s", err)
		return nil, err
	}

	var partitions struct {
		BlockDevices []struct {
			Name     string     `json:"name"`
			Type     string     `json:"type"`
			Children []Children `json:"children"`
		} `json:"blockdevices"`
	}

	if err := json.Unmarshal(output, &partitions); err != nil {
		PrintVerbose("DiskManager.getPartitions:err(2): %s", err)
		return nil, err
	}

	var result []Partition
	for _, blockDevice := range partitions.BlockDevices {
		if blockDevice.Name != device {
			continue
		}

		iterChildren(&blockDevice.Children, &result)
	}

	PrintVerbose("DiskManager.getPartitions: successfully got partitions for disk %s", device)

	return result, nil
}

// Mount mounts a partition to a directory
func (p *Partition) Mount(destination string) error {
	PrintVerbose("Partition.Mount: running...")

	if _, err := os.Stat(destination); os.IsNotExist(err) {
		if err := os.MkdirAll(destination, 0755); err != nil {
			PrintVerbose("Partition.Mount: error: %s", err)
			return err
		}
	}

	err := syscall.Mount(fmt.Sprintf("/dev/%s", p.Device), destination, p.FsType, 0, "")
	if err != nil {
		PrintVerbose("Partition.Mount: error(2): %s", err)
		return err
	}

	p.MountPoint = destination
	PrintVerbose("Partition.Mount: successfully mounted partition")
	return nil
}

// Unmount unmounts a partition
func (p *Partition) Unmount() error {
	PrintVerbose("Partition.Unmount: running...")

	if p.MountPoint == "" {
		PrintVerbose("Partition.Unmount: error: no mount point")
		return errors.New("no mount point")
	}

	err := syscall.Unmount(p.MountPoint, 0)
	if err != nil {
		PrintVerbose("Partition.Unmount: error(2): %s", err)
		return err
	}

	p.MountPoint = ""
	PrintVerbose("Partition.Unmount: successfully unmounted partition")
	return nil
}
