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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/vanilla-os/abroot/settings"
)

// ABSystem represents the system
type ABSystem struct {
	Checks   *Checks
	RootM    *ABRootManager
	Registry *Registry
	CurImage *ABImage
}

type QueuedFunction struct {
	Name     string
	Values   []interface{}
	Priority int
}

const (
	// ABSystem operations
	UPGRADE           = "upgrade"
	FORCE_UPGRADE     = "force-upgrade"
	DRY_RUN_UPGRADE   = "dry-run-upgrade"
	APPLY             = "package-apply"
	DRY_RUN_APPLY     = "dry-run-package-apply"
	INITRAMFS         = "initramfs"
	DRY_RUN_INITRAMFS = "dry-run-initramfs"

	// ABSystem rollback response
	ROLLBACK_UNNECESSARY = "rollback-unnecessary"
	ROLLBACK_SUCCESS     = "rollback-success"
	ROLLBACK_FAILED      = "rollback-failed"
)

const (
	MountUnitDir = "/etc/systemd/system"
)

type ABSystemOperation string
type ABRollbackResponse string

var (
	queue        []QueuedFunction
	lockFile     string = filepath.Join("/tmp", "ABSystem.Upgrade.lock")
	stageFile    string = filepath.Join("/tmp", "ABSystem.Upgrade.stage")
	userLockFile string = filepath.Join("/tmp", "ABSystem.Upgrade.user.lock")
	ErrNoUpdate  error  = errors.New("no update available")
)

