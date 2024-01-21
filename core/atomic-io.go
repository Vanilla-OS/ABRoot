package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
		Luca di Maio <https://github.com/89luca89>
	Copyright: 2024
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

import (
	"os"

	"golang.org/x/sys/unix"
)

// atomicSwap allows swapping 2 files or directories in-place and atomically,
// using the renameat2 syscall. This should be used instead of os.Rename,
// which is not atomic at all
func AtomicSwap(src, dst string) error {
	PrintVerboseInfo("AtomicSwap", "running...")

	orig, err := os.Open(src)
	if err != nil {
		PrintVerboseErr("AtomicSwap", 0, err)
		return err
	}

	newfile, err := os.Open(dst)
	if err != nil {
		PrintVerboseErr("AtomicSwap", 1, err)
		return err
	}

	err = unix.Renameat2(int(orig.Fd()), src, int(newfile.Fd()), dst, unix.RENAME_EXCHANGE)
	if err != nil {
		PrintVerboseErr("AtomicSwap", 2, err)
		return err
	}

	PrintVerboseInfo("AtomicSwap", "done")
	return nil
}
