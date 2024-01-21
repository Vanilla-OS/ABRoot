package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
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

// ABRootManager exposes methods to manage ABRoot partitions, this includes
// getting the present and future partitions, the boot partition, the init
// volume (when using LVM Thin-Provisioning), and the other partition. If you
// need to operate on an ABRoot partition, you should use this struct, each
// partition is a pointer to a Partition struct, which contains methods to
// operate on the partition itself
type ABRootManager struct {
	// Partitions is a list of partitions managed by ABRoot
	Partitions []ABRootPartition

	// VarPartition is the partition where /var is mounted
	VarPartition Partition
}

// ABRootPartition represents a partition managed by ABRoot
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
	PrintVerboseInfo("NewABRootManager", "running...")

	a := &ABRootManager{}
	a.GetPartitions()

	return a
}

// GetPartitions gets the root partitions from the current device
func (a *ABRootManager) GetPartitions() error {
	PrintVerboseInfo("ABRootManager.GetRootPartitions", "running...")

	diskM := NewDiskManager()
	rootLabels := []string{settings.Cnf.PartLabelA, settings.Cnf.PartLabelB}
	for _, label := range rootLabels {
		partition, err := diskM.GetPartitionByLabel(label)
		if err != nil {
			PrintVerboseErr("ABRootManager.GetRootPartitions", 0, err)
			return err
		}

		identifier, err := a.IdentifyPartition(partition)
		if err != nil {
			PrintVerboseErr("ABRootManager.GetRootPartitions", 1, err)
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
		PrintVerboseErr("ABRootManager.GetRootPartitions", 2, err)
		return err
	}
	a.VarPartition = partition

	PrintVerboseInfo("ABRootManager.GetRootPartitions", "successfully got root partitions")

	return nil
}

// IsCurrent checks if a partition is the current one
func (a *ABRootManager) IsCurrent(partition Partition) bool {
	PrintVerboseInfo("ABRootManager.IsCurrent", "running...")

	if partition.MountPoint == "/" {
		PrintVerboseInfo("ABRootManager.IsCurrent", "partition is current")
		return true
	}

	PrintVerboseInfo("ABRootManager.IsCurrent", "partition is not current")
	return false
}

// IdentifyPartition identifies a partition
func (a *ABRootManager) IdentifyPartition(partition Partition) (identifiedAs string, err error) {
	PrintVerboseInfo("ABRootManager.IdentifyPartition", "running...")

	if partition.Label == settings.Cnf.PartLabelA || partition.Label == settings.Cnf.PartLabelB {
		if partition.MountPoint == "/" {
			PrintVerboseInfo("ABRootManager.IdentifyPartition", "partition is present")
			return "present", nil
		}

		PrintVerboseInfo("ABRootManager.IdentifyPartition", "partition is future")
		return "future", nil
	}

	err = errors.New("partition is not managed by ABRoot")
	PrintVerboseErr("ABRootManager.IdentifyPartition", 0, err)
	return "", err
}

// GetPresent gets the present partition
func (a *ABRootManager) GetPresent() (partition ABRootPartition, err error) {
	PrintVerboseInfo("ABRootManager.GetPresent", "running...")

	for _, partition := range a.Partitions {
		if partition.IdentifiedAs == "present" {
			PrintVerboseInfo("ABRootManager.GetPresent", "successfully got present partition")
			return partition, nil
		}
	}

	err = errors.New("present partition not found")
	PrintVerboseErr("ABRootManager.GetPresent", 0, err)
	return ABRootPartition{}, err
}

// GetFuture gets the future partition
func (a *ABRootManager) GetFuture() (partition ABRootPartition, err error) {
	PrintVerboseInfo("ABRootManager.GetFuture", "running...")

	for _, partition := range a.Partitions {
		if partition.IdentifiedAs == "future" {
			PrintVerboseInfo("ABRootManager.GetFuture", "successfully got future partition")
			return partition, nil
		}
	}

	err = errors.New("future partition not found")
	PrintVerboseErr("ABRootManager.GetFuture", 0, err)
	return ABRootPartition{}, err
}

// GetOther gets the other partition
func (a *ABRootManager) GetOther() (partition ABRootPartition, err error) {
	PrintVerboseInfo("ABRootManager.GetOther", "running...")

	present, err := a.GetPresent()
	if err != nil {
		PrintVerboseErr("ABRootManager.GetOther", 0, err)
		return ABRootPartition{}, err
	}

	if present.Label == settings.Cnf.PartLabelA {
		PrintVerboseInfo("ABRootManager.GetOther", "successfully got other partition")
		return a.GetPartition(settings.Cnf.PartLabelB)
	}

	PrintVerboseInfo("ABRootManager.GetOther", "successfully got other partition")
	return a.GetPartition(settings.Cnf.PartLabelA)
}

// GetPartition gets a partition by label
func (a *ABRootManager) GetPartition(label string) (partition ABRootPartition, err error) {
	PrintVerboseInfo("ABRootManager.GetPartition", "running...")

	for _, partition := range a.Partitions {
		if partition.Label == label {
			PrintVerboseInfo("ABRootManager.GetPartition", "successfully got partition")
			return partition, nil
		}
	}

	err = errors.New("partition not found")
	PrintVerboseErr("ABRootManager.GetPartition", 0, err)
	return ABRootPartition{}, err
}

// GetBoot gets the boot partition from the current device
func (a *ABRootManager) GetBoot() (partition Partition, err error) {
	PrintVerboseInfo("ABRootManager.GetBoot", "running...")

	diskM := NewDiskManager()
	part, err := diskM.GetPartitionByLabel(settings.Cnf.PartLabelBoot)
	if err != nil {
		err = errors.New("boot partition not found")
		PrintVerboseErr("ABRootManager.GetBoot", 0, err)

		return Partition{}, err
	}

	PrintVerboseInfo("ABRootManager.GetBoot", "successfully got boot partition")
	return part, nil
}

// GetInit gets the init volume when using LVM Thin-Provisioning
func (a *ABRootManager) GetInit() (partition Partition, err error) {
	PrintVerboseInfo("ABRootManager.GetInit", "running...")

	// Make sure Thin-Provisioning is properly configured
	if !settings.Cnf.ThinProvisioning || settings.Cnf.ThinInitVolume == "" {
		return Partition{}, errors.New("ABRootManager.GetInit: error: system is not configured for thin-provisioning")
	}

	diskM := NewDiskManager()
	part, err := diskM.GetPartitionByLabel(settings.Cnf.ThinInitVolume)
	if err != nil {
		err = errors.New("init volume not found")
		PrintVerboseErr("ABRootManager.GetInit", 0, err)

		return Partition{}, err
	}

	PrintVerboseInfo("ABRootManager.GetInit", "successfully got init volume")
	return part, nil
}