// NewABSystem creates a new system
func NewABSystem() (*ABSystem, error) {
	PrintVerbose("NewABSystem: running...")

	i, err := NewABImageFromRoot()
	if err != nil {
		PrintVerbose("NewABSystem:err: %s", err)
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
	PrintVerbose("ABSystem.CheckAll: running...")

	err := s.Checks.PerformAllChecks()
	if err != nil {
		PrintVerbose("ABSystem.CheckAll:err: %s", err)
		return err
	}

	PrintVerbose("ABSystem.CheckAll: all checks passed")
	return nil
}

// CheckUpdate checks if there is an update available
func (s *ABSystem) CheckUpdate() (string, bool) {
	PrintVerbose("ABSystem.CheckUpdate: running...")
	return s.Registry.HasUpdate(s.CurImage.Digest)
}

// MergeUserEtcFiles merges user-related files from the new lower etc (/.system/etc)
// with the old upper etc, if present, saving the result in the new upper etc.
func (s *ABSystem) MergeUserEtcFiles(oldUpperEtc, newLowerEtc, newUpperEtc string) error {
	PrintVerbose("ABSystem.SyncLowerEtc: syncing /.system/etc -> %s", newLowerEtc)

	etcFiles := []string{
		"passwd",
		"group",
		"shells",
		"shadow",
		"subuid",
		"subgid",
	}

	etcDir := "/.system/etc"
	if _, err := os.Stat(etcDir); os.IsNotExist(err) {
		PrintVerbose("ABSystem.SyncLowerEtc:err: %s", err)
		return err
	}

	for _, file := range etcFiles {
		// Use file present in the immutable /etc if it exists. Otherwise, use the immutable one.
		_, err := os.Stat(oldUpperEtc + "/" + file)
		if err != nil {
			if os.IsNotExist(err) { // No changes were made to the file from its image base, skip merge
				continue
			} else {
				PrintVerbose("ABSystem.SyncLowerEtc:err(2): %s", err)
				return err
			}
		} else {
			firstFile := oldUpperEtc + "/" + file
			secondFile := newLowerEtc + "/" + file
			destination := newUpperEtc + "/" + file

			// write the diff to the destination
			err = MergeDiff(firstFile, secondFile, destination)
			if err != nil {
				PrintVerbose("ABSystem.SyncLowerEtc:err(3): %s", err)
				return err
			}
		}
	}

	PrintVerbose("ABSystem.SyncLowerEtc: sync completed")
	return nil
}

// SyncUpperEtc syncs the mutable etc directories from /var/lib/abroot/etc
func (s *ABSystem) SyncUpperEtc(newEtc string) error {
	PrintVerbose("ABSystem.SyncUpperEtc: Starting")

	current_part, err := s.RootM.GetPresent()
	if err != nil {
		PrintVerbose("ABSystem.SyncUpperEtc:err: %s", err)
		return err
	}

	etcDir := fmt.Sprintf("/var/lib/abroot/etc/%s", current_part.Label)
	if _, err := os.Stat(etcDir); os.IsNotExist(err) {
		PrintVerbose("ABSystem.SyncEtc:err(2): %s", err)
		return err
	}

	PrintVerbose("ABSystem.SyncUpperEtc: syncing /var/lib/abroot/etc/%s -> %s", current_part.Label, newEtc)

	err = exec.Command( // TODO: use the Rsync method here
		"rsync",
		"-a",
		"--exclude=passwd",
		"--exclude=group",
		"--exclude=shells",
		"--exclude=shadow",
		"--exclude=subuid",
		"--exclude=subgid",
		"--exclude=fstab",
		"--exclude=crypttab",
		etcDir+"/",
		newEtc,
	).Run()
	if err != nil {
		PrintVerbose("ABSystem.SyncUpperEtc:err: %s", err)
		return err
	}

	PrintVerbose("ABSystem.SyncUpperEtc: sync completed")
	return nil
}

// RunCleanUpQueue runs the functions in the queue or only the specified one
func (s *ABSystem) RunCleanUpQueue(fnName string) error {
	PrintVerbose("ABSystem.RunCleanUpQueue: running...")

	sort.Slice(queue, func(i, j int) bool {
		return queue[i].Priority < queue[j].Priority
	})

	itemsToRemove := []int{}
	for i, f := range queue {
		if fnName != "" && f.Name != fnName {
			continue
		}

		itemsToRemove = append(itemsToRemove, i)

		switch f.Name {
		case "umountFuture":
			PrintVerbose("ABSystem.RunCleanUpQueue: Executing umountFuture")
			futurePart := f.Values[0].(ABRootPartition)
			err := futurePart.Partition.Unmount()
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err: %s", err)
				return err
			}
		case "closeChroot":
			PrintVerbose("ABSystem.RunCleanUpQueue: Executing closeChroot")
			chroot := f.Values[0].(*Chroot)
			chroot.Close() // safe to ignore, already closed
		case "removeNewSystem":
			PrintVerbose("ABSystem.RunCleanUpQueue: Executing removeNewSystem")
			newSystem := f.Values[0].(string)
			err := os.RemoveAll(newSystem)
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(3): %s", err)
				return err
			}
		case "removeNewABImage":
			PrintVerbose("ABSystem.RunCleanUpQueue: Executing removeNewABImage")
			newImage := f.Values[0].(string)
			err := os.RemoveAll(newImage)
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(4): %s", err)
				return err
			}
		case "umountBoot":
			PrintVerbose("ABSystem.RunCleanUpQueue: Executing umountBoot")
			bootPart := f.Values[0].(Partition)
			err := bootPart.Unmount()
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(5): %s", err)
				return err
			}
		case "umountInit":
			PrintVerbose("ABSystem.RunCleanUpQueue: Executing umountInit")
			initPart := Partition{MountPoint: "/part-future/.system/boot/init"}
			err := initPart.Unmount()
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(6): %s", err)
				return err
			}
		case "unlockUpgrade":
			PrintVerbose("ABSystem.RunCleanUpQueue: Executing unlockUpgrade")
			err := s.UnlockUpgrade()
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(7): %s", err)
				return err
			}
		case "clearUnstagedPackages":
			PrintVerbose("ABSystem.RunCleanUpQueue: Executing clearUnstagedPackages")
			pkgM := NewPackageManager(false)
			err := pkgM.ClearUnstagedPackages()
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(8): %s", err)
				return err
			}
		}
	}

	// Remove matched items in reverse order to avoid changing indices
	for i := len(itemsToRemove) - 1; i >= 0; i-- {
		removeIdx := itemsToRemove[i]
		queue = append(queue[:removeIdx], queue[removeIdx+1:]...)
	}

	PrintVerbose("ABSystem.RunCleanUpQueue: completed")
	return nil
}

// AddToCleanUpQueue adds a function to the queue
func (s *ABSystem) AddToCleanUpQueue(name string, priority int, values ...interface{}) {
	queue = append(queue, QueuedFunction{
		Name:     name,
		Values:   values,
		Priority: priority,
	})
}

