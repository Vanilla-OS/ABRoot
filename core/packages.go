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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/vanilla-os/abroot/settings"
)

// PackageManager struct
type PackageManager struct{}

const (
	PackagesBaseDir    = "/etc/abroot"
	PackagesAddFile    = "packages.add"
	PackagesRemoveFile = "packages.remove"
)

// init creates the base files and directories
func init() {
	PrintVerbose("PackageManager.init: running...")

	err := os.MkdirAll(PackagesAddFile, 0755)
	if err != nil {
		PrintVerbose("PackageManager.init:error: " + err.Error())
		panic(err)
	}

	err = os.MkdirAll(PackagesRemoveFile, 0755)
	if err != nil {
		PrintVerbose("PackageManager.init:error: " + err.Error())
		panic(err)
	}

	_, err = os.Stat(filepath.Join(PackagesBaseDir, PackagesAddFile))
	if err != nil {
		PrintVerbose("PackageManager.init:error: " + err.Error())
		err = ioutil.WriteFile(
			filepath.Join(PackagesBaseDir, PackagesAddFile),
			[]byte(""),
			0644,
		)
		if err != nil {
			PrintVerbose("PackageManager.init:error: " + err.Error())
			panic(err)
		}
	}

	_, err = os.Stat(filepath.Join(PackagesBaseDir, PackagesRemoveFile))
	if err != nil {
		PrintVerbose("PackageManager.init:error: " + err.Error())
		err = ioutil.WriteFile(
			filepath.Join(PackagesBaseDir, PackagesRemoveFile),
			[]byte(""),
			0644,
		)
		if err != nil {
			PrintVerbose("PackageManager.init:error: " + err.Error())
			panic(err)
		}
	}

	PrintVerbose("PackageManager.init: done")
}

// NewPackageManager returns a new PackageManager struct
func NewPackageManager() *PackageManager {
	PrintVerbose("PackageManager.NewPackageManager: running...")
	return &PackageManager{}
}

// Add adds a package to the packages.add file
func (p *PackageManager) Add(pkg string) error {
	PrintVerbose("PackageManager.Add: running...")

	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageManager.Add:error: " + err.Error())
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

	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageManager.Remove:error: " + err.Error())
		return err
	}

	for i, p := range pkgs {
		if p == pkg {
			pkgs = append(pkgs[:i], pkgs[i+1:]...)
			break
		}
	}

	err = p.writeAddPackages(pkgs)
	if err != nil {
		PrintVerbose("PackageManager.Remove:error(2): " + err.Error())
		return err
	}

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

// GetAddPackages returns the packages in the packages.add file as string
func (p *PackageManager) GetAddPackagesString(sep string) (string, error) {
	PrintVerbose("PackageManager.GetAddPackagesString: running...")
	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageManager.GetAddPackagesString:error: " + err.Error())
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
		PrintVerbose("PackageManager.GetRemovePackagesString:error: " + err.Error())
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
		PrintVerbose("PackageManager.getPackages:error: " + err.Error())
		return pkgs, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		PrintVerbose("PackageManager.getPackages:error(2): " + err.Error())
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
		PrintVerbose("PackageManager.writeRemovePackages:error: " + err.Error())
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

func (p *PackageManager) writePackages(file string, pkgs []string) error {
	PrintVerbose("PackageManager.writePackages: running...")

	f, err := os.Create(filepath.Join(PackagesBaseDir, file))
	if err != nil {
		PrintVerbose("PackageManager.writePackages:error: " + err.Error())
		return err
	}
	defer f.Close()

	for _, pkg := range pkgs {
		if pkg == "" {
			continue
		}

		_, err = f.WriteString(fmt.Sprintf("%s\n", pkg))
		if err != nil {
			PrintVerbose("PackageManager.writePackages:error(2): " + err.Error())
			return err
		}
	}

	PrintVerbose("PackageManager.writePackages: packages written")
	return nil
}

func (p *PackageManager) GetFinalCmd() string {
	PrintVerbose("PackageManager.GetFinalCmd: running...")

	addPkgs, err := p.GetAddPackagesString(" ")
	if err != nil {
		PrintVerbose("PackageManager.GetFinalCmd:error: " + err.Error())
		return ""
	}

	removePkgs, err := p.GetRemovePackagesString(" ")
	if err != nil {
		PrintVerbose("PackageManager.GetFinalCmd:error(2): " + err.Error())
		return ""
	}

	cmd := fmt.Sprintf(
		"%s %s && %s %s",
		settings.Cnf.IPkgMngAdd, addPkgs,
		settings.Cnf.IPkgMngRm, removePkgs,
	)
	PrintVerbose("PackageManager.GetFinalCmd: returning cmd: " + cmd)
	return cmd
}
