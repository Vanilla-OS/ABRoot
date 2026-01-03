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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	EtcBuilder "github.com/linux-immutability-tools/EtcBuilder/cmd"
	digest "github.com/opencontainers/go-digest"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/sdk/pkg/v1/goodies"
)

// An ABSystem allows to perform system operations such as upgrades,
// package changes and rollback on an ABRoot-compliant system.
type ABSystem struct {
	// Checks contains an instance of Checks which allows to perform
	// compatibility checks on the system such as filesystem compatibility,
	// connectivity and root check.
	Checks *Checks

	// RootM contains an instance of the ABRootManager which allows to
	// manage the ABRoot partition scheme.
	RootM *ABRootManager

	// CurImage contains an instance of ABImage which represents the current
	// image used by the system (abimage.abr).
	CurImage *ABImage
}

// Supported ABSystemOperation types
const (
	// ABSystem operations
	UPGRADE           = "upgrade"
	FORCE_UPGRADE     = "force-upgrade"
	DRY_RUN_UPGRADE   = "dry-run-upgrade"
	APPLY             = "package-apply"
	DRY_RUN_APPLY     = "dry-run-package-apply"
	INITRAMFS         = "initramfs"
	DRY_RUN_INITRAMFS = "dry-run-initramfs"
)

// ABSystem rollback response
const (
	// can rollback
	ROLLBACK_RES_YES = "rollback-yes"

	// can't rollback
	ROLLBACK_RES_NO = "rollback-no"

	ROLLBACK_UNNECESSARY = "rollback-unnecessary"
	ROLLBACK_SUCCESS     = "rollback-success"
	ROLLBACK_FAILED      = "rollback-failed"
)

// ABSystemOperation represents a system operation to be performed by the
// ABSystem, must be given as a parameter to the RunOperation function.
type ABSystemOperation string

// ABRollbackResponse represents the response of a rollback operation
type ABRollbackResponse string

// Common variables and errors used by the ABSystem
var (
	operationLockFile     string = filepath.Join("/run", "abroot", "operation.lock")
	finalizingFile        string = filepath.Join("/run", "abroot", "finalizing")
	userStopFile          string = filepath.Join("/run", "abroot", "userStop")
	finishedOperationFile string = filepath.Join("/run", "abroot", "finished")

	// Errors
	ErrNoUpdate        error = errors.New("no update available")
	ErrUserStopped     error = errors.New("operation stopped per user request")
	ErrOperationLocked error = errors.New("another operation is currently running")
)

// NewABSystem initializes a new ABSystem, which contains all the functions
// to perform system operations such as upgrades, package changes and rollback.
// It returns a pointer to the initialized ABSystem and an error, if any.
func NewABSystem() (*ABSystem, error) {
	PrintVerboseInfo("NewABSystem: running...")

	i, err := NewABImageFromRoot()
	if err != nil {
		PrintVerboseErr("NewABSystem", 0, err)
		return nil, err
	}

	c := NewChecks()
	rm := NewABRootManager()

	return &ABSystem{
		Checks:   c,
		RootM:    rm,
		CurImage: i,
	}, nil
}

// CheckAll performs all checks from the Checks struct
func (s *ABSystem) CheckAll() error {
	PrintVerboseInfo("ABSystem.CheckAll", "running...")

	err := s.Checks.PerformAllChecks()
	if err != nil {
		PrintVerboseErr("ABSystem.CheckAll", 0, err)
		return err
	}

	PrintVerboseInfo("ABSystem.CheckAll", "all checks passed")
	return nil
}

// CheckUpdate checks if there is an update available
func (s *ABSystem) CheckUpdate() (digest.Digest, bool, error) {
	PrintVerboseInfo("ABSystem.CheckUpdate", "running...")
	return HasUpdate(s.CurImage.Digest)
}

