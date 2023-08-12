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
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/vanilla-os/abroot/settings"
)

// PackageManager struct
type PackageManager struct{}

const (
	PackagesBaseDir      = "/etc/abroot"
	PackagesAddFile      = "packages.add"
	PackagesRemoveFile   = "packages.remove"
	PackagesUnstagedFile = "packages.unstaged"
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
func NewPackageManager() *PackageManager {
	PrintVerbose("PackageManager.NewPackageManager: running...")

	err := os.MkdirAll(PackagesBaseDir, 0755)
	if err != nil {
		PrintVerbose("PackageManager.NewPackageManager:err: " + err.Error())
		panic(err)
	}

	_, err = os.Stat(filepath.Join(PackagesBaseDir, PackagesAddFile))
	if err != nil {
		err = os.WriteFile(
			filepath.Join(PackagesBaseDir, PackagesAddFile),
			[]byte(""),
			0644,
		)
		if err != nil {
			PrintVerbose("PackageManager.NewPackageManager:err: " + err.Error())
			panic(err)
		}
	}

	_, err = os.Stat(filepath.Join(PackagesBaseDir, PackagesRemoveFile))
	if err != nil {
		err = os.WriteFile(
			filepath.Join(PackagesBaseDir, PackagesRemoveFile),
			[]byte(""),
			0644,
		)
		if err != nil {
			PrintVerbose("PackageManager.NewPackageManager:err: " + err.Error())
			panic(err)
		}
	}

	_, err = os.Stat(filepath.Join(PackagesBaseDir, PackagesUnstagedFile))
	if err != nil {
		err = os.WriteFile(
			filepath.Join(PackagesBaseDir, PackagesUnstagedFile),
			[]byte(""),
			0644,
		)
		if err != nil {
			PrintVerbose("PackageManager.NewPackageManager:err: " + err.Error())
			panic(err)
		}
	}

	return &PackageManager{}
}

// Add adds a package to the packages.add file
func (p *PackageManager) Add(pkg string) error {
	PrintVerbose("PackageManager.Add: running...")

	err := p.ExistsInRepo(pkg)
	if err != nil {
		PrintVerbose("PackageManager.Add:err: " + err.Error())
		return err
	}

	// Add to unstaged packages first
	upkgs, err := p.GetUnstagedPackages()
	if err != nil {
		PrintVerbose("PackageManager.Add:err: " + err.Error())
		return err
	}
	upkgs = append(upkgs, UnstagedPackage{pkg, ADD})
	err = p.writeUnstagedPackages(upkgs)
	if err != nil {
		PrintVerbose("PackageManager.Add:err(2): " + err.Error())
		return err
	}

	// Modify added packages list
	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageManager.Add:err(3): " + err.Error())
		return err
	}

	for _, p := range pkgs {
		if p == pkg {
			PrintVerbose("PackageManager.Add: package already added")
			return nil
		}
	}

	pkgs = append(pkgs, pkg)

	PrintVerbose("PackageManager.Add: writing packages.add")
	return p.writeAddPackages(pkgs)
}

// Remove removes a package from the packages.add file
func (p *PackageManager) Remove(pkg string) error {
	PrintVerbose("PackageManager.Remove: running...")

	// Add to unstaged packages first
	upkgs, err := p.GetUnstagedPackages()
	if err != nil {
		PrintVerbose("PackageManager.Add:err: " + err.Error())
		return err
	}
	upkgs = append(upkgs, UnstagedPackage{pkg, REMOVE})
	err = p.writeUnstagedPackages(upkgs)
	if err != nil {
		PrintVerbose("PackageManager.Remove:err(2): " + err.Error())
		return err
	}

	// If package was added by the user, simply remove it from packages.add
	// Unstaged will take care of the rest
	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageManager.Remove:err(3): " + err.Error())
		return err
	}
	for i, ap := range pkgs {
		if ap == pkg {
			pkgs = append(pkgs[:i], pkgs[i+1:]...)
			PrintVerbose("PackageManager.Remove: removing manually added package")
			return p.writeAddPackages(pkgs)
		}
	}

	// Otherwise, add package to packages.remove
	PrintVerbose("PackageManager.Remove: writing packages.remove")
	return p.writeRemovePackages(pkg)
}

// GetAddPackages returns the packages in the packages.add file
func (p *PackageManager) GetAddPackages() ([]string, error) {
	PrintVerbose("PackageManager.GetAddPackages: running...")
	return p.getPackages(PackagesAddFile)
}

// GetRemovePackages returns the packages in the packages.remove file
func (p *PackageManager) GetRemovePackages() ([]string, error) {
	PrintVerbose("PackageManager.GetRemovePackages: running...")
	return p.getPackages(PackagesRemoveFile)
}