// RemoveFromCleanUpQueue removes a function from the queue
func (s *ABSystem) RemoveFromCleanUpQueue(name string) {
	for i := 0; i < len(queue); i++ {
		if queue[i].Name == name {
			queue = append(queue[:i], queue[i+1:]...)
			i--
		}
	}
}

// ResetQueue resets the queue
func (s *ABSystem) ResetQueue() {
	queue = []QueuedFunction{}
}

// GenerateFstab generates a fstab file for the future root
func (s *ABSystem) GenerateFstab(rootPath string, root ABRootPartition) error {
	PrintVerbose("ABSystem.GenerateFstab: generating fstab")

	template := `# /etc/fstab: static file system information.
# Generated by ABRoot
#
# <file system> <mount point>   <type>  <options>       <dump>  <pass>
UUID=%s  /  %s  defaults  0  0
/.system/usr  /.system/usr  none  bind,ro
%s  /var  auto  defaults  0  0
`
	varSource := ""
	if s.RootM.VarPartition.IsDevMapper() {
		varSource = fmt.Sprintf("/dev/mapper/%s", s.RootM.VarPartition.Device)
	} else {
		varSource = fmt.Sprintf("-U %s", s.RootM.VarPartition.Uuid)
	}

	fstab := fmt.Sprintf(
		template,
		root.Partition.Uuid,
		root.Partition.FsType,
		varSource,
	)

	err := os.WriteFile(filepath.Join(rootPath, "/etc/fstab"), []byte(fstab), 0644)
	if err != nil {
		PrintVerbose("ABSystem.GenerateFstab:err: %s", err)
		return err
	}

	PrintVerbose("ABSystem.GenerateFstab: fstab generated")
	return nil
}

// GenerateCrypttab identifies which devices are encrypted and generates
// the /etc/crypttab file for the specified root
func (s *ABSystem) GenerateCrypttab(rootPath string) error {
	PrintVerbose("ABSystem.GenerateCrypttab: generating crypttab")

	cryptEntries := [][]string{}

	// Check for encrypted roots
	for _, rootDevice := range s.RootM.Partitions {
		if strings.HasPrefix(rootDevice.Partition.Device, "luks-") {
			parent := rootDevice.Partition.Parent
			PrintVerbose("ABSystem.GenerateCrypttab: Adding %s to crypttab", parent.Device)

			cryptEntries = append(cryptEntries, []string{
				fmt.Sprintf("luks-%s", parent.Uuid),
				fmt.Sprintf("UUID=%s", parent.Uuid),
				"none",
				"luks,discard",
			})
		}
	}

	// Check for encrypted /var
	if strings.HasPrefix(s.RootM.VarPartition.Device, "luks-") {
		parent := s.RootM.VarPartition.Parent
		PrintVerbose("ABSystem.GenerateCrypttab: Adding %s to crypttab", parent.Device)

		cryptEntries = append(cryptEntries, []string{
			fmt.Sprintf("luks-%s", parent.Uuid),
			fmt.Sprintf("UUID=%s", parent.Uuid),
			"none",
			"luks,discard",
		})
	}

	crypttabContent := ""
	for _, entry := range cryptEntries {
		fmtEntry := strings.Join(entry, " ")
		crypttabContent += fmtEntry + "\n"
	}

	err := os.WriteFile(rootPath+"/etc/crypttab", []byte(crypttabContent), 0644)
	if err != nil {
		PrintVerbose("ABSystem.GenerateCrypttab:err(3): %s", err)
		return err
	}

	return nil
}

