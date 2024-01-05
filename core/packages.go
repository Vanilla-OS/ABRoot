package core

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/vanilla-os/abroot/settings"
)

// PackageManager struct
type PackageManager struct {
	dryRun  bool
	baseDir string
}

const (
	PackagesBaseDir       = "/etc/abroot"
	DryRunPackagesBaseDir = "/tmp/abroot"
	PackagesAddFile       = "packages.add"
	PackagesRemoveFile    = "packages.remove"
	PackagesUnstagedFile  = "packages.unstaged"
)

const (
	ADD    = "+"
	REMOVE = "-"
)

// An unstaged package is a package that is waiting to be applied
// to the next root.
//
// Every time a `pkg apply` or `upgrade` operation
// is executed, all unstaged packages are consumed and added/removed
// in the next root.
type UnstagedPackage struct {
	Name, Status string
}

// NewPackageManager returns a new PackageManager struct
func NewPackageManager(dryRun bool) *PackageManager {
	PrintVerboseInfo("PackageManager.NewPackageManager", "running...")

	baseDir := PackagesBaseDir
	if dryRun {
		baseDir = DryRunPackagesBaseDir
	}

	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		PrintVerboseErr("PackageManager.NewPackageManager", 0, err)
		panic(err)
	}

	_, err = os.Stat(filepath.Join(baseDir, PackagesAddFile))
	if err != nil {
		err = os.WriteFile(
			filepath.Join(baseDir, PackagesAddFile),
			[]byte(""),
			0644,
		)
		if err != nil {
			PrintVerboseErr("PackageManager.NewPackageManager", 1, err)
			panic(err)
		}
	}

	_, err = os.Stat(filepath.Join(baseDir, PackagesRemoveFile))
	if err != nil {
		err = os.WriteFile(
			filepath.Join(baseDir, PackagesRemoveFile),
			[]byte(""),
			0644,
		)
		if err != nil {
			PrintVerboseErr("PackageManager.NewPackageManager", 2, err)
			panic(err)
		}
	}

	_, err = os.Stat(filepath.Join(baseDir, PackagesUnstagedFile))
	if err != nil {
		err = os.WriteFile(
			filepath.Join(baseDir, PackagesUnstagedFile),
			[]byte(""),
			0644,
		)
		if err != nil {
			PrintVerboseErr("PackageManager.NewPackageManager", 3, err)
			panic(err)
		}
	}

	return &PackageManager{dryRun, baseDir}
}

// Add adds a package to the packages.add file
func (p *PackageManager) Add(pkg string) error {
	PrintVerboseInfo("PackageManager.Add", "running...")

	// Check if package exists in repo
	for _, _pkg := range strings.Split(pkg, " ") {
		err := p.ExistsInRepo(_pkg)
		if err != nil {
			PrintVerboseErr("PackageManager.Add", 0, err)
			return err
		}
	}

	// Add to unstaged packages first
	upkgs, err := p.GetUnstagedPackages()
	if err != nil {
		PrintVerboseErr("PackageManager.Add", 1, err)
		return err
	}
	upkgs = append(upkgs, UnstagedPackage{pkg, ADD})
	err = p.writeUnstagedPackages(upkgs)
	if err != nil {
		PrintVerboseErr("PackageManager.Add", 2, err)
		return err
	}

	// Modify added packages list
	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerboseErr("PackageManager.Add", 3, err)
		return err
	}

	for _, p := range pkgs {
		if p == pkg {
			PrintVerboseInfo("PackageManager.Add", "package already added")
			return nil
		}
	}

	pkgs = append(pkgs, pkg)

	PrintVerboseInfo("PackageManager.Add", "writing packages.add")
	return p.writeAddPackages(pkgs)
}

// Remove removes a package from the packages.add file
func (p *PackageManager) Remove(pkg string) error {
	PrintVerboseInfo("PackageManager.Remove", "running...")

	// Add to unstaged packages first
	upkgs, err := p.GetUnstagedPackages()
	if err != nil {
		PrintVerboseErr("PackageManager.Remove", 0, err)
		return err
	}
	upkgs = append(upkgs, UnstagedPackage{pkg, REMOVE})
	err = p.writeUnstagedPackages(upkgs)
	if err != nil {
		PrintVerboseErr("PackageManager.Remove", 1, err)
		return err
	}

	// If package was added by the user, simply remove it from packages.add
	// Unstaged will take care of the rest
	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerboseErr("PackageManager.Remove", 2, err)
		return err
	}
	for i, ap := range pkgs {
		if ap == pkg {
			pkgs = append(pkgs[:i], pkgs[i+1:]...)
			PrintVerboseInfo("PackageManager.Remove", "removing manually added package")
			return p.writeAddPackages(pkgs)
		}
	}

	// Otherwise, add package to packages.remove
	PrintVerboseInfo("PackageManager.Remove", "writing packages.remove")
	return p.writeRemovePackages(pkg)
}

