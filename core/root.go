package core

import (
	"errors"
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
	MountPoint   string
	MountOptions string
	Uuid         string
	FsType       string
}

// NewABRootManager creates a new ABRootManager
func NewABRootManager() *ABRootManager {
	a := &ABRootManager{}
	a.GetRootPartitions()
	return a
}

// GetRootPartitions gets the root partitions from the current device
func (a *ABRootManager) GetRootPartitions() error {
	diskM := NewDiskManager()
	disk, err := diskM.GetCurrentDisk()
	if err != nil {
		return err
	}

	for _, partition := range disk.Partitions {
		if partition.Label == "a" || partition.Label == "b" {
			identifier, err := a.IdentifyPartition(partition)
			if err != nil {
				return err
			}

			a.Partitions = append(a.Partitions, ABRootPartition{
				Label:        partition.Label,
				IdentifiedAs: identifier,
				Device:       disk.Device,
				MountPoint:   partition.MountPoint,
				MountOptions: partition.MountOptions,
				Uuid:         partition.Uuid,
				FsType:       partition.FsType,
			})
		}
	}

	return nil
}

// IdentifyPartition identifies a partition
func (a *ABRootManager) IdentifyPartition(partition Partition) (identifiedAs string, err error) {
	if partition.Label == "a" || partition.Label == "b" {
		if partition.MountPoint == "/" {
			return "present", nil
		}

		return "future", nil
	}

	return "", errors.New("partition is not managed by ABRoot")
}

// GetPresent gets the present partition
func (a *ABRootManager) GetPresent() (partition ABRootPartition, err error) {
	for _, partition := range a.Partitions {
		if partition.IdentifiedAs == "present" {
			return partition, nil
		}
	}

	return ABRootPartition{}, errors.New("present partition not found")
}

// GetFuture gets the future partition
func (a *ABRootManager) GetFuture() (partition ABRootPartition, err error) {
	for _, partition := range a.Partitions {
		if partition.IdentifiedAs == "future" {
			return partition, nil
		}
	}

	return ABRootPartition{}, errors.New("future partition not found")
}