// GenerateSystemdUnits generates systemd units that mount the mutable parts of the system
func (s *ABSystem) GenerateSystemdUnits(rootPath string, root ABRootPartition) error {
	PrintVerbose("ABSystem.GenerateSystemdUnits: generating units")

	extraDepends := ""
	if root.Partition.IsEncrypted() || s.RootM.VarPartition.IsEncrypted() {
		extraDepends = "cryptsetup.target"
	}

	type varmount struct {
		source      string
		destination string
		fsType      string
		options     string
	}

	mounts := []varmount{
		{"/var/home", "/home", "none", "bind"},
		{"/var/opt", "/opt", "none", "bind"},
		{"/var/lib/abroot/etc/" + root.Label + "/locales", "/.system/usr/lib/locale", "none", "bind"},
		{"overlay", "/.system/etc", "overlay", "lowerdir=/.system/etc,upperdir=/var/lib/abroot/etc/" + root.Label + ",workdir=/var/lib/abroot/etc/" + root.Label + "-work"},
	}

	afterVarTemplate := `[Unit]
Description=Mounts %s from var
After=local-fs-pre.target %s
Before=local-fs.target nss-user-lookup.target
RequiresMountsFor=/var

[Mount]
What=%s
Where=%s
Type=%s
Options=%s
`

	for _, mount := range mounts {
		PrintVerbose("ABSystem.GenerateSystemdUnits: generating unit for %s", mount.destination)

		unit := fmt.Sprintf(afterVarTemplate, mount.destination, extraDepends, mount.source, mount.destination, mount.fsType, mount.options)

		// the unit file needs to have the escaped mount point as its name
		out, err := exec.Command("systemd-escape", "--path", mount.destination).Output()
		if err != nil {
			PrintVerbose("ABSystem.GenerateSystemdUnits:err(1): failed to determine escaped path: %s", err)
			return err
		}
		mountUnitFile := "/" + strings.Replace(string(out), "\n", "", -1) + ".mount"

		err = os.WriteFile(filepath.Join(rootPath, MountUnitDir, mountUnitFile), []byte(unit), 0644)
		if err != nil {
			PrintVerbose("ABSystem.GenerateSystemdUnits:err(2): %s", err)
			return err
		}

		const targetWants string = "/local-fs.target.wants"

		err = os.MkdirAll(filepath.Join(rootPath, MountUnitDir, targetWants), 0755)
		if err != nil {
			PrintVerbose("ABSystem.GenerateSystemdUnits:err(3): %s", err)
			return err
		}

		err = os.Symlink(filepath.Join("../", mountUnitFile), filepath.Join(rootPath, MountUnitDir, targetWants, mountUnitFile))
		if err != nil {
			PrintVerbose("ABSystem.GenerateSystemdUnits:err(4): failed to create symlink: %s", err)
			return err
		}
	}

	PrintVerbose("ABSystem.GenerateSystemdUnits: units generated")
	return nil
}

