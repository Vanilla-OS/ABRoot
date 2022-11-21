package core

import (
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

// rsyncCmd executes the rsync command with the requested options.
func rsyncCmd(src, dst string, opts []string) error {
	args := []string{"-a"}
	args = append(args, opts...)
	args = append(args, src)
	args = append(args, dst)

	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// rsyncDryRun executes the rsync command with the --dry-run option.
func rsyncDryRun(src, dst string, excluded []string) error {
	opts := []string{"--dry-run"}

	if len(excluded) > 0 {
		for _, exclude := range excluded {
			opts = append(opts, "--exclude="+exclude)
		}
	}

	return rsyncCmd(src, dst, []string{"--dry-run"})
}

// atomicSwap allows swapping 2 files or directories in-place and atomically, using
// the renameat2 syscall.
func atomicSwap(src, dst string) error {
	orig, err := os.Open(src)
	if err != nil {
		return err
	}

	new, err := os.Open(dst)
	if err != nil {
		return err
	}

	err = unix.Renameat2(int(orig.Fd()), src, int(new.Fd()), dst, unix.RENAME_EXCHANGE)
	if err != nil {
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
func AtomicRsync(src, dst string, excluded []string, keepUnwanted bool) error {
	if err := rsyncDryRun(src, dst+"_new", excluded); err != nil {
		return err
	}

	opts := []string{"--link-dest", dst}

	if len(excluded) > 0 {
		for _, exclude := range excluded {
			opts = append(opts, "--exclude", exclude)
		}
	}

	if !keepUnwanted {
		opts = append(opts, "--delete")
	}

	err := rsyncCmd(src, dst+"_new", opts)
	if err != nil {
		return err
	}

	err = atomicSwap(dst, dst+"_new")
	if err != nil {
		return err
	}

	return os.RemoveAll(dst + "_new")
}
