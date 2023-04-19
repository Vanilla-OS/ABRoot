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
	"errors"

	"github.com/vanilla-os/abroot/settings"
)

// ABRootManager represents the ABRoot manager
type ABRootManager struct {
	Partitions []ABRootPartition
}

// ABRootPartition represents an ABRoot partition
type ABRootPartition struct {
	Label        string // a,b
	IdentifiedAs string // present,future
	Device       string
	Partition    Partition
	MountPoint   string
	MountOptions string
	Uuid         string
	FsType       string
}

// NewABRootManager creates a new ABRootManager
func NewABRootManager() *ABRootManager {
	PrintVerbose("NewABRootManager: running...")

	a := &ABRootManager{}
	a.GetRootPartitions()

	return a
}

// GetRootPartitions gets the root partitions from the current device
func (a *ABRootManager) GetRootPartitions() error {
	PrintVerbose("ABRootManager.GetRootPartitions: running...")

	diskM := NewDiskManager()
	disk, err := diskM.GetCurrentDisk()
	if err != nil {
		PrintVerbose("ABRootManager.GetRootPartitions: error: %s", err)
		return err
	}

	for _, partition := range disk.Partitions {
		if partition.Label == settings.Cnf.PartLabelA || partition.Label == settings.Cnf.PartLabelB {
			identifier, err := a.IdentifyPartition(partition)
			if err != nil {
				PrintVerbose("ABRootManager.GetRootPartitions: error: %s", err)
				return err
			}

			a.Partitions = append(a.Partitions, ABRootPartition{
				Label:        partition.Label,
				IdentifiedAs: identifier,
				Device:       disk.Device,
				Partition:    partition,
				MountPoint:   partition.MountPoint,
				MountOptions: partition.MountOptions,
				Uuid:         partition.Uuid,
				FsType:       partition.FsType,
			})
		}
	}

	PrintVerbose("ABRootManager.GetRootPartitions: successfully got root partitions")

	return nil
}

// IdentifyPartition identifies a partition
func (a *ABRootManager) IdentifyPartition(partition Partition) (identifiedAs string, err error) {
	PrintVerbose("ABRootManager.IdentifyPartition: running...")

	if partition.Label == settings.Cnf.PartLabelA || partition.Label == settings.Cnf.PartLabelB {
		if partition.MountPoint == "/" {
			PrintVerbose("ABRootManager.IdentifyPartition: partition is present")
			return "present", nil
		}

		PrintVerbose("ABRootManager.IdentifyPartition: partition is future")
		return "future", nil
	}

	err = errors.New("partition is not managed by ABRoot")
	PrintVerbose("ABRootManager.IdentifyPartition: error: %s", err)
	return "", err
}

// GetPresent gets the present partition
func (a *ABRootManager) GetPresent() (partition ABRootPartition, err error) {
	PrintVerbose("ABRootManager.GetPresent: running...")

	for _, partition := range a.Partitions {
		if partition.IdentifiedAs == "present" {
			PrintVerbose("ABRootManager.GetPresent: successfully got present partition")
			return partition, nil
		}
	}

	err = errors.New("present partition not found")
	PrintVerbose("ABRootManager.GetPresent: error: %s", err)
	return ABRootPartition{}, err
}

// GetFuture gets the future partition
func (a *ABRootManager) GetFuture() (partition ABRootPartition, err error) {
	PrintVerbose("ABRootManager.GetFuture: running...")

	for _, partition := range a.Partitions {
		if partition.IdentifiedAs == "future" {
			PrintVerbose("ABRootManager.GetFuture: successfully got future partition")
			return partition, nil
		}
	}

	err = errors.New("future partition not found")
	PrintVerbose("ABRootManager.GetFuture: error: %s", err)
	return ABRootPartition{}, err
}

// GetBoot gets the boot partition from the current device
func (a *ABRootManager) GetBoot() (partition Partition, err error) {
	PrintVerbose("ABRootManager.GetBoot: running...")

	diskM := NewDiskManager()
	disk, err := diskM.GetCurrentDisk()
	if err != nil {
		PrintVerbose("ABRootManager.GetBoot: error: %s", err)
		return Partition{}, err
	}

	for _, partition := range disk.Partitions {
		if partition.Label == settings.Cnf.PartLabelBoot {
			PrintVerbose("ABRootManager.GetBoot: successfully got boot partition")
			return partition, nil
		}
	}

	err = errors.New("boot partition not found")
	PrintVerbose("ABRootManager.GetBoot: error: %s", err)
	return Partition{}, err
}
