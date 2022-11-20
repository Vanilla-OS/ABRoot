package core

import (
	"fmt"
	"os"
)

var (
	lockPath = "/tmp/abroot-transactions.lock"
)

func LockTransaction() error {
	/*
	 * LockTransaction locks the transactional shell.
	 * It does so by creating a lock file.
	 */
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

func UnlockTransaction() error {
	/*
	 * UnlockTransaction unlocks the transactional shell.
	 * It does so by removing the lock file.
	 */
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

func AreTransactionsLocked() bool {
	/*
	 * IsLocked returns true if the transactional shell is locked.
	 */
	if _, err := os.Stat(lockPath); err != nil {
		return false
	}

	return true
}

func NewTransaction() error {
	/*
	 * NewTransaction starts a new transaction.
	 * It does so by creating a new overlayfs with the current root as the
	 * lower layer, and then locking the transactional shell.
	 */
	if AreTransactionsLocked() {
		fmt.Println("Transactions are locked, another one is already running or a reboot is required.")
		return nil
	}

	if err := NewOverlayFS([]string{"/"}); err != nil {
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

func CancelTransaction() error {
	/*
	 * CancelTransaction cancels the current transaction.
	 * It does so by unlocking the transactional shell and unmounting the
	 * overlayfs.
	 */
	if err := UnlockTransaction(); err != nil {
		return err
	}

	if err := CleanupOverlayPaths(); err != nil {
		return err
	}

	return nil
}

func ApplyTransaction() error {
	/*
	 * ApplyTransaction applies the current transaction.
	 * It does so by merging the overlayfs into the future root, and then
	 * updating the boot.
	 */
	if err := MountFutureRoot(); err != nil {
		_ = CancelTransaction()
		return err
	}

	if err := MergeOverlayFS("/partFuture"); err != nil {
		_ = CancelTransaction()
		return err
	}

	if err := UpdateRootBoot(true); err != nil {
		_ = UnmountFutureRoot()
		_ = CancelTransaction()
		return err
	}

	if err := CleanupOverlayPaths(); err != nil {
		return err
	}

	return nil
}

func TransactionalExec(command string) error {
	/*
	 * NewTransactionalShell starts a new transactional shell.
	 * It does so by creating a new transaction, and then chrooting into the
	 * overlayfs.
	 */
	if err := NewTransaction(); err != nil {
		return err
	}

	if err := ChrootOverlayFS("", false, command); err != nil {
		_ = CancelTransaction
		return err
	}

	return nil
}

func NewTransactionalShell() error {
	if err := TransactionalExec(""); err != nil {
		return err
	}

	return nil
}
