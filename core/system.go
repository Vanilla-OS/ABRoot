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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	EtcBuilder "github.com/linux-immutability-tools/EtcBuilder/cmd"
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

	// Registry contains an instance of the Registry used to retrieve resources
	// from the configured Docker registry.
	Registry *Registry

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
	lockFile     string = filepath.Join("/tmp", "ABSystem.Upgrade.lock")
	stageFile    string = filepath.Join("/tmp", "ABSystem.Upgrade.stage")
	userLockFile string = filepath.Join("/tmp", "ABSystem.Upgrade.user.lock")

	// Errors
	ErrNoUpdate error = errors.New("no update available")
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
	r := NewRegistry()
	rm := NewABRootManager()

	return &ABSystem{
		Checks:   c,
		RootM:    rm,
		Registry: r,
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
func (s *ABSystem) CheckUpdate() (string, bool) {
	PrintVerboseInfo("ABSystem.CheckUpdate", "running...")
	return s.Registry.HasUpdate(s.CurImage.Digest)
}

func (s *ABSystem) CreateRootSymlinks(systemNewPath string) error {
	PrintVerboseInfo("ABSystem.CreateRootSymlinks", "creating symlinks")
	links := []string{"mnt", "proc", "run", "dev", "media", "root", "sys", "tmp", "var"}

	for _, link := range links {
		linkName := filepath.Join(systemNewPath, link)

		err := os.RemoveAll(linkName)
		if err != nil {
			PrintVerboseErr("ABSystem.CreateRootSymlinks", 1, err)
			return err
		}

		targetName := filepath.Join("/", link)

		err = os.Symlink(targetName, linkName)
		if err != nil {
			PrintVerboseErr("ABSystem.CreateRootSymlinks", 2, err)
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
//		Applies package changes, but doesn't update the system.
//	INITRAMFS:
//		Updates the initramfs for the future root, but doesn't update the system.
func (s *ABSystem) RunOperation(operation ABSystemOperation, freeSpace bool) error {
	PrintVerboseInfo("ABSystem.RunOperation", "starting", operation)

	cq := goodies.NewCleanupQueue()
	defer cq.Run()

	// Stage 0: Check if upgrade is possible
	// -------------------------------------
	PrintVerboseSimple("[Stage 0] -------- ABSystemRunOperation")

	if s.UpgradeLockExists() {
		if isAbrootRunning() {
			PrintVerboseWarn("ABSystemRunOperation", 0, "upgrades are locked, another is running")
			return errors.New("upgrades are locked, another is running")
		}

		err := removeUpgradeLock()
		if err != nil {
			PrintVerboseErr("ABSystemRunOperation", 0, err)
			return err
		}
	}

	err := s.LockUpgrade()
	if err != nil {
		PrintVerboseErr("ABSystemRunOperation", 0.1, err)
		return err
	}

	// here we create the stage file to indicate that the upgrade is in progress
	// and the process can safely be stopped
	err = s.CreateStageFile()
	if err != nil {
		PrintVerboseErr("ABSystemRunOperation", 0.2, err)
		return err
	}

	cq.Add(func(args ...interface{}) error {
		return s.UnlockUpgrade()
	}, nil, 100, &goodies.NoErrorHandler{}, false)

	// Stage 1: Check if there is an update available
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 1] -------- ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerboseErr("ABSystemRunOperation", 1, err)
		return err
	}

	var imageDigest string
	if operation != APPLY && operation != INITRAMFS {
		var res bool
		imageDigest, res = s.CheckUpdate()
		if !res {
			if operation != FORCE_UPGRADE {
				PrintVerboseErr("ABSystemRunOperation", 1.1, err)
				return ErrNoUpdate
			}
			imageDigest = s.CurImage.Digest
			PrintVerboseWarn("ABSystemRunOperation", 1.2, "No update available but --force is set. Proceeding...")
		}
	} else {
		imageDigest = s.CurImage.Digest
	}

	// Stage 2: Get the future root and boot partitions,
	// 			mount future to /part-future and clean up
	// 			old .system_new and abimage-new.abr (it is
	// 			possible that last transaction was interrupted
	// 			before the clean up was done). Finally run
	// 			the IntegrityCheck on the future root.
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 2] -------- ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerboseErr("ABSystemRunOperation", 2, err)
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

	err = partFuture.Partition.Mount("/part-future/")
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 2.3, err)
		return err
	}

	os.RemoveAll("/part-future/.system_new")
	os.RemoveAll("/part-future/abimage-new.abr") // errors are safe to ignore

	cq.Add(func(args ...interface{}) error {
		return partFuture.Partition.Unmount()
	}, nil, 90, &goodies.NoErrorHandler{}, false)

	err = RepairRootIntegrity(partFuture.Partition.MountPoint)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 2.4, err)
		return err
	}

	// Stage 3: Make a imageRecipe with user packages
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 3] -------- ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerboseErr("ABSystemRunOperation", 3, err)
		return err
	}

	futurePartition, err := s.RootM.GetFuture()
	if err != nil {
		PrintVerboseErr("ABSystemRunOperation", 3.1, err)
		return err
	}

	labels := map[string]string{
		"maintainer":  "'Generated by ABRoot'",
		"ABRoot.root": futurePartition.Label,
	}
	args := map[string]string{}
	pkgM, err := NewPackageManager(false)
	if err != nil {
		PrintVerboseErr("ABSystemRunOperation", 3.2, err)
		return err
	}

	pkgsFinal := pkgM.GetFinalCmd(operation)
	if pkgsFinal == "" {
		pkgsFinal = "true"
	}
	content := `RUN ` + pkgsFinal

	var imageName string
	switch operation {
	case APPLY, DRY_RUN_APPLY, INITRAMFS, DRY_RUN_INITRAMFS:
		presentPartition, err := s.RootM.GetPresent()
		if err != nil {
			PrintVerboseErr("ABSystemRunOperation", 3.3, err)
			return err
		}
		imageName, err = RetrieveImageForRoot(presentPartition.Label)
		if err != nil {
			PrintVerboseErr("ABSystemRunOperation", 3.4, err)
			return err
		}
		// Handle case where an image for the current root may not exist
		// in storage
		if imageName == "" {
			imageName = settings.Cnf.FullImageName
		}
	default:
		imageName = strings.Split(settings.Cnf.FullImageName, ":")[0]
		imageName += "@" + imageDigest
		labels["ABRoot.BaseImageDigest"] = imageDigest
	}

	// Stage 3.1: Delete old image
	switch operation {
	case DRY_RUN_UPGRADE, DRY_RUN_APPLY, DRY_RUN_INITRAMFS:
	default:
		err = DeleteImageForRoot(futurePartition.Label)
		if err != nil {
			PrintVerboseErr("ABSystemRunOperation", 3.5, err)
			return err
		}
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

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerboseErr("ABSystemRunOperation", 4, err)
		return err
	}

	abrootTrans := filepath.Join(partFuture.Partition.MountPoint, "abroot-trans")
	systemOld := filepath.Join(partFuture.Partition.MountPoint, ".system")
	systemNew := filepath.Join(partFuture.Partition.MountPoint, ".system.new")
	if freeSpace || os.Getenv("ABROOT_FREE_SPACE") != "" {
		PrintVerboseInfo("ABSystemRunOperation", "Deleting future system to free space, this will render the future root temporarily unavailable")
		err := os.RemoveAll(systemOld)
		if err != nil {
			PrintVerboseErr("ABSystemRunOperation", 4, err)
			return err
		}
		err = os.MkdirAll(systemOld, 0o755)
		if err != nil {
			PrintVerboseErr("ABSystemRunOperation", 4.1, err)
			return err
		}
	} else {
		PrintVerboseInfo("ABSystemRunOperation", "Creating a reflink clone of the old system to copy into")
		err := os.RemoveAll(systemNew)
		if err != nil {
			PrintVerboseErr("ABSystemRunOperation", 4.11, "could not cleanup old systemNew folder", err)
			return err
		}
		err = exec.Command("cp", "--reflink", "-a", systemOld, systemNew).Run()
		if err != nil {
			PrintVerboseWarn("ABSystemRunOperation", 4.12, "reflink copy of system failed, falling back to slow copy because:", err)
			// can be safely ignored
			// file system doesn't support CoW
		}
	}

	err = OciExportRootFs(
		"abroot-"+uuid.New().String(),
		imageRecipe,
		abrootTrans,
		systemNew,
	)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 4.2, err)
		return err
	}

	cq.Add(func(args ...interface{}) error {
		return pkgM.ClearUnstagedPackages()
	}, nil, 10, &goodies.NoErrorHandler{}, false)

	// Stage 5: Write abimage.abr.new and config to future/
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 5] -------- ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerboseErr("ABSystemRunOperation", 5, err)
		return err
	}

	abimage, err := NewABImage(imageDigest, settings.Cnf.FullImageName)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 5.1, err)
		return err
	}

	err = abimage.WriteTo(partFuture.Partition.MountPoint, "new")
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

	err = settings.WriteConfigToFile(filepath.Join(systemNew, "/usr/share/abroot/abroot.json"))
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 5.25, err)
		return err
	}

	err = pkgM.WriteSummaryToFile(filepath.Join(systemNew, "/usr/share/abroot/package-summary"))
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 5.26, err)
		return err
	}

	// from this point on, it is not possible to stop the upgrade
	// so we remove the stage file. Note that interrupting the upgrade
	// from this point on will not leave the system in an inconsistent
	// state, but it could leave the future partition in a dirty state
	// preventing it from booting.
	err = s.RemoveStageFile()
	if err != nil {
		PrintVerboseErr("ABSystemRunOperation", 5.3, err)
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

	partPresent, err := s.RootM.GetPresent()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 7.01, "failed to get present partition:", err)
	}

	chroot, err := NewChroot(
		systemNew,
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

	newKernelVer := getKernelVersion(filepath.Join(systemNew, "boot"))
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

		initMountpoint = filepath.Join(systemNew, "boot", "init")
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

		err = CopyFile(
			filepath.Join(systemNew, "boot", "vmlinuz-"+newKernelVer),
			filepath.Join(futureInitDir, "vmlinuz-"+newKernelVer),
		)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 7.5, err)
			return err
		}
		err = CopyFile(
			filepath.Join(systemNew, "boot", "initrd.img-"+newKernelVer),
			filepath.Join(futureInitDir, "initrd.img-"+newKernelVer),
		)
		if err != nil {
			PrintVerboseErr("ABSystem.RunOperation", 7.6, err)
			return err
		}

		os.Remove(filepath.Join(systemNew, "boot", "vmlinuz-"+newKernelVer))
		os.Remove(filepath.Join(systemNew, "boot", "initrd.img-"+newKernelVer))

		rootUuid = initPartition.Uuid
	} else {
		rootUuid = partFuture.Partition.Uuid
	}

	err = generateABGrubConf(
		newKernelVer,
		systemNew,
		rootUuid,
		partFuture.Label,
		generatedGrubConfigPath,
	)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 7.7, err)
		return err
	}

	// Create links back to the root system
	err = s.CreateRootSymlinks(systemNew)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 7.8, err)
		return err
	}

	// Stage 7: Sync /etc
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 7] -------- ABSystemRunOperation")

	oldEtc := "/.system/sysconf" // The current etc WITHOUT anything overlayed
	presentEtc, err := s.RootM.GetPresent()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 8, err)
		return err
	}
	futureEtc, err := s.RootM.GetFuture()
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 8.1, err)
		return err
	}
	oldUpperEtc := fmt.Sprintf("/var/lib/abroot/etc/%s", presentEtc.Label)
	newUpperEtc := fmt.Sprintf("/var/lib/abroot/etc/%s", futureEtc.Label)

	// make sure the future etc directories exist, ignoring errors
	newWorkEtc := fmt.Sprintf("/var/lib/abroot/etc/%s-work", futureEtc.Label)
	os.MkdirAll(newUpperEtc, 0o755)
	os.MkdirAll(newWorkEtc, 0o755)

	err = EtcBuilder.ExtBuildCommand(oldEtc, systemNew+"/sysconf", oldUpperEtc, newUpperEtc)
	if err != nil {
		PrintVerboseErr("AbSystem.RunOperation", 8.2, err)
		return err
	}

	// Stage 8: Mount boot partition
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 8] -------- ABSystemRunOperation")

	uuid := uuid.New().String()
	tmpBootMount := filepath.Join("/tmp", uuid)
	err = os.Mkdir(tmpBootMount, 0o755)
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

	// Stage 9: Atomic swap the rootfs and abimage.abr
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 9] -------- ABSystemRunOperation")

	err = AtomicSwap(systemOld, systemNew)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 10, err)
		return err
	}

	cq.Add(func(args ...interface{}) error {
		return os.RemoveAll(systemNew)
	}, nil, 20, &goodies.NoErrorHandler{}, false)

	oldABImage := filepath.Join(partFuture.Partition.MountPoint, "abimage.abr")
	newABImage := filepath.Join(partFuture.Partition.MountPoint, "abimage-new.abr")

	// PartFuture may not have /abimage.abr if it got corrupted or was wiped.
	// In these cases, create a dummy file for the atomic swap.
	if _, err = os.Stat(oldABImage); os.IsNotExist(err) {
		PrintVerboseInfo("ABSystem.RunOperation", "Creating dummy /part-future/abimage.abr")
		os.Create(oldABImage)
	}

	err = AtomicSwap(oldABImage, newABImage)
	if err != nil {
		PrintVerboseErr("ABSystem.RunOperation", 10.1, err)
		return err
	}

	cq.Add(func(args ...interface{}) error {
		return os.RemoveAll(newABImage)
	}, nil, 30, &goodies.NoErrorHandler{}, false)

	// Stage 10: Atomic swap the bootloader
	// ------------------------------------------------
	PrintVerboseSimple("[Stage 10] -------- ABSystemRunOperation")

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

	PrintVerboseInfo("ABSystem.RunOperation", "upgrade completed")
	return nil
}

