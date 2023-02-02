package core

import (
	"bufio"
	"fmt"
	"github.com/vanilla-os/orchid/cmdr"
	"os"
	"os/exec"
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

var (
	diffIgnore = []string{".cache", "/var", "boot/"}
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
	_, err := os.Stat(lockPath)
	if err != nil {
		return nil // transactions are already unlocked
	}

	err = os.Remove(lockPath)
	if err != nil {
		fmt.Printf("failed to remove lock file: %s", err)
		return err
	}

	return nil
}

// IsLocked returns true if the transactional shell is locked.
func AreTransactionsLocked() bool {
	_, err := os.Stat(lockPath)
	return err == nil
}

// ReadKargsFile reads kernel arguments from file. If no kargs file exists,
// it reads from kargs.default.
func ReadKargsFile() (string, error) {
	PrintVerbose("step:  GetKargs")

	content := []byte{}

	content, err := os.ReadFile(KargsPath)
	if err != nil {
		var defaultErr error

		content, defaultErr = os.ReadFile(KargsDefaultPath)
		if defaultErr != nil {
			return "", defaultErr
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
		err := UnmountFutureRoot()
		if err != nil {
			return err
		}
	}

	PrintVerbose("step:  NewOverlayFS")

	err := NewOverlayFS([]string{"/"})
	if err != nil {
		_ = UnlockTransaction()
		return err
	}

	PrintVerbose("step:  RunStartRules")

	err = RunStartRules()
	if err != nil {
		_ = UnlockTransaction()
		return err
	}

	PrintVerbose("step:  LockTransaction")

	err = LockTransaction()
	if err != nil {
		_ = CleanupOverlayPaths()
		return err
	}

	return nil
}

// CancelTransaction cancels the current transaction.
// It does so by unlocking the transactional shell and unmounting the
// overlayfs.
func CancelTransaction() error {
	err := UnlockTransaction()
	if err != nil {
		PrintVerbose("err:  CancelTransaction: %s", err)
		return err
	}

	err = CleanupOverlayPaths()
	if err != nil {
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

	err := RunEndRules()
	if err != nil {
		_ = UnmountFutureRoot()
		_ = CancelTransaction()

		return err
	}

	PrintVerbose("step:  MountFutureRoot")

	err = MountFutureRoot()
	if err != nil {
		_ = CancelTransaction()
		return err
	}

	PrintVerbose("step:  MergeOverlayFS")

	err = MergeOverlayFS("/partFuture")
	if err != nil {
		_ = CancelTransaction()
		return err
	}

	PrintVerbose("step:  UpdateRootBoot")

	kargs, err := ReadKargsFile()
	if err != nil {
		_ = CancelTransaction()
		return err
	}

	err = UpdateRootBoot(true, kargs)
	if err != nil {
		_ = UnmountFutureRoot()
		_ = CancelTransaction()

		return err
	}

	PrintVerbose("step:  UpdateFsTab")

	err = UpdateFsTab()
	if err != nil {
		_ = UnmountFutureRoot()
		_ = CancelTransaction()

		return err
	}

	PrintVerbose("step:  CleanupOverlayPaths")

	err = CleanupOverlayPaths()
	if err != nil {
		return err
	}

	return nil
}

// TransactionDiff prints a list of added, modified, and removed files
// from the lastest transaction.
func TransactionDiff() {
	PrintVerbose("step:  TransactionDiff")

	if !AreTransactionsLocked() {
		cmdr.Warning.Println("No transaction has been made since last reboot. Nothing to diff.")
		return
	}

	spinner, _ := cmdr.Spinner.Start("Gathering changes made by transaction...")
	cmd := exec.Command("rsync", "-avxHAXSn", "--delete", "--out-format=%i%n", "/.system", "/partFuture")

	// force english locale because output changes based on language
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "LANG=en_US.UTF-8")

	diff, _ := cmd.Output()
	scanner := bufio.NewScanner(strings.NewReader(string(diff)))

	var onlyPresent []string

	var onlyFuture []string

	var differ []string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		args := line[:12]
		filename := line[11:]

		skip := false

		for i := 0; i < len(diffIgnore); i++ {
			if strings.Contains(filename, diffIgnore[i]) || filename[len(filename)-1] == '/' {
				skip = true
			}
		}

		if skip {
			continue
		}

		if args[2] == '+' {
			onlyPresent = append(onlyPresent, strings.Join(strings.Split(filename, "/")[1:], "/"))
		} else if args[0] == '>' {
			differ = append(differ, strings.Join(strings.Split(filename, "/")[1:], "/"))
		} else if args[0] == '*' {
			onlyFuture = append(onlyFuture, strings.Join(strings.Split(filename, "/")[1:], "/"))
		}
	}
	spinner.Success()

	var bulletItems []cmdr.BulletListItem

	style := cmdr.NewStyle(cmdr.Bold, cmdr.FgRed)
	style.Println("Removed:")

	for i := 0; i < len(onlyPresent); i++ {
		filename := onlyPresent[i]

		bulletItems = append(bulletItems, cmdr.BulletListItem{
			Level: 1,
			Text:  "/" + filename,
		})
	}

	err := cmdr.BulletList.WithItems(bulletItems).Render()
	if err != nil {
		cmdr.Error.Println(err)
		return
	}

	fmt.Print("\n")

	bulletItems = []cmdr.BulletListItem{}

	style = cmdr.NewStyle(cmdr.Bold, cmdr.FgGreen)
	style.Println("Added:")

	for i := 0; i < len(onlyFuture); i++ {
		filename := onlyFuture[i]

		bulletItems = append(bulletItems, cmdr.BulletListItem{
			Level: 1,
			Text:  "/" + filename,
		})
	}

	err = cmdr.BulletList.WithItems(bulletItems).Render()
	if err != nil {
		cmdr.Error.Println(err)
		return
	}

	fmt.Print("\n")

	bulletItems = []cmdr.BulletListItem{}

	style = cmdr.NewStyle(cmdr.Bold, cmdr.FgYellow)
	style.Println("Modified:")

	for i := 0; i < len(differ); i++ {
		filename := differ[i]

		bulletItems = append(bulletItems, cmdr.BulletListItem{
			Level: 1,
			Text:  "/" + filename,
		})
	}

	err = cmdr.BulletList.WithItems(bulletItems).Render()
	if err != nil {
		cmdr.Error.Println(err)
		return
	}
}

// TransactionalExec runs a command in a transactional shell.
// It does so by creating a new transaction, and then chrooting into the
// overlayfs.
func transactionalExec(command string, newTransaction bool) (out string, err error) {
	if newTransaction {
		err = NewTransaction()
		if err != nil {
			return "", err
		}
	}

	out, err = ChrootOverlayFS("", false, command, false)
	if err != nil {
		_ = CancelTransaction()
		return out, err
	}

	err = ApplyTransaction()
	if err != nil {
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
	out, err = TransactionalExec("")
	if err != nil {
		return out, err
	}

	return "", nil
}

// RunStartRules runs the start transaction rules defined by the distribution
// developers in /etc/abroot/start-transaction-rules.d/.
func RunStartRules() error {
	files := getRulesFiles(startRulesPath)
	for _, file := range files {
		_, err := ChrootOverlayFS("", false, fmt.Sprintf("/bin/sh %s", file), false)
		if err != nil {
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
		_, err := ChrootOverlayFS("", false, fmt.Sprintf("/bin/sh %s", file), false)
		if err != nil {
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