func (s *ABSystem) Rebase(name string, dryRun bool) error {

	if name == "" {
		return fmt.Errorf("no image provided")
	}

	if strings.Contains(name, ".") {
		registrySplit := strings.SplitN(name, "/", 2)
		settings.Cnf.Registry = registrySplit[0]
		name = registrySplit[1]
	}
	nameTagSplit := strings.Split(name, ":")
	name = nameTagSplit[0]
	if len(nameTagSplit) < 1 {
		fmt.Errorf("No tag provided")
	}
	settings.Cnf.Tag = nameTagSplit[1]
	if name != "" {
		settings.Cnf.Name = name
	}

	_, _, err := s.CheckUpdate()
	if err != nil {
		return err
	}

	if !dryRun {
		err := settings.WriteConfigToFile(settings.CnfPathAdmin)
		if err != nil {
			return err
		}
	}

	return nil

}

// RunOperation executes a root-switching operation from the options below:
//
//	UPGRADE:
//		Upgrades to a new image, if available,
//	FORCE_UPGRADE:
//		Forces the upgrade operation, even if no new image is available,
//	APPLY:
//		Applies package changes, and updates the system if an update is available.
//	INITRAMFS:
//		Updates the initramfs for the future root, but doesn't update the system.
func (s *ABSystem) RunOperation(operation ABSystemOperation, deleteBeforeCopy bool) error {
	PrintVerboseInfo("ABSystem.RunOperation", "starting", operation)

	cq := goodies.NewCleanupQueue()
	defer cq.Run()

	// Stage 0: Check if upgrade is possible
	// -------------------------------------
	PrintVerboseSimple("[Stage 0] -------- ABSystemRunOperation")

	if s.finishedFileExists() {
		PrintVerboseWarn("ABSystem.RunOperation", 0, "reboot required")
		return errors.New("another operation finished successfully, a reboot is required")
	}

	err := s.LockOperation()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 0.1, "could not create lock file:", err)
		return fmt.Errorf("could not create lock file: %w", err)
	}

	cq.Add(func(args ...interface{}) error {
		return s.UnlockOperation()
	}, nil, 100, &goodies.NoErrorHandler{}, false)

	// removes the finalizing file if it exists
	err = s.removeFinalizingFile()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 0.2, err)
		return err
	}

	// Stage 1: Check if there is an update available
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 1] -------- ABSystemRunOperation")

	if UserStopRequested() {
		err = ErrUserStopped
		PrintVerboseErr("ABSystem.RunOperation", 2, err)
		return err
	}

	var imageDigest digest.Digest
	if operation != INITRAMFS {
		var res bool
		imageDigest, res, err = s.CheckUpdate()
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 1, err)
			return err
		}
		if !res {
			if operation != FORCE_UPGRADE && operation != APPLY && operation != DRY_RUN_APPLY {
				PrintVerboseErr("ABSystem.RunOperation", 1.1, err)
				return ErrNoUpdate
			}
			imageDigest = s.CurImage.Digest
			if operation == FORCE_UPGRADE {
				PrintVerboseWarn("ABSystem.RunOperation", 1.2, "No update available but --force is set. Proceeding...")
			}
		}
	} else {
		imageDigest = s.CurImage.Digest
	}

	// Stage 2: Get the present root, future root and boot partitions,
	// 			mount future to /part-future and clean up
	// 			old /part-future/new directory (it is
	// 			possible that last transaction was interrupted
	// 			before the clean up was done).
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 2] -------- ABSystemRunOperation")

	if UserStopRequested() {
		err = ErrUserStopped
		PrintVerboseErr("ABSystem.RunOperation", 2, err)
		return err
	}

	partPresent, err := s.RootM.GetPresent()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 2.05, err)
		return err
	}

	partFuture, err := s.RootM.GetFuture()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 2.1, err)
		return err
	}

	partBoot, err := s.RootM.GetBoot()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 2.2, err)
		return err
	}

	partFuture.Partition.Unmount() // just in case
	partBoot.Unmount()

	futureRoot := "/part-future"
	err = partFuture.Partition.Mount(futureRoot)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 2.3, err)
		return err
	}

	cq.Add(func(args ...interface{}) error {
		return partFuture.Partition.Unmount()
	}, nil, 90, &goodies.NoErrorHandler{}, false)

	// Stage 3: Make a imageRecipe with user packages
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 3] -------- ABSystemRunOperation")

	if UserStopRequested() {
		err = ErrUserStopped
		PrintVerboseErr("ABSystem.RunOperation", 2, err)
		return err
	}

	// Stage 3.1: Delete old images
	switch operation {
	case DRY_RUN_UPGRADE, DRY_RUN_APPLY, DRY_RUN_INITRAMFS:
	default:
		err = DeleteAllButLatestImage()
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 3.1, err)
			return err
		}
	}

	labels := map[string]string{
		"maintainer":  "'Generated by ABRoot'",
		"ABRoot.root": partFuture.Label,
	}
	args := map[string]string{}
	pkgM, err := NewPackageManager(false)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 3.2, err)
		return err
	}

	pkgsFinal, err := pkgM.GetFinalCmd()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 3.3, err)
	}
	if pkgsFinal == "" {
		pkgsFinal = "true"
	}
	content := `RUN ` + pkgsFinal

	var imageName string
	switch operation {
	case INITRAMFS, DRY_RUN_INITRAMFS:
		imageName, err = RetrieveImageForRoot(partPresent.Label)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 3.4, err)
			return err
		}
		// Handle case where an image for the current root may not exist
		// in storage
		if imageName == "" {
			imageName = settings.GetFullImageNameWithTag()
		}
	default:
		imageName = settings.GetFullImageName()
		imageName += "@" + imageDigest.String()
		labels["ABRoot.BaseImageDigest"] = imageDigest.String()
	}

	imageRecipe := NewImageRecipe(
		imageName,
		labels,
		args,
		content,
	)

	// Stage 4: Extract the rootfs
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 4] -------- ABSystemRunOperation")

	if UserStopRequested() {
		err = ErrUserStopped
		PrintVerboseErr("ABSystem.RunOperation", 2, err)
		return err
	}

	if deleteBeforeCopy || os.Getenv("ABROOT_DELETE_BEFORE_COPY") != "" {
		PrintVerboseInfo("ABSystemRunOperation", "Deleting future system, this will render the future root temporarily unavailable")
		err := ClearDirectory(partFuture.Partition.MountPoint, nil)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 4, err)
			return err
		}
	}

	abrootTrans := filepath.Join(futureRoot, "abroot-trans")
	err = OciExportRootFs(
		"abroot-"+uuid.New().String(),
		imageRecipe,
		abrootTrans,
		futureRoot,
	)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 4.2, err)
		return err
	}

	// Stage 4.1: Delete old images
	switch operation {
	case DRY_RUN_UPGRADE, DRY_RUN_APPLY, DRY_RUN_INITRAMFS:
	default:
		err = DeleteAllButLatestImage()
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 3.1, err)
			return err
		}
	}

	// Stage 4.2: Repair root integrity
	err = RepairRootIntegrity(futureRoot)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 2.4, err)
		return err
	}

	// Stage 5: Write new abimage.abr and config to future/
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 5] -------- ABSystemRunOperation")

	if UserStopRequested() {
		err = ErrUserStopped
		PrintVerboseErr("ABSystem.RunOperation", 2, err)
		return err
	}

	abimage, err := NewABImage(imageDigest, settings.GetFullImageNameWithTag())
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 5.1, err)
		return err
	}

	err = abimage.WriteTo(futureRoot)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 5.2, err)
		return err
	}

	varParent := s.RootM.VarPartition.Parent
	if varParent != nil && varParent.IsEncrypted() {
		device := varParent.Device
		if varParent.IsDevMapper() {
			device = "/dev/mapper/" + device
		} else {
			device = "/dev/" + device
		}

		settings.Cnf.PartCryptVar = device
	}

	err = settings.WriteConfigToFile(filepath.Join(futureRoot, "/usr/share/abroot/abroot.json"))
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 5.25, err)
		return err
	}

	err = pkgM.WriteSummaryToRoot(futureRoot)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 5.26, err)
		return err
	}

	if UserStopRequested() {
		err = ErrUserStopped
		PrintVerboseErr("ABSystem.RunOperation", 2, err)
		return err
	}

	err = s.createFinalizingFile()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 5.3, err)
		return err
	}

	// Stage (dry): If dry-run, exit here before writing to disk
	// ------------------------------------------------
	switch operation {
	case DRY_RUN_UPGRADE, DRY_RUN_APPLY, DRY_RUN_INITRAMFS:
		PrintVerboseInfo("ABSystem.RunOperation", "dry-run completed")
		return nil
	}

	// Stage 6: Update the bootloader
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 6] -------- ABSystemRunOperation")

	chroot, err := NewChroot(
		futureRoot,
		partFuture.Partition.Uuid,
		partFuture.Partition.Device,
		true,
		filepath.Join("/var/lib/abroot/etc", partPresent.Label),
	)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 7.02, err)
		return err
	}

	generatedGrubConfigPath := "/boot/grub/grub.cfg"

	grubCommand := fmt.Sprintf(settings.Cnf.UpdateGrubCmd, generatedGrubConfigPath)
	err = chroot.Execute(grubCommand)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 7.1, err)
		return err
	}

	err = chroot.Execute(settings.Cnf.UpdateInitramfsCmd) // ensure initramfs is updated
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 7.2, err)
		return err
	}

	err = chroot.Close()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 7.25, err)
		return err
	}

	newKernelVer := getKernelVersion(filepath.Join(futureRoot, "boot"))
	if newKernelVer == "" {
		err := errors.New("could not get kernel version")
		PrintVerboseErr("ABSystem.RunOperation", 7.26, err)
		return err
	}

	var rootUuid string
	// If Thin-Provisioning set, mount init partition and move linux and initrd
	// images to it.
	var initMountpoint string
	if settings.Cnf.ThinProvisioning {
		initPartition, err := s.RootM.GetInit()
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 7.3, err)
			return err
		}

		initMountpoint = filepath.Join(futureRoot, "boot", "init")
		err = initPartition.Mount(initMountpoint)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 7.4, err)
			return err
		}

		cq.Add(func(args ...interface{}) error {
			return initPartition.Unmount()
		}, nil, 80, &goodies.NoErrorHandler{}, false)

		futureInitDir := filepath.Join(initMountpoint, partFuture.Label)

		err = os.RemoveAll(futureInitDir)
		if err != nil {
			PrintVerboseWarn("ABSystem.RunOperation", 7.44)
		}
		err = os.MkdirAll(futureInitDir, 0o755)
		if err != nil {
			PrintVerboseWarn("ABSystem.RunOperation", 7.47, err)
		}

		err = MoveFile(
			filepath.Join(futureRoot, "boot", "vmlinuz-"+newKernelVer),
			filepath.Join(futureInitDir, "vmlinuz-"+newKernelVer),
		)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 7.5, err)
			return err
		}
		err = MoveFile(
			filepath.Join(futureRoot, "boot", "initrd.img-"+newKernelVer),
			filepath.Join(futureInitDir, "initrd.img-"+newKernelVer),
		)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 7.6, err)
			return err
		}
		err = MoveFile(
			filepath.Join(futureRoot, "boot", "config-"+newKernelVer),
			filepath.Join(futureInitDir, "config-"+newKernelVer),
		)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 7.7, err)
			return err
		}
		err = MoveFile(
			filepath.Join(futureRoot, "boot", "System.map-"+newKernelVer),
			filepath.Join(futureInitDir, "System.map-"+newKernelVer),
		)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 7.8, err)
			return err
		}

		rootUuid = initPartition.Uuid
	} else {
		rootUuid = partFuture.Partition.Uuid
	}

	err = generateABGrubConf(
		newKernelVer,
		futureRoot,
		rootUuid,
		partFuture.Label,
		generatedGrubConfigPath,
	)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 7.9, err)
		return err
	}

	// Stage 7: Sync /etc
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 7] -------- ABSystemRunOperation")

	oldEtc := "/sysconf" // The current etc WITHOUT anything overlayed
	oldUpperEtc := fmt.Sprintf("/var/lib/abroot/etc/%s", partPresent.Label)
	newUpperEtc := fmt.Sprintf("/var/lib/abroot/etc/%s", partFuture.Label)

	// make sure the future etc directories exist, ignoring errors
	newWorkEtc := fmt.Sprintf("/var/lib/abroot/etc/%s-work", partFuture.Label)
	os.MkdirAll(newUpperEtc, 0o755)
	os.MkdirAll(newWorkEtc, 0o755)

	err = EtcBuilder.ExtBuildCommand(oldEtc, futureRoot+"/sysconf", oldUpperEtc, newUpperEtc)
	if err != nil {
		PrintVerboseErr("AbSystem.RunOperation", 8, err)
		return err
	}

	// Stage 8: Mount boot partition
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 8] -------- ABSystemRunOperation")

	tmpBootMount := "/run/abroot/tmp-boot-mount-1/"
	err = os.MkdirAll(tmpBootMount, 0o755)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 9, err)
		return err
	}

	err = partBoot.Mount(tmpBootMount)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 9.1, err)
		return err
	}

	cq.Add(func(args ...interface{}) error {
		return partBoot.Unmount()
	}, nil, 100, &goodies.NoErrorHandler{}, false)

	// Stage 9: Atomic swap the bootloader
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 9] -------- ABSystemRunOperation")

	grub, err := NewGrub(partBoot)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 11, err)
		return err
	}

	// Only swap grub entries if we're booted into the present partition
	isPresent, err := grub.IsBootedIntoPresentRoot()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 11.1, err)
		return err
	}
	if isPresent {
		grubCfgCurrent := filepath.Join(tmpBootMount, "grub/grub.cfg")
		grubCfgFuture := filepath.Join(tmpBootMount, "grub/grub.cfg.future")

		// Just like in Stage 9, tmpBootMount/grub/grub.cfg.future may not exist.
		if _, err = os.Stat(grubCfgFuture); os.IsNotExist(err) {
			PrintVerboseInfo("ABSystem.RunOperation", "Creating grub.cfg.future")

			grubCfgContents, err := os.ReadFile(grubCfgCurrent)
			if err != nil {
				PrintVerboseErr("ABSystem.RunOperation", 11.2, err)
			}

			var replacerPairs []string
			if grub.FutureRoot == "a" {
				replacerPairs = []string{
					"default=1", "default=0",
					"Previous State (A)", "Current State (A)",
					"Current State (B)", "Previous State (B)",
				}
			} else {
				replacerPairs = []string{
					"default=0", "default=1",
					"Current State (A)", "Previous State (A)",
					"Previous State (B)", "Current State (B)",
				}
			}

			replacer := strings.NewReplacer(replacerPairs...)
			os.WriteFile(grubCfgFuture, []byte(replacer.Replace(string(grubCfgContents))), 0o644)
		}

		err = AtomicSwap(grubCfgCurrent, grubCfgFuture)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 11.3, err)
			return err
		}
	}

	err = s.createFinishedFile()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 11.4, err)
		return fmt.Errorf("could not write finished file: %w", err)
	}

	PrintVerboseInfo("ABSystem.RunOperation", "upgrade completed")
	return nil
}

