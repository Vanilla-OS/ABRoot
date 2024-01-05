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
	"strings"
	"syscall"
)

// DiskManager exposes functions to interact with the system's disks
// and partitions (e.g. mount, unmount, get partitions, etc.)
type DiskManager struct{}

// Partition represents either a standard partition or a device-mapper
// partition, such as an LVM volume
type Partition struct {
	Label        string
	MountPoint   string
	MountOptions string
	Uuid         string
	FsType       string

	// If standard partition, Device will be the partition's name (e.g. sda1, nvme0n1p1).
	// If LUKS-encrypted or LVM volume, Device will be the name in device-mapper.
	Device string

	// If the partition is LUKS-encrypted or an LVM volume, the logical volume
	// opened in /dev/mapper will be a child of the physical partition in /dev.
	// Otherwise, the partition will be a direct child of the block device, and
	// Parent will be nil.
	//
	// The same logic applies for encrypted LVM volumes. When this is the case,
	// the filesystem hirearchy is as follows:
	//
	//         NAME               FSTYPE
	//   -- sda1                LVM2_member
	//    |-- myVG-myLV         crypto_LUKS
	//      |-- luks-volume     btrfs
	//
	// In this case, the parent of "luks-volume" is "myVG-myLV", which,
	// in turn, has "sda1" as parent. Since "sda1" is a physical partition,
	// its parent is nil.
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

// NewDiskManager creates and returns a pointer to a new DiskManager instance
// from which you can interact with the system's disks and partitions
func NewDiskManager() *DiskManager {
	return &DiskManager{}
}

// GetPartitionByLabel finds a partition by searching for its label.
// If no partition can be found with the given label, returns error.
func (d *DiskManager) GetPartitionByLabel(label string) (Partition, error) {
	PrintVerboseInfo("DiskManager.GetPartitionByLabel", "retrieving partitions")

	partitions, err := d.getPartitions("")
	if err != nil {
		PrintVerboseErr("DiskManager.GetPartitionByLabel", 0, err)
		return Partition{}, err
	}

	for _, part := range partitions {
		if part.Label == label {
			PrintVerboseInfo("DiskManager.GetPartitionByLabel", "Partition with UUID", part.Uuid, "has label", label)
			return part, nil
		}
	}

	errMsg := fmt.Errorf("could not find partition with label %s", label)
	PrintVerboseErr("DiskManager.GetPartitionByLabel", 1, errMsg)
	return Partition{}, errMsg
}

// iterChildren iterates through the children of a device or partition
// recursively
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

		currentPartitions := len(*result)
		iterChildren(&child.Children, result)
		detectedPartitions := len(*result) - currentPartitions

		// Populate children's reference to parent
		for i := currentPartitions; i < len(*result); i++ {
			if (*result)[i].Parent == nil {
				(*result)[i].Parent = &(*result)[len(*result)-detectedPartitions-1]
			}
		}
	}
}

// getPartitions gets a disk's partitions. If device is an empty string, gets
// all partitions from all disks
func (d *DiskManager) getPartitions(device string) ([]Partition, error) {
	PrintVerboseInfo("DiskManager.getPartitions", "running...")

	output, err := exec.Command("lsblk", "-J", "-o", "NAME,FSTYPE,LABEL,MOUNTPOINT,UUID").Output()
	if err != nil {
		PrintVerboseErr("DiskManager.getPartitions", 0, err)
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
		PrintVerboseErr("DiskManager.getPartitions", 1, err)
		return nil, err
	}

	var result []Partition
	for _, blockDevice := range partitions.BlockDevices {
		if device != "" && blockDevice.Name != device {
			continue
		}

		iterChildren(&blockDevice.Children, &result)
	}

	PrintVerboseInfo("DiskManager.getPartitions", "successfully got partitions for disk", device)

	return result, nil
}

// Mount mounts a partition to a directory, returning an error if any occurs
func (p *Partition) Mount(destination string) error {
	PrintVerboseInfo("Partition.Mount", "running...")

	if _, err := os.Stat(destination); os.IsNotExist(err) {
		if err := os.MkdirAll(destination, 0755); err != nil {
			PrintVerboseErr("Partition.Mount", 0, err)
			return err
		}
	}

	devicePath := "/dev/"
	if p.IsDevMapper() {
		devicePath += "mapper/"
	}
	devicePath += p.Device

	err := syscall.Mount(devicePath, destination, p.FsType, 0, "")
	if err != nil {
		PrintVerboseErr("Partition.Mount", 1, err)
		return err
	}

	p.MountPoint = destination
	PrintVerboseInfo("Partition.Mount", "successfully mounted", devicePath, "to", destination)
	return nil
}

// Unmount unmounts a partition
func (p *Partition) Unmount() error {
	PrintVerboseInfo("Partition.Unmount", "running...")

	if p.MountPoint == "" {
		PrintVerboseErr("Partition.Unmount", 0, errors.New("no mount point"))
		return errors.New("no mount point")
	}

	err := syscall.Unmount(p.MountPoint, 0)
	if err != nil {
		PrintVerboseErr("Partition.Unmount", 1, err)
		return err
	}

	PrintVerboseInfo("Partition.Unmount", "successfully unmounted", p.MountPoint)
	p.MountPoint = ""

	return nil
}

// Returns whether the partition is a device-mapper virtual partition
func (p *Partition) IsDevMapper() bool {
	return p.Parent != nil
}

// IsEncrypted returns whether the partition is encrypted
func (p *Partition) IsEncrypted() bool {
	return strings.HasPrefix(p.Device, "luks-")
}
