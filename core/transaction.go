package core

import (
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