// GetAddPackages returns the packages in the packages.add file
func (p *PackageManager) GetAddPackages() ([]string, error) {
	PrintVerboseInfo("PackageManager.GetAddPackages", "running...")
	return p.getPackages(PackagesAddFile)
}

// GetRemovePackages returns the packages in the packages.remove file
func (p *PackageManager) GetRemovePackages() ([]string, error) {
	PrintVerboseInfo("PackageManager.GetRemovePackages", "running...")
	return p.getPackages(PackagesRemoveFile)
}

// GetUnstagedPackages returns the package changes that are yet to be applied
func (p *PackageManager) GetUnstagedPackages() ([]UnstagedPackage, error) {
	PrintVerboseInfo("PackageManager.GetUnstagedPackages", "running...")
	pkgs, err := p.getPackages(PackagesUnstagedFile)
	if err != nil {
		PrintVerboseErr("PackageManager.GetUnstagedPackages", 0, err)
		return nil, err
	}

	unstagedList := []UnstagedPackage{}
	for _, line := range pkgs {
		if line == "" || line == "\n" {
			continue
		}

		splits := strings.SplitN(line, " ", 2)
		unstagedList = append(unstagedList, UnstagedPackage{splits[1], splits[0]})
	}

	return unstagedList, nil
}

// GetUnstagedPackagesPlain returns the package changes that are yet to be applied
// as strings
func (p *PackageManager) GetUnstagedPackagesPlain() ([]string, error) {
	PrintVerboseInfo("PackageManager.GetUnstagedPackagesPlain", "running...")
	pkgs, err := p.GetUnstagedPackages()
	if err != nil {
		PrintVerboseErr("PackageManager.GetUnstagedPackagesPlain", 0, err)
		return nil, err
	}

	unstagedList := []string{}
	for _, pkg := range pkgs {
		unstagedList = append(unstagedList, pkg.Name)
	}

	return unstagedList, nil
}

// ClearUnstagedPackages removes all packages from the unstaged list
func (p *PackageManager) ClearUnstagedPackages() error {
	PrintVerboseInfo("PackageManager.ClearUnstagedPackages", "running...")
	return p.writeUnstagedPackages([]UnstagedPackage{})
}

// GetAddPackagesString returns the packages in the packages.add file as a string
func (p *PackageManager) GetAddPackagesString(sep string) (string, error) {
	PrintVerboseInfo("PackageManager.GetAddPackagesString", "running...")
	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerboseErr("PackageManager.GetAddPackagesString", 0, err)
		return "", err
	}

	PrintVerboseInfo("PackageManager.GetAddPackagesString", "done")
	return strings.Join(pkgs, sep), nil
}

// GetRemovePackagesString returns the packages in the packages.remove file as a string
func (p *PackageManager) GetRemovePackagesString(sep string) (string, error) {
	PrintVerboseInfo("PackageManager.GetRemovePackagesString", "running...")
	pkgs, err := p.GetRemovePackages()
	if err != nil {
		PrintVerboseErr("PackageManager.GetRemovePackagesString", 0, err)
		return "", err
	}

	PrintVerboseInfo("PackageManager.GetRemovePackagesString", "done")
	return strings.Join(pkgs, sep), nil
}

func (p *PackageManager) getPackages(file string) ([]string, error) {
	PrintVerboseInfo("PackageManager.getPackages", "running...")

	pkgs := []string{}
	f, err := os.Open(filepath.Join(p.baseDir, file))
	if err != nil {
		PrintVerboseErr("PackageManager.getPackages", 0, err)
		return pkgs, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		PrintVerboseErr("PackageManager.getPackages", 1, err)
		return pkgs, err
	}

	pkgs = strings.Split(strings.TrimSpace(string(b)), "\n")

	PrintVerboseInfo("PackageManager.getPackages", "returning packages")
	return pkgs, nil
}