// Rollback swaps the master grub files if the current root is not the default
func (s *ABSystem) Rollback(checkOnly bool) (response ABRollbackResponse, err error) {
	PrintVerboseInfo("ABSystem.Rollback", "starting")

	cq := goodies.NewCleanupQueue()
	defer cq.Run()

	// we won't allow upgrades while rolling back
	if !checkOnly {
		err = s.LockUpgrade()
		if err != nil {
			PrintVerboseErr("ABSystem.Rollback", 0, err)
			return ROLLBACK_FAILED, err
		}
	}

	partBoot, err := s.RootM.GetBoot()
	if err != nil {
		PrintVerboseErr("ABSystem.Rollback", 1, err)
		return ROLLBACK_FAILED, err
	}

	uuid := uuid.New().String()
	tmpBootMount := filepath.Join("/tmp", uuid)
	err = os.Mkdir(tmpBootMount, 0o755)
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
	err = s.UnlockUpgrade()
	if err != nil {
		PrintVerboseErr("ABSystem.Rollback", 9, err)
		PrintVerboseInfo("ABSystem.Rollback", "rollback completed with unlock failure")
	}

	PrintVerboseInfo("ABSystem.Rollback", "rollback completed")
	return ROLLBACK_SUCCESS, nil
}

// UserLockRequested checks if the user lock file exists and returns a boolean
// note that if the user lock file exists, it means that the user explicitly
// requested the upgrade to be locked (using an update manager for example)
func (s *ABSystem) UserLockRequested() bool {
	if _, err := os.Stat(userLockFile); os.IsNotExist(err) {
		return false
	}

	PrintVerboseInfo("ABSystem.UserLockRequested", "lock file exists")
	return true
}