// GetUnstagedPackages returns the package changes that are yet to be applied
func (p *PackageManager) GetUnstagedPackages() ([]UnstagedPackage, error) {
	PrintVerbose("PackageManager.GetUnstagedPackages: running...")
	pkgs, err := p.getPackages(PackagesUnstagedFile)
	if err != nil {
		PrintVerbose("PackageManager.GetUnstagedPackages:err: " + err.Error())
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

// ClearUnstagedPackages removes all packages from the unstaged list
func (p *PackageManager) ClearUnstagedPackages() error {
	PrintVerbose("PackageManager.ClearUnstagedPackages: running...")
	return p.writeUnstagedPackages([]UnstagedPackage{})
}

// GetAddPackages returns the packages in the packages.add file as string
func (p *PackageManager) GetAddPackagesString(sep string) (string, error) {
	PrintVerbose("PackageManager.GetAddPackagesString: running...")
	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageManager.GetAddPackagesString:err: " + err.Error())
		return "", err
	}

	PrintVerbose("PackageManager.GetAddPackagesString: done")
	return strings.Join(pkgs, sep), nil
}

// GetRemovePackages returns the packages in the packages.remove file as string
func (p *PackageManager) GetRemovePackagesString(sep string) (string, error) {
	PrintVerbose("PackageManager.GetRemovePackagesString: running...")
	pkgs, err := p.GetRemovePackages()
	if err != nil {
		PrintVerbose("PackageManager.GetRemovePackagesString:err: " + err.Error())
		return "", err
	}

	PrintVerbose("PackageManager.GetRemovePackagesString: done")
	return strings.Join(pkgs, sep), nil
}

func (p *PackageManager) getPackages(file string) ([]string, error) {
	PrintVerbose("PackageManager.getPackages: running...")

	pkgs := []string{}
	f, err := os.Open(filepath.Join(PackagesBaseDir, file))
	if err != nil {
		PrintVerbose("PackageManager.getPackages:err: " + err.Error())
		return pkgs, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		PrintVerbose("PackageManager.getPackages:err(2): " + err.Error())
		return pkgs, err
	}

	pkgs = strings.Split(string(b), "\n")

	PrintVerbose("PackageManager.getPackages: returning packages")
	return pkgs, nil
}

func (p *PackageManager) writeAddPackages(pkgs []string) error {
	PrintVerbose("PackageManager.writeAddPackages: running...")
	return p.writePackages(PackagesAddFile, pkgs)
}

func (p *PackageManager) writeRemovePackages(pkg string) error {
	PrintVerbose("PackageManager.writeRemovePackages: running...")

	pkgs, err := p.GetRemovePackages()
	if err != nil {
		PrintVerbose("PackageManager.writeRemovePackages:err: " + err.Error())
		return err
	}

	for _, p := range pkgs {
		if p == pkg {
			PrintVerbose("PackageManager.writeRemovePackages: package already added")
			return nil
		}
	}

	pkgs = append(pkgs, pkg)

	PrintVerbose("PackageManager.writeRemovePackages: writing packages.remove")
	return p.writePackages(PackagesRemoveFile, pkgs)
}

func (p *PackageManager) writeUnstagedPackages(pkgs []UnstagedPackage) error {
	PrintVerbose("PackageManager.writeUnstagedPackages: running...")

	pkgFmt := []string{}
	for _, pkg := range pkgs {
		pkgFmt = append(pkgFmt, fmt.Sprintf("%s %s", pkg.Status, pkg.Name))
	}

	return p.writePackages(PackagesUnstagedFile, pkgFmt)
}

func (p *PackageManager) writePackages(file string, pkgs []string) error {
	PrintVerbose("PackageManager.writePackages: running...")

	f, err := os.Create(filepath.Join(PackagesBaseDir, file))
	if err != nil {
		PrintVerbose("PackageManager.writePackages:err: " + err.Error())
		return err
	}
	defer f.Close()

	for _, pkg := range pkgs {
		if pkg == "" {
			continue
		}

		_, err = f.WriteString(fmt.Sprintf("%s\n", pkg))
		if err != nil {
			PrintVerbose("PackageManager.writePackages:err(2): " + err.Error())
			return err
		}
	}

	PrintVerbose("PackageManager.writePackages: packages written")
	return nil
}

func (p *PackageManager) processApplyPackages() (string, string) {
	PrintVerbose("PackageManager.processApplyPackages: running...")

	unstaged, err := p.GetUnstagedPackages()
	if err != nil {
		PrintVerbose("PackageManager.processApplyPackages:err: %s", err.Error())
	}

	var addPkgs, removePkgs []string
	for _, pkg := range unstaged {
		if pkg.Status == ADD {
			addPkgs = append(addPkgs, pkg.Name)
		} else if pkg.Status == REMOVE {
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
		PrintVerbose("PackageManager.GetFinalCmd:err: " + err.Error())
		return "", ""
	}

	removePkgs, err := p.GetRemovePackagesString(" ")
	if err != nil {
		PrintVerbose("PackageManager.GetFinalCmd:err(2): " + err.Error())
		return "", ""
	}

	if len(addPkgs) == 0 && len(removePkgs) == 0 {
		PrintVerbose("PackageManager.GetFinalCmd: no packages to install or remove")
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
	PrintVerbose("PackageManager.GetFinalCmd: running...")

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

	PrintVerbose("PackageManager.GetFinalCmd: returning cmd: " + cmd)
	return cmd
}

func (p *PackageManager) ExistsInRepo(pkg string) error {
	PrintVerbose("PackageManager.ExistsInRepo: running...")

	if settings.Cnf.IPkgMngApi == "" {
		PrintVerbose("PackageManager.ExistsInRepo: no API url set, will not check if package exists. This could lead to errors")
		return nil
	}

	if !strings.Contains(settings.Cnf.IPkgMngApi, "{packageName}") {
		return fmt.Errorf("PackageManager.ExistsInRepo: API url does not contain {packageName} placeholder. ABRoot is probably misconfigured, please report the issue to the maintainers of the distribution")
	}

	url := strings.Replace(settings.Cnf.IPkgMngApi, "{packageName}", pkg, 1)
	PrintVerbose("PackageManager.ExistsInRepo: checking if package exists in repo: " + url)

	resp, err := http.Get(url)
	if err != nil {
		PrintVerbose("PackageManager.ExistsInRepo:err: " + err.Error())
		return err
	}

	if resp.StatusCode != 200 {
		PrintVerbose("PackageManager.ExistsInRepo: package does not exist in repo")
		return fmt.Errorf("package does not exist in repo")
	}

	PrintVerbose("PackageManager.ExistsInRepo: package exists in repo")
	return nil
}
