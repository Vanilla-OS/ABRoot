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
	Partitions   []ABRootPartition
	VarPartition Partition
}

// ABRootPartition represents an ABRoot partition
type ABRootPartition struct {
	Label        string // Matches `partLabelA` and `partLabelB` settings entries
	IdentifiedAs string // Either `present` or `future`
	Partition    Partition
	MountPoint   string
	MountOptions string
	Uuid         string
	FsType       string
	Current      bool
}

// NewABRootManager creates a new ABRootManager
func NewABRootManager() *ABRootManager {
	PrintVerbose("NewABRootManager: running...")

	a := &ABRootManager{}
	a.GetPartitions()

	return a
}

// GetPartitions gets the root partitions from the current device
func (a *ABRootManager) GetPartitions() error {
	PrintVerbose("ABRootManager.GetRootPartitions: running...")

	diskM := NewDiskManager()
	rootLabels := []string{settings.Cnf.PartLabelA, settings.Cnf.PartLabelB}
	for _, label := range rootLabels {
		partition, err := diskM.GetPartitionByLabel(label)
		if err != nil {
			PrintVerbose("ABRootManager.GetRootPartitions: error: %s", err)
			return err
		}

		identifier, err := a.IdentifyPartition(partition)
		if err != nil {
			PrintVerbose("ABRootManager.GetRootPartitions: error: %s", err)
			return err
		}

		isCurrent := a.IsCurrent(partition)
		a.Partitions = append(a.Partitions, ABRootPartition{
			Label:        partition.Label,
			IdentifiedAs: identifier,
			Partition:    partition,
			MountPoint:   partition.MountPoint,
			MountOptions: partition.MountOptions,
			Uuid:         partition.Uuid,
			FsType:       partition.FsType,
			Current:      isCurrent,
		})
	}

	partition, err := diskM.GetPartitionByLabel(settings.Cnf.PartLabelVar)
	if err != nil {
		PrintVerbose("ABRootManager.GetRootPartitions: error: %s", err)
		return err
	}
	a.VarPartition = partition

	PrintVerbose("ABRootManager.GetRootPartitions: successfully got root partitions")

	return nil
}

// IsCurrent checks if a partition is the current one
func (a *ABRootManager) IsCurrent(partition Partition) bool {
	PrintVerbose("ABRootManager.IsCurrent: running...")

	if partition.MountPoint == "/" {
		PrintVerbose("ABRootManager.IsCurrent: partition is current")
		return true
	}

	PrintVerbose("ABRootManager.IsCurrent: partition is not current")
	return false
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

// GetOther gets the other partition
func (a *ABRootManager) GetOther() (partition ABRootPartition, err error) {
	PrintVerbose("ABRootManager.GetOther: running...")

	present, err := a.GetPresent()
	if err != nil {
		PrintVerbose("ABRootManager.GetOther: error: %s", err)
		return ABRootPartition{}, err
	}

	if present.Label == settings.Cnf.PartLabelA {
		PrintVerbose("ABRootManager.GetOther: successfully got other partition")
		return a.GetPartition(settings.Cnf.PartLabelB)
	}

	PrintVerbose("ABRootManager.GetOther: successfully got other partition")
	return a.GetPartition(settings.Cnf.PartLabelA)
}

// GetPartition gets a partition by label
func (a *ABRootManager) GetPartition(label string) (partition ABRootPartition, err error) {
	PrintVerbose("ABRootManager.GetPartition: running...")

	for _, partition := range a.Partitions {
		if partition.Label == label {
			PrintVerbose("ABRootManager.GetPartition: successfully got partition")
			return partition, nil
		}
	}

	err = errors.New("partition not found")
	PrintVerbose("ABRootManager.GetPartition: error: %s", err)
	return ABRootPartition{}, err
}

// GetBoot gets the boot partition from the current device
func (a *ABRootManager) GetBoot() (partition Partition, err error) {
	PrintVerbose("ABRootManager.GetBoot: running...")

	diskM := NewDiskManager()
	part, err := diskM.GetPartitionByLabel(settings.Cnf.PartLabelBoot)
	if err != nil {
		err = errors.New("boot partition not found")
		PrintVerbose("ABRootManager.GetBoot: error: %s", err)

		return Partition{}, err
	}

	PrintVerbose("ABRootManager.GetBoot: successfully got boot partition")
	return part, nil
}

// GetInit gets the init volume when using LVM Thin-Provisioning
func (a *ABRootManager) GetInit() (partition Partition, err error) {
	PrintVerbose("ABRootManager.GetInit: running...")

	// Make sure Thin-Provisioning is properly configured
	if !settings.Cnf.ThinProvisioning || settings.Cnf.ThinInitVolume == "" {
		return Partition{}, errors.New("ABRootManager.GetInit: error: system is not configured for thin-provisioning")
	}

	diskM := NewDiskManager()
	part, err := diskM.GetPartitionByLabel(settings.Cnf.ThinInitVolume)
	if err != nil {
		err = errors.New("init volume not found")
		PrintVerbose("ABRootManager.GetInit: error: %s", err)

		return Partition{}, err
	}

	PrintVerbose("ABRootManager.GetBoot: successfully got init volume")
	return part, nil
}