// RunOperation executes a root-switching operation from the options below:
//
//	UPGRADE: Upgrades to a new image, if available,
//	FORCE_UPGRADE: Forces the upgrade operation, even if no new image is available,
//	APPLY: Applies package changes, but doesn't update the system.
//	INITRAMFS: Updates the initramfs for the future root, but doesn't update the system.
func (s *ABSystem) RunOperation(operation ABSystemOperation) error {
	PrintVerbose("ABSystem.RunOperation: starting %s", operation)

	s.ResetQueue()

	defer s.RunCleanUpQueue("")

	// Stage 0: Check if upgrade is possible
	// -------------------------------------
	PrintVerbose("[Stage 0] -------- ABSystemRunOperation")

	if s.UpgradeLockExists() {
		err := errors.New("upgrades are locked, another is running or need a reboot")
		PrintVerbose("ABSystemRunOperation:err(0): %s", err)
		return err
	}

	err := s.LockUpgrade()
	if err != nil {
		PrintVerbose("ABSystemRunOperation:err(0.1): %s", err)
		return err
	}

	// here we create the stage file, which is used to determine if
	// the upgrade can be interrupted or not. If the stage file is
	// present, it means that the upgrade is in a state where it is
	// still possible to interrupt it, otherwise it is not. This is
	// useful for third-party applications like update managers.
	err = s.CreateStageFile()
	if err != nil {
		PrintVerbose("ABSystemRunOperation:err(0.2): %s", err)
		return err
	}

	s.AddToCleanUpQueue("unlockUpgrade", 200)

	// Stage 1: Check if there is an update available
	// ------------------------------------------------
	PrintVerbose("[Stage 1] -------- ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerbose("ABSystemRunOperation:err(1): %s", err)
		return err
	}

	var imageDigest string
	if operation != APPLY && operation != INITRAMFS {
		var res bool
		imageDigest, res = s.CheckUpdate()
		if !res {
			if operation != FORCE_UPGRADE {
				PrintVerbose("ABSystemRunOperation:err(1.1): %s", err)
				return ErrNoUpdate
			}
			imageDigest = s.CurImage.Digest
			PrintVerbose("ABSystemRunOperation: No update available but --force is set. Proceeding...")
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
	PrintVerbose("[Stage 2] -------- ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerbose("ABSystemRunOperation:err(2): %s", err)
		return err
	}

	partFuture, err := s.RootM.GetFuture()
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(2.1): %s", err)
		return err
	}

	partBoot, err := s.RootM.GetBoot()
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(2.2): %s", err)
		return err
	}

	partFuture.Partition.Unmount() // just in case
	partBoot.Unmount()

	err = partFuture.Partition.Mount("/part-future/")
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(2.3): %s", err)
		return err
	}

	os.RemoveAll("/part-future/.system_new")
	os.RemoveAll("/part-future/abimage-new.abr") // errors are safe to ignore

	s.AddToCleanUpQueue("umountFuture", 90, partFuture)

	_, err = NewIntegrityCheck(partFuture, settings.Cnf.AutoRepair)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(2.4): %s", err)
		return err
	}

	// Stage 3: Make a imageRecipe with user packages
	// ------------------------------------------------
	PrintVerbose("[Stage 3] -------- ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerbose("ABSystemRunOperation:err(3): %s", err)
		return err
	}

	futurePartition, err := s.RootM.GetFuture()
	if err != nil {
		PrintVerbose("ABSystemRunOperation:err(3.1): %s", err)
		return err
	}

	labels := map[string]string{
		"maintainer":  "'Generated by ABRoot'",
		"ABRoot.root": futurePartition.Label,
	}
	args := map[string]string{}
	pkgM := NewPackageManager(false)
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
			PrintVerbose("ABSystemRunOperation:err(3.2): %s", err)
			return err
		}
		imageName, err = RetrieveImageForRoot(presentPartition.Label)
		if err != nil {
			PrintVerbose("ABSystemRunOperation:err(3.3): %s", err)
			return err
		}
		// Handle case where an image for the current root may not exist in storage
		if imageName == "" {
			imageName = settings.Cnf.FullImageName
		}
	default:
		imageName = settings.Cnf.FullImageName + "@" + imageDigest
		labels["ABRoot.BaseImageDigest"] = imageDigest
	}

	imageRecipe := NewImageRecipe(
		imageName,
		labels,
		args,
		content,
	)

	// Stage 4: Extract the rootfs
	// ------------------------------------------------
	PrintVerbose("[Stage 4] -------- ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerbose("ABSystemRunOperation:err(4): %s", err)
		return err
	}

	abrootTrans := filepath.Join(partFuture.Partition.MountPoint, "abroot-trans")
	systemOld := filepath.Join(partFuture.Partition.MountPoint, ".system")
	systemNew := filepath.Join(partFuture.Partition.MountPoint, ".system.new")
	if os.Getenv("ABROOT_FREE_SPACE") != "" {
		PrintVerbose("ABSystemRunOperation: ABROOT_FREE_SPACE is set, deleting future system to free space, this is potentially harmful, assuming we are in a test environment")
		err := os.RemoveAll(systemOld)
		if err != nil {
			PrintVerbose("ABSystemRunOperation:err(4.0): %s", err)
			return err
		}
		err = os.MkdirAll(systemOld, 0755)
		if err != nil {
			PrintVerbose("ABSystemRunOperation:err(4.0.1): %s", err)
			return err
		}
	}

	err = OciExportRootFs(
		"abroot-"+uuid.New().String(),
		imageRecipe,
		abrootTrans,
		systemNew,
	)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(4.1): %s", err)
		return err
	}

	s.AddToCleanUpQueue("clearUnstagedPackages", 10)

	// Stage 5: Write abimage.abr.new to future/
	// ------------------------------------------------
	PrintVerbose("[Stage 5] -------- ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerbose("ABSystemRunOperation:err(5): %s", err)
		return err
	}

	abimage, err := NewABImage(imageDigest, settings.Cnf.FullImageName)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(5.1): %s", err)
		return err
	}

	err = abimage.WriteTo(partFuture.Partition.MountPoint, "new")
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(5.2): %s", err)
		return err
	}

	// from this point on, it is not possible to stop the upgrade
	// so we remove the stage file. Note that interrupting the upgrade
	// from this point on will not leave the system in an inconsistent
	// state, but it could leave the future partition in a dirty state
	// preventing it from booting.
	err = s.RemoveStageFile()
	if err != nil {
		PrintVerbose("ABSystemRunOperation:err(5.3): %s", err)
		return err
	}

	// Stage 6: Generate /etc/fstab, /etc/crypttab, and the
	// SystemD units to setup mountpoints
	// ------------------------------------------------
	PrintVerbose("[Stage 6] -------- ABSystemRunOperation")

	err = s.GenerateFstab(systemNew, partFuture)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(6): %s", err)
		return err
	}

	err = s.GenerateCrypttab(systemNew)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(6.1): %s", err)
		return err
	}

	err = s.GenerateSystemdUnits(systemNew, partFuture)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(6.2): Failed to Generate SystemdUnits: %s", err)
		return err
	}

	// Stage (dry): If dry-run, exit here before writing to disk
	// ------------------------------------------------
	switch operation {
	case DRY_RUN_UPGRADE, DRY_RUN_APPLY, DRY_RUN_INITRAMFS:
		PrintVerbose("ABSystem.RunOperation: dry-run completed")
		return nil
	}

	// Stage 6.3: Delete old image
	err = DeleteImageForRoot(futurePartition.Label)
	if err != nil {
		PrintVerbose("ABSystemRunOperation:err(6.3): %s", err)
		return err
	}

	// Stage 7: Update the bootloader
	// ------------------------------------------------
	PrintVerbose("[Stage 7] -------- ABSystemRunOperation")

	chroot, err := NewChroot(
		systemNew,
		partFuture.Partition.Uuid,
		partFuture.Partition.Device,
	)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(7): %s", err)
		return err
	}

	s.AddToCleanUpQueue("closeChroot", 10, chroot)

	err = chroot.ExecuteCmds( // *1 let grub-mkconfig do its job
		[]string{
			"grub-mkconfig -o /boot/grub/grub.cfg",
			"exit",
		},
	)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(7.1): %s", err)
		return err
	}

	preExec := settings.Cnf.IPkgMngPre
	postExec := settings.Cnf.IPkgMngPost
	initramfsArgs := []string{}
	if preExec != "" {
		initramfsArgs = append(initramfsArgs, preExec)
	}
	initramfsArgs = append(initramfsArgs, "update-initramfs -u")
	if postExec != "" {
		initramfsArgs = append(initramfsArgs, postExec)
	}
	initramfsArgs = append(initramfsArgs, "exit")
	err = chroot.ExecuteCmds(initramfsArgs) // ensure initramfs is updated
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(7.1.1): %s", err)
		return err
	}

	s.RunCleanUpQueue("closeChroot")

	var rootUuid string
	// If Thin-Provisioning set, mount init partition and move linux and initrd images to it
	if settings.Cnf.ThinProvisioning {
		initPartition, err := s.RootM.GetInit()
		if err != nil {
			PrintVerbose("ABSystem.RunOperation:err(7.2): %s", err)
			return err
		}

		initMountpoint := filepath.Join(systemNew, "boot", "init")
		err = initPartition.Mount(initMountpoint)
		if err != nil {
			PrintVerbose("ABSystem.RunOperation:err(7.3): %s", err)
			return err
		}
		s.AddToCleanUpQueue("umountInit", 80)

		kernelVersion := getKernelVersion(filepath.Join(systemNew, "boot"))
		err = CopyFile(
			filepath.Join(systemNew, "boot", "vmlinuz-"+kernelVersion),
			filepath.Join(initMountpoint, partFuture.Label, "vmlinuz-"+kernelVersion),
		)
		if err != nil {
			PrintVerbose("ABSystem.RunOperation:err(7.4): %s", err)
			return err
		}
		err = CopyFile(
			filepath.Join(systemNew, "boot", "initrd.img-"+kernelVersion),
			filepath.Join(initMountpoint, partFuture.Label, "initrd.img-"+kernelVersion),
		)
		if err != nil {
			PrintVerbose("ABSystem.RunOperation:err(7.5): %s", err)
			return err
		}

		os.Remove(filepath.Join(systemNew, "boot", "vmlinuz-"+kernelVersion))
		os.Remove(filepath.Join(systemNew, "boot", "initrd.img-"+kernelVersion))

		rootUuid = initPartition.Uuid
	} else {
		rootUuid = partFuture.Partition.Uuid
	}

	err = generateABGrubConf( // *2 but we don't care about grub.cfg
		systemNew,
		rootUuid,
		partFuture.Label,
	)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(7.6): %s", err)
		return err
	}

	// Stage 8: Sync /etc
	// ------------------------------------------------
	PrintVerbose("[Stage 8] -------- ABSystemRunOperation")

	presentEtc, err := s.RootM.GetPresent()
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(8): %s", err)
		return err
	}
	futureEtc, err := s.RootM.GetFuture()
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(8.1): %s", err)
		return err
	}
	oldUpperEtc := fmt.Sprintf("/var/lib/abroot/etc/%s", presentEtc.Label)
	newUpperEtc := fmt.Sprintf("/var/lib/abroot/etc/%s", futureEtc.Label)

	// Clean new upper etc to prevent deleted files from persisting
	err = os.RemoveAll(newUpperEtc)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(8.2): %s", err)
		return err
	}
	err = os.Mkdir(newUpperEtc, 0755)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(8.3): %s", err)
		return err
	}

	err = s.MergeUserEtcFiles(oldUpperEtc, filepath.Join(systemNew, "/etc"), newUpperEtc)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(8.4): %s", err)
		return err
	}

	s.RunCleanUpQueue("clearUnstagedPackages")

	err = s.SyncUpperEtc(newUpperEtc)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(8.5): %s", err)
		return err
	}

	// Stage 9: Mount boot partition
	// ------------------------------------------------
	PrintVerbose("[Stage 9] -------- ABSystemRunOperation")

	uuid := uuid.New().String()
	tmpBootMount := filepath.Join("/tmp", uuid)
	err = os.Mkdir(tmpBootMount, 0755)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(9): %s", err)
		return err
	}

	err = partBoot.Mount(tmpBootMount)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(9.1): %s", err)
		return err
	}

	s.AddToCleanUpQueue("umountBoot", 100, partBoot)

	// Stage 10: Atomic swap the rootfs and abimage.abr
	// ------------------------------------------------
	PrintVerbose("[Stage 10] -------- ABSystemRunOperation")

	err = AtomicSwap(systemOld, systemNew)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(10): %s", err)
		return err
	}

	s.AddToCleanUpQueue("removeNewSystem", 20, systemNew)

	oldABImage := filepath.Join(partFuture.Partition.MountPoint, "abimage.abr")
	newABImage := filepath.Join(partFuture.Partition.MountPoint, "abimage-new.abr")

	// PartFuture may not have /abimage.abr if it got corrupted or was wiped.
	// In these cases, create a dummy file for the atomic swap.
	if _, err = os.Stat(oldABImage); os.IsNotExist(err) {
		PrintVerbose("ABSystem.RunOperation: Creating dummy /part-future/abimage.abr")
		os.Create(oldABImage)
	}

	err = AtomicSwap(oldABImage, newABImage)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(10.1): %s", err)
		return err
	}

	s.AddToCleanUpQueue("removeNewABImage", 30, newABImage)

	// Stage 11: Atomic swap the bootloader
	// ------------------------------------------------
	PrintVerbose("[Stage 11] -------- ABSystemRunOperation")

	grub, err := NewGrub(partBoot)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(11): %s", err)
		return err
	}

	// Only swap grub entries if we're booted into the present partition
	isPresent, err := grub.IsBootedIntoPresentRoot()
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(11.1): %s", err)
		return err
	}
	if isPresent {
		grubCfgCurrent := filepath.Join(tmpBootMount, "grub/grub.cfg")
		grubCfgFuture := filepath.Join(tmpBootMount, "grub/grub.cfg.future")

		// Just like in Stage 10, tmpBootMount/grub/grub.cfg.future may not exist.
		if _, err = os.Stat(grubCfgFuture); os.IsNotExist(err) {
			PrintVerbose("ABSystem.RunOperation: Creating grub.cfg.future")

			grubCfgContents, err := os.ReadFile(grubCfgCurrent)
			if err != nil {
				PrintVerbose("ABSystem.RunOperation:err(11.2): %s", err)
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
			os.WriteFile(grubCfgFuture, []byte(replacer.Replace(string(grubCfgContents))), 0644)
		}

		err = AtomicSwap(grubCfgCurrent, grubCfgFuture)
		if err != nil {
			PrintVerbose("ABSystem.RunOperation:err(11.3): %s", err)
			return err
		}
	}

	PrintVerbose("ABSystem.RunOperation: upgrade completed")
	return nil
}