// UpgradeLockExists checks if the lock file exists and returns a boolean
func (s *ABSystem) UpgradeLockExists() bool {
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		return false
	}

	PrintVerboseInfo("ABSystem.UpgradeLockExists", "lock file exists")
	return true
}

// LockUpgrade creates a lock file, preventing upgrades from proceeding
func (s *ABSystem) LockUpgrade() error {
	_, err := os.Create(lockFile)
	if err != nil {
		PrintVerboseErr("ABSystem.LockUpgrade", 0, err)
		return err
	}

	PrintVerboseInfo("ABSystem.LockUpgrade", "lock file created")
	return nil
}

// UnlockUpgrade removes the lock file, allowing upgrades to proceed
func (s *ABSystem) UnlockUpgrade() error {
	err := os.Remove(lockFile)
	if err != nil {
		PrintVerboseErr("ABSystem.UnlockUpgrade", 0, err)
		return err
	}

	PrintVerboseInfo("ABSystem.UnlockUpgrade", "lock file removed")
	return nil
}

// CreateStageFile creates the stage file, which is used to determine if
// the upgrade can be interrupted or not. If the stage file is present, it
// means that the upgrade is in a state where it is still possible to
// interrupt it, otherwise it is not. This is useful for third-party
// applications like update managers.
func (s *ABSystem) CreateStageFile() error {
	_, err := os.Create(stageFile)
	if err != nil {
		PrintVerboseErr("ABSystem.CreateStageFile", 0, err)
		return err
	}

	PrintVerboseInfo("ABSystem.CreateStageFile", "stage file created")
	return nil
}