// Rollback swaps the master grub files if the current root is not the default
func (s *ABSystem) Rollback(checkOnly bool) (response ABRollbackResponse, err error) {
	PrintVerboseInfo("ABSystem.Rollback", "starting")

	cq := goodies.NewCleanupQueue()
	defer cq.Run()

	if s.finishedFileExists() {
		if checkOnly {
			return ROLLBACK_RES_NO, nil
		}
		return ROLLBACK_FAILED, errors.New("an operation finished successfully, can't roll back until reboot")
	}

	// we won't allow upgrades while rolling back
	if !checkOnly {
		err = s.LockOperation()
		if err != nil {
			PrintVerboseErr("ABSystem.Rollback", 0, err)
			return ROLLBACK_FAILED, fmt.Errorf("can't lock operation: %w", err)
		}
	}

	partBoot, err := s.RootM.GetBoot()
	if err != nil {
		PrintVerboseErr("ABSystem.Rollback", 1, err)
		return ROLLBACK_FAILED, err
	}

	tmpBootMount := "/run/abroot/tmp-boot-mount-2/"
	err = os.MkdirAll(tmpBootMount, 0o755)
	if err != nil {
		PrintVerboseErr("ABSystem.Rollback", 2, err)
		return ROLLBACK_FAILED, err
	}

	err = partBoot.Mount(tmpBootMount)
	if err != nil {
		PrintVerboseErr("ABSystem.Rollback", 3, err)
		return ROLLBACK_FAILED, err
	}

	cq.Add(func(args ...interface{}) error {
		return partBoot.Unmount()
	}, nil, 100, &goodies.NoErrorHandler{}, false)

	grub, err := NewGrub(partBoot)
	if err != nil {
		PrintVerboseErr("ABSystem.Rollback", 4, err)
		return ROLLBACK_FAILED, err
	}

	// Only swap grub entries if we're booted into the present partition
	isPresent, err := grub.IsBootedIntoPresentRoot()
	if err != nil {
		PrintVerboseErr("ABSystem.Rollback", 5, err)
		return ROLLBACK_FAILED, err
	}

	// If checkOnly is true, we stop here and return the appropriate response
	if checkOnly {
		response = ROLLBACK_RES_YES
		if isPresent {
			response = ROLLBACK_RES_NO
		}
		return response, nil
	}

	if isPresent {
		PrintVerboseInfo("ABSystem.Rollback", "current root is the default, nothing to do")
		return ROLLBACK_UNNECESSARY, nil
	}

	grubCfgCurrent := filepath.Join(tmpBootMount, "grub/grub.cfg")
	grubCfgFuture := filepath.Join(tmpBootMount, "grub/grub.cfg.future")

	// Just like in Stage 9, tmpBootMount/grub/grub.cfg.future may not exist.
	if _, err = os.Stat(grubCfgFuture); os.IsNotExist(err) {
		PrintVerboseInfo("ABSystem.Rollback", "Creating grub.cfg.future")

		grubCfgContents, err := os.ReadFile(grubCfgCurrent)
		if err != nil {
			PrintVerboseErr("ABSystem.Rollback", 6, err)
		}

		var replacerPairs []string
		if grub.FutureRoot == "a" {
			replacerPairs = []string{
				"default=1", "default=0",
				"A (previous)", "A (current)",
				"B (current)", "B (previous)",
			}
		} else {
			replacerPairs = []string{
				"default=0", "default=1",
				"A (current)", "A (previous)",
				"B (previous)", "B (current)",
			}
		}

		replacer := strings.NewReplacer(replacerPairs...)
		os.WriteFile(grubCfgFuture, []byte(replacer.Replace(string(grubCfgContents))), 0o644)
	}

	err = AtomicSwap(grubCfgCurrent, grubCfgFuture)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 7, err)
		return ROLLBACK_FAILED, err
	}

	// allow upgrades after rolling back
	err = s.UnlockOperation()
	if err != nil {
		PrintVerboseErr("ABSystem.Rollback", 9, err)
		PrintVerboseInfo("ABSystem.Rollback", "rollback completed with unlock failure")
	}

	PrintVerboseInfo("ABSystem.Rollback", "rollback completed")
	return ROLLBACK_SUCCESS, nil
}

