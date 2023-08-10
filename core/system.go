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
	UPGRADE      = "upgrade"
	FOCE_UPGRADE = "force-upgrade"
	APPLY        = "package-apply"
)

const (
	MountScriptPath  = "/usr/sbin/.abroot-mountpoints"
	MountUnitDir     = "/etc/systemd/system"
	SystemDTargetDir = "/etc/systemd/system/cryptsetup.target.wants"
	MountUnitFile    = "/abroot-mount.service"
)

type ABSystemOperation string

var (
	queue         []QueuedFunction
	lockFile      string = filepath.Join("/tmp", "ABSystem.Upgrade.lock")
	userLockFile  string = filepath.Join("/tmp", "ABSystem.Upgrade.user.lock")
	NoUpdateError error  = errors.New("no update available")
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

// SyncLowerEtc syncs the lower etc dir (/.system/etc)
func (s *ABSystem) SyncLowerEtc(newEtc string) error {
	PrintVerbose("ABSystem.SyncLowerEtc: syncing /.system/etc -> %s", newEtc)

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
		sourceFile := etcDir + "/" + file
		destFile := newEtc + "/" + file

		// write the diff to the destination
		err := MergeDiff(sourceFile, destFile)
		if err != nil {
			PrintVerbose("ABSystem.SyncLowerEtc:err(2): %s", err)
			return err
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

	for _, f := range queue {
		if fnName != "" && f.Name != fnName {
			continue
		}

		switch f.Name {
		case "umountFuture":
			futurePart := f.Values[0].(ABRootPartition)
			err := futurePart.Partition.Unmount()
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err: %s", err)
				return err
			}
		case "closeChroot":
			chroot := f.Values[0].(*Chroot)
			chroot.Close() // safe to ignore, already closed
		case "removeNewSystem":
			newSystem := f.Values[0].(string)
			err := os.RemoveAll(newSystem)
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(3): %s", err)
				return err
			}
		case "removeNewABImage":
			newImage := f.Values[0].(string)
			err := os.RemoveAll(newImage)
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(4): %s", err)
				return err
			}
		case "umountBoot":
			bootPart := f.Values[0].(Partition)
			err := bootPart.Unmount()
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(5): %s", err)
				return err
			}
		case "unlockUpgrade":
			err := s.UnlockUpgrade()
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(6): %s", err)
				return err
			}
		case "clearUnstagedPackages":
			pkgM := NewPackageManager()
			err := pkgM.ClearUnstagedPackages()
			if err != nil {
				PrintVerbose("ABSystem.RunCleanUpQueue:err(7): %s", err)
				return err
			}
		}
	}

	s.ResetQueue()

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
`
	fstab := fmt.Sprintf(
		template,
		root.Partition.Uuid,
		root.Partition.FsType,
	)

	err := os.WriteFile(rootPath+"/etc/fstab", []byte(fstab), 0644)
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

// GenerateMountpointsScript generates a /usr/sbin/.abroot-mountpoints file for the future root
func (s *ABSystem) GenerateMountpointsScript(rootPath string, root ABRootPartition) error {
	PrintVerbose("ABSystem.GenerateMountpointsScript: generating script")

	template := `#!/usr/bin/bash
echo "ABRoot: Initializing mount points..."

# /var mount
mount %s /var

# /etc overlay
mount -t overlay overlay -o lowerdir=/.system/etc,upperdir=/var/lib/abroot/etc/%s,workdir=/var/lib/abroot/etc/%s-work /etc

# /var binds
mount -o bind /var/home /home
mount -o bind /var/opt /opt
mount -o bind,ro /.system/usr /usr
`
	mountExtCmd := ""
	if strings.HasPrefix(s.RootM.VarPartition.Device, "luks-") {
		parent := s.RootM.VarPartition.Parent
		mountExtCmd = fmt.Sprintf("/dev/mapper/luks-%s", parent.Uuid)
	} else {
		mountExtCmd = fmt.Sprintf("-U %s", s.RootM.VarPartition.Uuid)
	}
	mountpoints := fmt.Sprintf(template, mountExtCmd, root.Label, root.Label)

	err := os.WriteFile(rootPath+MountScriptPath, []byte(mountpoints), 0755)
	if err != nil {
		PrintVerbose("ABSystem.GenerateMountpointsScript:err(3): %s", err)
		return err
	}

	err = os.Chmod(rootPath+MountScriptPath, 0755)
	if err != nil {
		PrintVerbose("ABSystem.GenerateMountpointsScript:err(4): %s", err)
		return err
	}

	PrintVerbose("ABSystem.GenerateMountpointsScript: script generated")
	return nil
}

// GenerateMountpointsSystemDUnit creates a SystemD unit to setup the /var partition and its
// derivatives such as /home, /opt, and /etc.
// This unit must run as soon as all encrypted partitions (if any) have been decrupted.
func (s *ABSystem) GenerateMountpointsSystemDUnit(rootPath string, root ABRootPartition) error {
	PrintVerbose("ABSystem.GenerateMountpointsSystemDUnit: generating script")

	template := `[Unit]
Description=Mount partitions
Requires=cryptsetup.target
After=cryptsetup.target

