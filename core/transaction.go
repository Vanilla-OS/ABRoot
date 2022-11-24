package core

import (
	"fmt"
	"os"
)

var (
	lockPath = "/tmp/abroot-transactions.lock"
)

// LockTransaction locks the transactional shell.
// It does so by creating a lock file.
func LockTransaction() error {
	if AreTransactionsLocked() {
		return nil
	}

	f, err := os.Create(lockPath)
	if err != nil {
		return err
	}

	f.Close()

	return nil
}

// UnlockTransaction unlocks the transactional shell.
// It does so by removing the lock file.
func UnlockTransaction() error {
	if _, err := os.Stat(lockPath); err != nil {
		return nil // transactions are already unlocked
	}

	err := os.Remove(lockPath)
	if err != nil {
		fmt.Printf("failed to remove lock file: %s", err)
		return err
	}

	return nil
}

// IsLocked returns true if the transactional shell is locked.
func AreTransactionsLocked() bool {
	if _, err := os.Stat(lockPath); err != nil {
		return false
	}

	return true
}

// NewTransaction starts a new transaction.
// It does so by creating a new overlayfs with the current root as the
// lower layer, and then locking the transactional shell.
func NewTransaction() error {
	if AreTransactionsLocked() {
		fmt.Println("Transactions are locked, another one is already running or a reboot is required.")
		return nil
	}

	if err := NewOverlayFS([]string{"/"}); err != nil {
		if err := UnlockTransaction(); err != nil {
			fmt.Printf("failed to unlock transactions: %s", err)
		}
		return err
	}

	if err := LockTransaction(); err != nil {
		if err := CleanupOverlayPaths(); err != nil {
			return err
		}

		return err
	}

	return nil
}

// CancelTransaction cancels the current transaction.
// It does so by unlocking the transactional shell and unmounting the
// overlayfs.
func CancelTransaction() error {
	if err := UnlockTransaction(); err != nil {
		return err
	}

	if err := CleanupOverlayPaths(); err != nil {
		return err
	}

	return nil
}

// ApplyTransaction applies the current transaction.
// It does so by merging the overlayfs into the future root, and then
// updating the boot.
func ApplyTransaction() error {
	PrintVerbose("step:  MountFutureRoot")
	if err := MountFutureRoot(); err != nil {
		_ = CancelTransaction()
		return err
	}

	PrintVerbose("step:  MergeOverlayFS")
	if err := MergeOverlayFS("/partFuture"); err != nil {
		_ = CancelTransaction()
		return err
	}

	PrintVerbose("step:  UpdateRootBoot")
	if err := UpdateRootBoot(true); err != nil {
		_ = UnmountFutureRoot()
		_ = CancelTransaction()

		return err
	}

	PrintVerbose("step:  UpdateFsTab")
	if err := UpdateFsTab(); err != nil {
		_ = UnmountFutureRoot()
		_ = CancelTransaction()
		return err
	}

	PrintVerbose("step:  CleanupOverlayPaths")
	if err := CleanupOverlayPaths(); err != nil {
		return err
	}

	return nil
}

// TransactionalExec runs a command in a transactional shell.
// It does so by creating a new transaction, and then chrooting into the
// overlayfs.
func transactionalExec(command string, newTransaction bool) (out string, err error) {
	if newTransaction {
		if err := NewTransaction(); err != nil {
			return "", err
		}
	}

	if out, err := ChrootOverlayFS("", false, command, false); err != nil {
		_ = CancelTransaction()
		return out, err
	}

	if err := ApplyTransaction(); err != nil {
		_ = CancelTransaction()
		return "", err
	}
	return "", nil
}

func TransactionalExec(command string) (out string, err error) {
	return transactionalExec(command, true)
}

func TransactionalExecContinue(command string) (out string, err error) {
	return transactionalExec(command, false)
}

func NewTransactionalShell() (out string, err error) {
	if out, err := TransactionalExec(""); err != nil {
		return out, err
	}

	return "", nil
}