// LockOperation creates a lock file, preventing upgrades from proceeding
//
// Returns ErrOperationLocked if the operation is already locked by a running abroot instance
func (s *ABSystem) LockOperation() error {
	pid := os.Getpid()
	pidData := []byte(strconv.Itoa(pid))

	err := os.MkdirAll(filepath.Dir(operationLockFile), 0o755)
	if err != nil {
		PrintVerboseWarn("ABSystem.LockOperation", 1, err)
	}

	if _, err = os.Stat(operationLockFile); err == nil {
		if s.isLockfilePidActive() {
			return ErrOperationLocked
		}
	}

	err = os.WriteFile(operationLockFile, pidData, 0o644)
	if err != nil {
		os.Remove(operationLockFile)
		return fmt.Errorf("can't write lockfile: %w", err)
	}

	PrintVerboseInfo("ABSystem.LockOperation", "lock file created")
	return nil
}

// UnlockOperation removes the lock file, allowing upgrades to proceed
func (s *ABSystem) UnlockOperation() error {
	err := os.Remove(operationLockFile)
	if err != nil {
		PrintVerboseErr("ABSystem.UnlockOperation", 0, err)
		return err
	}

	PrintVerboseInfo("ABSystem.UnlockOperation", "lock file removed")
	return nil
}

