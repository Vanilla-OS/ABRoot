package cmd

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
	"fmt"
	"os"
	"slices"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/orchid/cmdr"
)

type DiskLayoutError struct {
	Device string
}

func (e *DiskLayoutError) Error() string {
	return fmt.Sprintf("device %s has an unsupported layout", e.Device)
}

type PartNotFoundError struct {
	Partition string
}

func (e *PartNotFoundError) Error() string {
	return fmt.Sprintf("partition %s could not be found", e.Partition)
}

func NewMountSysCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"mount-sys",
		"",
		"",
		mountSysCmd,
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"dry-run",
			"d",
			"perform a dry run of the operation",
			false,
		),
	)

	cmd.Example = "abroot mount-sys"

	cmd.Hidden = true

	return cmd
}

// helper function which only returns syntax errors and prints other ones
func mountSysCmd(cmd *cobra.Command, args []string) error {
	err := mountSys(cmd, args)
	if err != nil {
		cmdr.Error.Println(err)
		os.Exit(1)
	}
	return nil
}

func mountSys(cmd *cobra.Command, _ []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println("This operation requires root.")
		return nil
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	manager := core.NewABRootManager()
	present, err := manager.GetPresent()
	if err != nil {
		return err
	}

	// remount as writeable
	if !dryRun {
		err := syscall.Mount("/", "/", "", syscall.MS_REMOUNT, "")
		if err != nil {
			cmdr.Error.Println("failed to remount root", err)
		}
	}

	err = mountVar(manager.VarPartition, dryRun)
	if err != nil {
		cmdr.Error.Println(err)
		os.Exit(5)
	}

	if !dryRun {
		err = core.RepairRootIntegrity("/")
		if err != nil {
			cmdr.Error.Println(err)
			os.Exit(4)
		}
	}

	err = mountBindMounts(dryRun)
	if err != nil {
		cmdr.Error.Println(err)
		os.Exit(6)
	}

	if present.Label == "" {
		return &PartNotFoundError{"current root"}
	}
	err = mountOverlayMounts(present.Label, dryRun)
	if err != nil {
		cmdr.Error.Println(err)
		os.Exit(7)
	}

	if present.Uuid == "" {
		return &PartNotFoundError{"current root"}
	}
	err = adjustFstab(present.Uuid, dryRun)
	if err != nil {
		cmdr.Error.Println(err)
		os.Exit(8)
	}

	if dryRun {
		cmdr.Info.Println("Dry run complete.")
	} else {
		cmdr.Info.Println("The system mounts have been performed successfully.")
	}

	return nil
}

func mountVar(varPart core.Partition, dryRun bool) error {
	cmdr.FgDefault.Println("mounting " + varPart.Device + " in /var")

	if varPart.Device == "" {
		return &PartNotFoundError{settings.Cnf.PartLabelVar}
	}

	if !dryRun {
		err := varPart.Mount("/var")
		if err != nil {
			return err
		}
	}

	return nil
}

func mountBindMounts(dryRun bool) error {
	type bindMount struct {
		from, to string
		options  uintptr
	}

	binds := []bindMount{
		{"/.system/usr", "/.system/usr", syscall.MS_RDONLY},
	}

	for _, bind := range binds {
		cmdr.FgDefault.Println("bind-mounting " + bind.from + " to " + bind.to)
		if !dryRun {
			err := syscall.Mount(bind.from, bind.to, "", syscall.MS_BIND|bind.options, "")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func mountOverlayMounts(rootLabel string, dryRun bool) error {
	type overlayMount struct {
		destination       string
		lowerdirs         []string
		upperdir, workdir string
	}

	overlays := []overlayMount{
		{"/etc", []string{"/etc"}, "/var/lib/abroot/etc/" + rootLabel, "/var/lib/abroot/etc/" + rootLabel + "-work"},
		{"/opt", []string{"/opt"}, "/var/opt", "/var/opt-work"},
	}

	for _, overlay := range overlays {
		if _, err := os.Lstat(overlay.workdir); os.IsNotExist(err) {
			err := os.MkdirAll(overlay.workdir, 0o755)
			cmdr.Warning.Println(err)
			// failing the boot here won't help so ingore any error
		}

		lowerCombined := strings.Join(overlay.lowerdirs, ":")
		options := "lowerdir=" + lowerCombined + ",upperdir=" + overlay.upperdir + ",workdir=" + overlay.workdir

		cmdr.FgDefault.Println("mounting overlay mount " + overlay.destination + " with options " + options)

		if !dryRun {
			err := syscall.Mount("overlay", overlay.destination, "overlay", 0, options)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func adjustFstab(uuid string, dryRun bool) error {
	cmdr.FgDefault.Println("switching the root in fstab")

	const fstabFile = "/etc/fstab"
	systemMounts := []string{"/", "/var", "/usr", "/etc"}
	varBindMountExists := map[string]bool{
		"/home":  false,
		"/media": false,
		"/mnt":   false,
		"/root":  false,
	}

	fstabContentsRaw, err := os.ReadFile(fstabFile)
	if err != nil {
		return err
	}

	fstabContents := string(fstabContentsRaw)

	lines := strings.Split(fstabContents, "\n")

	linesNew := make([]string, 0, len(lines))

	for _, line := range lines {

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			linesNew = append(linesNew, line)
			continue
		}

		words := strings.Fields(line)
		if len(words) < 2 {
			linesNew = append(linesNew, line)
			continue
		}

		mountpoint := words[1]

		if _, ok := varBindMountExists[mountpoint]; ok {
			varBindMountExists[mountpoint] = true
			linesNew = append(linesNew, line)
			continue
		}

		// mounting to /var/home for example worked in the past
		// this migrates those lines to use /home instead
		if mountpointWithoutVar, found := strings.CutPrefix(mountpoint, "/var"); found {
			if _, ok := varBindMountExists[mountpointWithoutVar]; ok {
				cmdr.FgDefault.Println("Removing /var prefix from mount", mountpoint)

				varBindMountExists[mountpointWithoutVar] = true
				words[1] = mountpointWithoutVar
				lineNew := strings.Join(words, " ")
				linesNew = append(linesNew, lineNew)
				continue
			}
		}

		if slices.Contains(systemMounts, mountpoint) {
			cmdr.FgDefault.Println("Deleting line: ", line)
			continue
		}

		linesNew = append(linesNew, line)
	}

	for varBindMount, existsAlready := range varBindMountExists {
		if existsAlready {
			continue
		}

		newVarBindMountLine := fmt.Sprintf("/var%s %s none defaults,bind 0 0", varBindMount, varBindMount)

		cmdr.FgDefault.Println("Adding line: ", newVarBindMountLine)

		linesNew = append([]string{newVarBindMountLine}, linesNew...)
	}

	currentRootLine := "UUID=" + uuid + " / btrfs defaults 0 0"

	cmdr.FgDefault.Println("Adding line: ", currentRootLine)

	linesNew = append([]string{currentRootLine}, linesNew...)

	newFstabContents := strings.Join(linesNew, "\n")

	newFstabFile := fstabFile + ".new"

	if !dryRun {
		cmdr.FgDefault.Println("writing new fstab file")
		err := os.WriteFile(newFstabFile, []byte(newFstabContents), 0o644)
		if err != nil {
			return err
		}
		err = core.AtomicSwap(fstabFile, newFstabFile)
		if err != nil {
			return err
		}
		err = os.Rename(newFstabFile, fstabFile+".old")
		if err != nil {
			cmdr.Warning.Println("Old Fstab file will keep .new suffix")
			// ignore, backup is not neccessary to boot
		}
	}

	return nil
}
