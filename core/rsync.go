package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
		Luca di Maio <https://github.com/89luca89>
		Mateus B. Melchiades <https://github.com/matbme>
	Copyright: 2023
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/vanilla-os/orchid/cmdr"
)

// rsyncCmd executes the rsync command with the requested options.
// If silent is true, rsync progress will not appear in stdout.
func rsyncCmd(src, dst string, opts []string, silent bool) error {
	args := []string{"-avxHAX"}
	args = append(args, opts...)
	args = append(args, src)
	args = append(args, dst)

	cmd := exec.Command("rsync", args...)
	stdout, _ := cmd.StdoutPipe()

	var totalFiles int

	if !silent {
		countCmdOut, _ := exec.Command(
			"/bin/sh",
			"-c",
			fmt.Sprintf("echo -n $(($(rsync --dry-run %s | wc -l) - 4))", strings.Join(args, " ")),
		).Output()
		totalFiles, _ = strconv.Atoi(string(countCmdOut))
	}

	reader := bufio.NewReader(stdout)

	err := cmd.Start()
	if err != nil {
		return err
	}

	if !silent {
		verbose := IsVerbose()

		p, _ := cmdr.ProgressBar.WithTotal(totalFiles).WithTitle("Sync in progress").WithMaxWidth(120).Start()
		maxLineLen := cmdr.TerminalWidth() / 4

		for i := 0; i < p.Total; i++ {
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)

			if verbose {
				cmdr.Info.Println(line + " synced")
			}

			if len(line) > maxLineLen {
				startingLen := len(line) - maxLineLen + 1
				line = "<" + line[startingLen:]
			} else {
				padding := maxLineLen - len(line)
				line += strings.Repeat(" ", padding)
			}

			p.UpdateTitle("Syncing " + line)
			p.Increment()
		}
	} else {
		stdout.Close()
	}

	err = cmd.Wait()
	if err != nil {
		// exit status 24 is a warning, not an error, we don't care about it
		// since rsync is going to be removed in the OCI version
		if !strings.Contains(err.Error(), "exit status 24") {
			return err
		}
	}

	return nil
}

// rsyncDryRun executes the rsync command with the --dry-run option.
func rsyncDryRun(src, dst string, excluded []string) error {
	opts := []string{"--dry-run"}

	if len(excluded) > 0 {
		for _, exclude := range excluded {
			opts = append(opts, "--exclude="+exclude)
		}
	}

	return rsyncCmd(src, dst, opts, false)
}

// AtomicRsync executes the rsync command in an atomic-like manner.
// It does so by dry-running the rsync, and if it succeeds, it runs
// the rsync again performing changes.
// If the keepUnwanted option
// is set to true, it will omit the --delete option, so that the already
// existing and unwanted files will not be deleted.
// To ensure the changes are applied atomically, we rsync on a _new directory first,
// and use atomicSwap to replace the _new with the dst directory.
func AtomicRsync(src, dst string, transitionalPath string, finalPath string, excluded []string, keepUnwanted bool) error {
	if _, err := os.Stat(transitionalPath); os.IsNotExist(err) {
		err = os.Mkdir(transitionalPath, 0755)
		if err != nil {
			PrintVerbose("err:AtomicRsync: %s", err)
			return err
		}
	}

	PrintVerbose("step:  rsyncDryRun")

	err := rsyncDryRun(src, transitionalPath, excluded)
	if err != nil {
		return err
	}

	opts := []string{"--link-dest", dst, "--exclude", finalPath, "--exclude", transitionalPath}

	if len(excluded) > 0 {
		for _, exclude := range excluded {
			opts = append(opts, "--exclude", exclude)
		}
	}

	if !keepUnwanted {
		opts = append(opts, "--delete")
	}

	PrintVerbose("step:  rsyncCmd")

	err = rsyncCmd(src, transitionalPath, opts, true)
	if err != nil {
		return err
	}

	PrintVerbose("step:  atomicSwap")

	err = AtomicSwap(transitionalPath, finalPath)
	if err != nil {
		return err
	}

	PrintVerbose("step:  RemoveAll")

	return os.RemoveAll(transitionalPath)
}
