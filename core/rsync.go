package core

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"golang.org/x/sys/unix"
)

// rsyncCmd executes the rsync command with the requested options.
func rsyncCmd(src, dst string, opts []string) error {
	args := []string{"-avxHAX"}
	args = append(args, opts...)
	args = append(args, src)
	args = append(args, dst)

	cmd := exec.Command("rsync", args...)
	stdout, _ := cmd.StdoutPipe()

	reader := bufio.NewReader(stdout)

	count_cmd_out, _ := exec.Command(
		"/bin/sh",
		"-c",
		fmt.Sprintf("echo -n $(($(rsync -avxHAX --dry-run %s %s | wc -l) - 4))", src, dst),
	).Output()
	total_files, _ := strconv.Atoi(string(count_cmd_out))

	p, _ := pterm.DefaultProgressbar.WithTotal(int(total_files)).WithTitle("Sync in progress").Start()

	err := cmd.Start()
	if err != nil {
		return err
	}

	max_line_len := int(pterm.GetTerminalWidth() / 4)
	pterm.Info.Println(max_line_len)
	for i := 0; i < p.Total; i++ {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if len(line) > max_line_len {
			starting_len := len(line) - max_line_len + 1
			line = "…" + line[starting_len:]
		} else {
			padding := max_line_len - len(line)
			line = line + strings.Repeat(" ", padding)
		}

		p.UpdateTitle("Syncing " + line)
		p.Increment()
	}

	err = cmd.Wait()
	if err != nil {
		return err
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

	return rsyncCmd(src, dst, opts)
}

// atomicSwap allows swapping 2 files or directories in-place and atomically, using
// the renameat2 syscall.
func atomicSwap(src, dst string) error {
	orig, err := os.Open(src)
	if err != nil {
		PrintVerbose("err:atomicSwap: %s", err)
		return err
	}

	new, err := os.Open(dst)
	if err != nil {
		PrintVerbose("err:atomicSwap: %s", err)
		return err
	}

	PrintVerbose("step:  Renameat2")
	err = unix.Renameat2(int(orig.Fd()), src, int(new.Fd()), dst, unix.RENAME_EXCHANGE)
	if err != nil {
		PrintVerbose("err:atomicSwap: %s", err)
		return err
	}

	return nil
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
	if err := rsyncDryRun(src, transitionalPath, excluded); err != nil {
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
	err := rsyncCmd(src, transitionalPath, opts)
	if err != nil {
		return err
	}

	PrintVerbose("step:  atomicSwap")
	err = atomicSwap(transitionalPath, finalPath)
	if err != nil {
		return err
	}

	PrintVerbose("step:  RemoveAll")
	return os.RemoveAll(transitionalPath)
}