func (p *PackageManager) writeAddPackages(pkgs []string) error {
	PrintVerboseInfo("PackageManager.writeAddPackages", "running...")
	return p.writePackages(PackagesAddFile, pkgs)
}

func (p *PackageManager) writeRemovePackages(pkg string) error {
	PrintVerboseInfo("PackageManager.writeRemovePackages", "running...")

	pkgs, err := p.GetRemovePackages()
	if err != nil {
		PrintVerboseErr("PackageManager.writeRemovePackages", 0, err)
		return err
	}

	for _, p := range pkgs {
		if p == pkg {
			PrintVerboseInfo("PackageManager.writeRemovePackages", "package already added")
			return nil
		}
	}

	pkgs = append(pkgs, pkg)

	PrintVerboseInfo("PackageManager.writeRemovePackages", "writing packages.remove")
	return p.writePackages(PackagesRemoveFile, pkgs)
}

func (p *PackageManager) writeUnstagedPackages(pkgs []UnstagedPackage) error {
	PrintVerboseInfo("PackageManager.writeUnstagedPackages", "running...")

	pkgFmt := []string{}
	for _, pkg := range pkgs {
		pkgFmt = append(pkgFmt, fmt.Sprintf("%s %s", pkg.Status, pkg.Name))
	}

	return p.writePackages(PackagesUnstagedFile, pkgFmt)
}

func (p *PackageManager) writePackages(file string, pkgs []string) error {
	PrintVerboseInfo("PackageManager.writePackages", "running...")

	f, err := os.Create(filepath.Join(p.baseDir, file))
	if err != nil {
		PrintVerboseErr("PackageManager.writePackages", 0, err)
		return err
	}
	defer f.Close()

	for _, pkg := range pkgs {
		if pkg == "" {
			continue
		}

		_, err = f.WriteString(fmt.Sprintf("%s\n", pkg))
		if err != nil {
			PrintVerboseErr("PackageManager.writePackages", 1, err)
			return err
		}
	}

	PrintVerboseInfo("PackageManager.writePackages", "packages written")
	return nil
}

func (p *PackageManager) processApplyPackages() (string, string) {
	PrintVerboseInfo("PackageManager.processApplyPackages", "running...")

	unstaged, err := p.GetUnstagedPackages()
	if err != nil {
		PrintVerboseErr("PackageManager.processApplyPackages", 0, err)
	}

	var addPkgs, removePkgs []string
	for _, pkg := range unstaged {
		switch pkg.Status {
		case ADD:
			addPkgs = append(addPkgs, pkg.Name)
		case REMOVE:
			removePkgs = append(removePkgs, pkg.Name)
		}
	}

	finalAddPkgs := ""
	if len(addPkgs) > 0 {
		finalAddPkgs = fmt.Sprintf("%s %s", settings.Cnf.IPkgMngAdd, strings.Join(addPkgs, " "))
	}

	finalRemovePkgs := ""
	if len(removePkgs) > 0 {
		finalRemovePkgs = fmt.Sprintf("%s %s", settings.Cnf.IPkgMngRm, strings.Join(removePkgs, " "))
	}

	return finalAddPkgs, finalRemovePkgs
}

func (p *PackageManager) processUpgradePackages() (string, string) {
	addPkgs, err := p.GetAddPackagesString(" ")
	if err != nil {
		PrintVerboseErr("PackageManager.processUpgradePackages", 0, err)
		return "", ""
	}

	removePkgs, err := p.GetRemovePackagesString(" ")
	if err != nil {
		PrintVerboseErr("PackageManager.processUpgradePackages", 1, err)
		return "", ""
	}

	if len(addPkgs) == 0 && len(removePkgs) == 0 {
		PrintVerboseInfo("PackageManager.processUpgradePackages", "no packages to install or remove")
		return "", ""
	}

	finalAddPkgs := ""
	if addPkgs != "" {
		finalAddPkgs = fmt.Sprintf("%s %s", settings.Cnf.IPkgMngAdd, addPkgs)
	}

	finalRemovePkgs := ""
	if removePkgs != "" {
		finalRemovePkgs = fmt.Sprintf("%s %s", settings.Cnf.IPkgMngRm, removePkgs)
	}

	return finalAddPkgs, finalRemovePkgs
}