// RemoveStageFile removes the stage file disabling the ability to interrupt
// the upgrade process
func (s *ABSystem) RemoveStageFile() error {
	err := os.Remove(stageFile)
	if err != nil {
		PrintVerboseErr("ABSystem.RemoveStageFile", 0, err)
		return err
	}

	PrintVerboseInfo("ABSystem.RemoveStageFile", "stage file removed")
	return nil
}

// isAbrootRunning checks if an instance of the `abroot` command is running
// other than the current process
func isAbrootRunning() bool {
	pid := os.Getpid()
	procs, err := os.ReadDir("/proc")
	if err != nil {
		return false
	}

	for _, file := range procs {
		if file.IsDir() {
			if _, err := strconv.Atoi(file.Name()); err == nil {
				cmdline, err := os.ReadFile("/proc/" + file.Name() + "/cmdline")
				exe, exeErr := os.Readlink("/proc/" + file.Name() + "/exe")
				if (err == nil && strings.Contains(string(cmdline), "abroot")) || (exeErr == nil && strings.Contains(exe, "abroot")) {
					procPid, _ := strconv.Atoi(file.Name())
					if procPid != pid {
						return true
					}
				}
			}
		}
	}
	return false
}

// removeUpgradeLock removes the lock file, allowing upgrades to proceed
func removeUpgradeLock() error {
	err := os.Remove(lockFile)
	if err != nil {
		PrintVerboseErr("removeUpgradeLock", 0, err)
		return err
	}

	PrintVerboseInfo("removeUpgradeLock", "upgrade lock removed")
	return nil
}
