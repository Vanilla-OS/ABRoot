package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	lockPath         = "/tmp/abroot-transactions.lock"
	startRulesPath   = "/etc/abroot/start-transaction-rules.d/"
	KargsPath        = "/etc/abroot/kargs"
	KargsDefaultPath = "/etc/default/abroot_kargs"
	endRulesPath     = "/etc/abroot/end-transaction-rules.d/"
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

// ReadKargsFile reads kernel arguments from file. If no kargs file exists,
// it reads from kargs.default.
func ReadKargsFile() (string, error) {
	PrintVerbose("step:  GetKargs")
	content := []byte{}
	content, err := os.ReadFile(KargsPath)
	if err != nil {
		var default_err error
		content, default_err = os.ReadFile(KargsDefaultPath)
		if default_err != nil {
			return "", default_err
		}
	}

	// Prevent accidental newline from breaking arguments
	if content[len(content)-1] == 10 {
		return string(content[:len(content)-1]), nil
	}

	return string(content), nil
}

// NewTransaction starts a new transaction.
// It does so by creating a new overlayfs with the current root as the
// lower layer, and then locking the transactional shell.
func NewTransaction() error {
	PrintVerbose("step:  AreTransactionsLocked")
	if AreTransactionsLocked() {
		fmt.Println("Transactions are locked, another one is already running or a reboot is required.")
		os.Exit(0)
	}

	if IsMounted("/partFuture") {
		if err := UnmountFutureRoot(); err != nil {
			return err
		}
	}

	PrintVerbose("step:  NewOverlayFS")
	if err := NewOverlayFS([]string{"/"}); err != nil {
		UnlockTransaction()
		return err
	}

	PrintVerbose("step:  RunStartRules")
	if err := RunStartRules(); err != nil {
		UnlockTransaction()
		return err
	}

	PrintVerbose("step:  LockTransaction")
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
		PrintVerbose("err:  CancelTransaction: %s", err)
		return err
	}

	if err := CleanupOverlayPaths(); err != nil {
		PrintVerbose("err:  CancelTransaction: %s", err)
		return err
	}

	return nil
}

// ApplyTransaction applies the current transaction.
// It does so by merging the overlayfs into the future root, and then
// updating the boot.
func ApplyTransaction() error {
	PrintVerbose("step:  RunEndRules")
	if err := RunEndRules(); err != nil {
		_ = UnmountFutureRoot()
		_ = CancelTransaction()
		return err
	}

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
	kargs, err := ReadKargsFile()
	if err != nil {
		_ = CancelTransaction()
		return err
	}
	if err := UpdateRootBoot(true, kargs); err != nil {
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

// RunStartRules runs the start transaction rules defined by the distribution
// developers in /etc/abroot/start-transaction-rules.d/.
func RunStartRules() error {
	files := getRulesFiles(startRulesPath)
	for _, file := range files {
		if _, err := ChrootOverlayFS("", false, fmt.Sprintf("/bin/sh %s", file), false); err != nil {
			return err
		}
	}

	return nil
}

// RunEndRules runs the end transaction rules defined by the distribution
// developers in /etc/abroot/end-transaction-rules.d/.
func RunEndRules() error {
	files := getRulesFiles(endRulesPath)
	for _, file := range files {
		if _, err := ChrootOverlayFS("", false, fmt.Sprintf("/bin/sh %s", file), false); err != nil {
			return err
		}
	}

	return nil
}

// getRulesFiles returns a list of files in the given directory.
func getRulesFiles(path string) []string {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	var rules []string
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".rules") {
			continue
		}

		rules = append(rules, filepath.Join(path, file.Name()))
	}

	return rules
}