func (p *PackageManager) GetFinalCmd(operation ABSystemOperation) string {
	PrintVerboseInfo("PackageManager.GetFinalCmd", "running...")

	var finalAddPkgs, finalRemovePkgs string
	if operation == APPLY {
		finalAddPkgs, finalRemovePkgs = p.processApplyPackages()
	} else {
		finalAddPkgs, finalRemovePkgs = p.processUpgradePackages()
	}

	cmd := ""
	if finalAddPkgs != "" && finalRemovePkgs != "" {
		cmd = fmt.Sprintf("%s && %s", finalAddPkgs, finalRemovePkgs)
	} else if finalAddPkgs != "" {
		cmd = finalAddPkgs
	} else if finalRemovePkgs != "" {
		cmd = finalRemovePkgs
	}

	// No need to add pre/post hooks to an empty operation
	if cmd == "" {
		return cmd
	}

	preExec := settings.Cnf.IPkgMngPre
	postExec := settings.Cnf.IPkgMngPost
	if preExec != "" {
		cmd = fmt.Sprintf("%s && %s", preExec, cmd)
	}
	if postExec != "" {
		cmd = fmt.Sprintf("%s && %s", cmd, postExec)
	}

	PrintVerboseInfo("PackageManager.GetFinalCmd", "returning cmd: "+cmd)
	return cmd
}

// assertPkgMngApiSetUp checks whether the repo API is properly configured.
// If a configuration exists but is malformed, returns an error.
func assertPkgMngApiSetUp() (bool, error) {
	if settings.Cnf.IPkgMngApi == "" {
		PrintVerboseInfo("PackageManager.assertPkgMngApiSetUp", "no API url set, will not check if package exists. This could lead to errors")
		return false, nil
	}

	_, err := url.ParseRequestURI(settings.Cnf.IPkgMngApi)
	if err != nil {
		return false, fmt.Errorf("PackageManager.assertPkgMngApiSetUp: Value set as API url (%s) is not a valid URL", settings.Cnf.IPkgMngApi)
	}

	if !strings.Contains(settings.Cnf.IPkgMngApi, "{packageName}") {
		return false, fmt.Errorf("PackageManager.assertPkgMngApiSetUp: API url does not contain {packageName} placeholder. ABRoot is probably misconfigured, please report the issue to the maintainers of the distribution")
	}

	PrintVerboseInfo("PackageManager.assertPkgMngApiSetUp", "Repo is set up properly")
	return true, nil
}

func (p *PackageManager) ExistsInRepo(pkg string) error {
	PrintVerboseInfo("PackageManager.ExistsInRepo", "running...")

	ok, err := assertPkgMngApiSetUp()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	url := strings.Replace(settings.Cnf.IPkgMngApi, "{packageName}", pkg, 1)
	PrintVerboseInfo("PackageManager.ExistsInRepo", "checking if package exists in repo: "+url)

	resp, err := http.Get(url)
	if err != nil {
		PrintVerboseErr("PackageManager.ExistsInRepo", 0, err)
		return err
	}

	if resp.StatusCode != 200 {
		PrintVerboseInfo("PackageManager.ExistsInRepo", "package does not exist in repo")
		return fmt.Errorf("package does not exist in repo: %s", pkg)
	}

	PrintVerboseInfo("PackageManager.ExistsInRepo", "package exists in repo")
	return nil
}

// GetRepoContentsForPkg retrieves package information from the repository API
func GetRepoContentsForPkg(pkg string) (map[string]interface{}, error) {
	PrintVerboseInfo("PackageManager.GetRepoContentsForPkg", "running...")

	ok, err := assertPkgMngApiSetUp()
	if err != nil {
		return map[string]interface{}{}, err
	}
	if !ok {
		return map[string]interface{}{}, errors.New("PackageManager.GetRepoContentsForPkg: no API url set, cannot query package information")
	}

	url := strings.Replace(settings.Cnf.IPkgMngApi, "{packageName}", pkg, 1)
	PrintVerboseInfo("PackageManager.GetRepoContentsForPkg", "fetching package information in: "+url)

	resp, err := http.Get(url)
	if err != nil {
		PrintVerboseErr("PackageManager.GetRepoContentsForPkg", 0, err)
		return map[string]interface{}{}, err
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		PrintVerboseErr("PackageManager.GetRepoContentsForPkg", 1, err)
		return map[string]interface{}{}, err
	}

	pkgInfo := map[string]interface{}{}
	err = json.Unmarshal(contents, &pkgInfo)
	if err != nil {
		PrintVerboseErr("PackageManager.GetRepoContentsForPkg", 2, err)
		return map[string]interface{}{}, err
	}

	return pkgInfo, nil
}