// Rollback swaps the master grub files if the current root is not the default
func (s *ABSystem) Rollback() (response ABRollbackResponse, err error) {
	PrintVerbose("ABSystem.Rollback: starting")

	s.ResetQueue()

	defer s.RunCleanUpQueue("")

	// we won't allow upgrades while rolling back
	err = s.LockUpgrade()
	if err != nil {
		PrintVerbose("ABSystem.Rollback:err(0): %s", err)
		return ROLLBACK_FAILED, err
	}

	partBoot, err := s.RootM.GetBoot()
	if err != nil {
		PrintVerbose("ABSystem.Rollback:err(1): %s", err)
		return ROLLBACK_FAILED, err
	}

	uuid := uuid.New().String()
	tmpBootMount := filepath.Join("/tmp", uuid)
	err = os.Mkdir(tmpBootMount, 0755)
	if err != nil {
		PrintVerbose("ABSystem.Rollback:err(2): %s", err)
		return ROLLBACK_FAILED, err
	}

	err = partBoot.Mount(tmpBootMount)
	if err != nil {
		PrintVerbose("ABSystem.Rollback:err(3): %s", err)
		return ROLLBACK_FAILED, err
	}

	s.AddToCleanUpQueue("umountBoot", 100, partBoot)

	grub, err := NewGrub(partBoot)
	if err != nil {
		PrintVerbose("ABSystem.Rollback:err(4): %s", err)
		return ROLLBACK_FAILED, err
	}

	// Only swap grub entries if we're booted into the present partition
	isPresent, err := grub.IsBootedIntoPresentRoot()
	if err != nil {
		PrintVerbose("ABSystem.Rollback:err(5): %s", err)
		return ROLLBACK_FAILED, err
	}

	if isPresent {
		PrintVerbose("ABSystem.Rollback: current root is the default, nothing to do")
		return ROLLBACK_UNNECESSARY, nil
	}

	grubCfgCurrent := filepath.Join(tmpBootMount, "grub/grub.cfg")
	grubCfgFuture := filepath.Join(tmpBootMount, "grub/grub.cfg.future")

	// Just like in Stage 10, tmpBootMount/grub/grub.cfg.future may not exist.
	if _, err = os.Stat(grubCfgFuture); os.IsNotExist(err) {
		PrintVerbose("ABSystem.Rollback: Creating grub.cfg.future")

		grubCfgContents, err := os.ReadFile(grubCfgCurrent)
		if err != nil {
			PrintVerbose("ABSystem.Rollback:err(6): %s", err)
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
		os.WriteFile(grubCfgFuture, []byte(replacer.Replace(string(grubCfgContents))), 0644)
	}

	err = AtomicSwap(grubCfgCurrent, grubCfgFuture)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(7): %s", err)
		return ROLLBACK_FAILED, err
	}

	PrintVerbose("ABSystem.Rollback: rollback completed")
	return ROLLBACK_SUCCESS, nil
}