[Service]
Type=oneshot
ExecStart=%s
`
	unit := fmt.Sprintf(template, MountScriptPath)

	err := os.WriteFile(rootPath+MountUnitDir+MountUnitFile, []byte(unit), 0755)
	if err != nil {
		PrintVerbose("ABSystem.GenerateMountpointsSystemDUnit:err(2): %s", err)
		return err
	}

	err = os.MkdirAll(rootPath+SystemDTargetDir, 0755)
	if err != nil {
		PrintVerbose("ABSystem.GenerateMountpointsSystemDUnit:err(3): %s", err)
		return err
	}

	err = os.Symlink(rootPath+MountUnitDir+MountUnitFile, rootPath+SystemDTargetDir+MountUnitFile)
	if err != nil {
		PrintVerbose("ABSystem.GenerateMountpointsSystemDUnit:err(4): %s", err)
		return err
	}

	PrintVerbose("ABSystem.GenerateMountpointsSystemDUnit: unit generated")
	return nil
}

// RunOperation executes a root-switching operation from the options below:
//
//	UPGRADE: Upgrades to a new image, if available,
//	FORCE_UPGRADE: Forces the upgrade operation, even if no new image is available,
//	APPLY: Applies package changes, but doesn't update the system.
func (s *ABSystem) RunOperation(operation ABSystemOperation) error {
	PrintVerbose("ABSystem.RunOperation: starting %s", operation)

	s.ResetQueue()

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
	if operation != APPLY {
		var res bool
		imageDigest, res = s.CheckUpdate()
		if !res {
			if operation != FOCE_UPGRADE {
				PrintVerbose("ABSystemRunOperation:err(1.1): %s", err)
				s.RunCleanUpQueue("")
				return NoUpdateError
			}
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
		PrintVerbose("ABSystem.RunOperation:err(2.3: %s", err)
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
	pkgM := NewPackageManager()
	pkgsFinal := pkgM.GetFinalCmd(operation)
	if pkgsFinal == "" {
		pkgsFinal = "true"
	}
	content := `RUN ` + pkgsFinal

	var imageName string
	if operation == APPLY {
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
	} else {
		imageName = strings.Split(settings.Cnf.FullImageName, ":")[0] + "@" + imageDigest
		labels["ABRoot.BaseImageDigest"] = s.CurImage.Digest
	}

	// Delete old image
	err = DeleteImageForRoot(futurePartition.Label)
	if err != nil {
		PrintVerbose("ABSystemRunOperation:err(3.4): %s", err)
		return err
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
	PrintVerbose("[Stage 5] ABSystemRunOperation")

	if s.UserLockRequested() {
		err := errors.New("upgrade locked per user request")
		PrintVerbose("ABSystemRunOperation:err(5): %s", err)
		return err
	}

	abimage := NewABImage(imageDigest, settings.Cnf.FullImageName)
	err = abimage.WriteTo(partFuture.Partition.MountPoint, "new")
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(5.1): %s", err)
		return err
	}

	// Stage 6: Generate /etc/fstab, /etc/crypttab, /usr/sbin/.abroot-mountpoints, and the
	// SystemD unit to setup mountpoints
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

	err = s.GenerateMountpointsScript(systemNew, partFuture)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(6.2): %s", err)
		return err
	}

	err = s.GenerateMountpointsSystemDUnit(systemNew, partFuture)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(6.3): %s", err)
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

	s.RunCleanUpQueue("closeChroot")
	s.RemoveFromCleanUpQueue("closeChroot")

	err = generateABGrubConf( // *2 but we don't care about grub.cfg
		systemNew,
		partFuture.Partition.Uuid,
	)
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(7.2): %s", err)
		return err
	}

	// Stage 8: Sync /etc
	// ------------------------------------------------
	PrintVerbose("[Stage 8] -------- ABSystemRunOperation")

	err = s.SyncLowerEtc(filepath.Join(systemNew, "/etc"))
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(8): %s", err)
		return err
	}

	futureEtc, err := s.RootM.GetFuture()
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(8.1): %s", err)
		return err
	}

	err = s.SyncUpperEtc(fmt.Sprintf("/var/lib/abroot/etc/%s", futureEtc.Label))
	if err != nil {
		PrintVerbose("ABSystem.RunOperation:err(8.2): %s", err)
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

	if grub.futureRoot != partFuture.Label {
		grubCfgCurrent := filepath.Join(tmpBootMount, "grub/grub.cfg")
		grubCfgFuture := filepath.Join(tmpBootMount, "grub/grub.cfg.future")

		// Just like in Stage 10, tmpBootMount/grub/grub.cfg.future may not exist.
		if _, err = os.Stat(grubCfgFuture); os.IsNotExist(err) {
			PrintVerbose("ABSystem.RunOperation: Creating grub.cfg.future")

			grubCfgContents, err := os.ReadFile(grubCfgCurrent)
			if err != nil {
				PrintVerbose("ABSystem.RunOperation:err(11.1): %s", err)
			}

			var replacerPairs []string
			if grub.futureRoot == "a" {
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
			PrintVerbose("ABSystem.RunOperation:err(11.2): %s", err)
			return err
		}
	}

	PrintVerbose("ABSystem.RunOperation: upgrade completed")
	s.RunCleanUpQueue("")
	return nil
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