func (s *ABSystem) finishedFileExists() bool {
	_, err := os.Stat(finishedOperationFile)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *ABSystem) createFinishedFile() error {
	os.MkdirAll(filepath.Dir(finishedOperationFile), 0o755)

	_, err := os.Create(finishedOperationFile)
	if err != nil {
		return err
	}
	return nil
}

func (s *ABSystem) isLockfilePidActive() bool {
	runningPid, err := os.ReadFile(operationLockFile)

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	if err != nil {
		PrintVerboseErr("ABSystem.RemoveStageFile", 0, err)
		return true
	}

	if string(runningPid) == "" {
		PrintVerboseWarn("ABSystem.isLockfilePidActive", 1, "lock file does not contain PID")
		return true
	}

	if _, err := os.Stat(filepath.Join("/proc", string(runningPid))); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func (s *ABSystem) createFinalizingFile() error {
	os.MkdirAll(filepath.Dir(finalizingFile), 0o755)

	_, err := os.Create(finalizingFile)
	if err != nil {
		return err
	}
	return nil
}

func (s *ABSystem) removeFinalizingFile() error {
	err := os.Remove(finalizingFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// UserStopRequested checks if the user lock file exists and returns a boolean
// note that if the user lock file exists, it means that the user explicitly
// requested the upgrade to be locked (using an update manager for example)
func UserStopRequested() bool {
	if _, err := os.Stat(userStopFile); os.IsNotExist(err) {
		return false
	}

	PrintVerboseInfo("ABSystem.UserStopRequested", "lock file exists")
	return true
}

// MakeStopRequest requests all other abroot operations to stop
//
// It also prevents any new operations from running.
func MakeStopRequest() error {
	os.MkdirAll(filepath.Dir(userStopFile), 0o755)
	err := os.WriteFile(userStopFile, []byte{}, 0o644)
	if err != nil {
		return fmt.Errorf("could not write stop file: %w", err)
	}
	return nil
}

// CancelStopRequest removes the stop request
func CancelStopRequest() error {
	if !UserStopRequested() {
		return nil
	}

	err := os.Remove(userStopFile)
	if err != nil {
		return fmt.Errorf("could remove stop file: %w", err)
	}
	return nil
}