func (s *ABSystem) UserLockRequested() bool {
	if _, err := os.Stat(userLockFile); os.IsNotExist(err) {
		return false
	}

	PrintVerbose("ABSystem.UserLockRequested: lock file exists")
	return true
}

func (s *ABSystem) UpgradeLockExists() bool {
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		return false
	}

	PrintVerbose("ABSystem.UpgradeLockExists: lock file exists")
	return true
}

func (s *ABSystem) LockUpgrade() error {
	_, err := os.Create(lockFile)
	if err != nil {
		PrintVerbose("ABSystem.LockUpgrade: %s", err)
		return err
	}

	PrintVerbose("ABSystem.LockUpgrade: lock file created")
	return nil
}

func (s *ABSystem) UnlockUpgrade() error {
	err := os.Remove(lockFile)
	if err != nil {
		PrintVerbose("ABSystem.UnlockUpgrade: %s", err)
		return err
	}

	PrintVerbose("ABSystem.UnlockUpgrade: lock file removed")
	return nil
}

func (s *ABSystem) CreateStageFile() error {
	_, err := os.Create(stageFile)
	if err != nil {
		PrintVerbose("ABSystem.CreateStageFile: %s", err)
		return err
	}

	PrintVerbose("ABSystem.CreateStageFile: stage file created")
	return nil
}

func (s *ABSystem) RemoveStageFile() error {
	err := os.Remove(stageFile)
	if err != nil {
		PrintVerbose("ABSystem.RemoveStageFile: %s", err)
		return err
	}

	PrintVerbose("ABSystem.RemoveStageFile: stage file removed")
	return nil
}
