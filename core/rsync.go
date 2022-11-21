package core

import (
	"os"
	"os/exec"
)

func rsyncCmd(src, dst string, opts []string) error {
	/*
	 * rsyncCmd executes the rsync command with the requested options.
	 */
	args := []string{"-a"}
	args = append(args, opts...)
	args = append(args, src)
	args = append(args, dst)

	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func RsyncDryRun(src, dst string, excluded []string) error {
	/*
	 * rsyncDryRun executes the rsync command with the --dry-run option.
	 */
	opts := []string{"--dry-run"}

	if len(excluded) > 0 {
		for _, exclude := range excluded {
			opts = append(opts, "--exclude="+exclude)
		}
	}

	return rsyncCmd(src, dst, []string{"--dry-run"})
}

func AtomicRsync(src, dst string, excluded []string, keepUnwanted bool) error {
	/*
	 * AtomicRsync executes the rsync command in an atomic-like manner.
	 * It does so by dry-running the rsync, and if it succeeds, it runs
	 * the rsync again performing changes. If the keepUnwanted option
	 * is set to true, it will omit the --delete option, so that the already
	 * existing and unwanted files will not be deleted.
	 */
	if err := RsyncDryRun(src, dst, excluded); err != nil {
		return err
	}

	opts := []string{}

	if len(excluded) > 0 {
		for _, exclude := range excluded {
			opts = append(opts, "--exclude", exclude)
		}
	}

	if !keepUnwanted {
		opts = append(opts, "--delete")
	}

	return rsyncCmd(src, dst, opts)
}
