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

// NewPackageManager returns a new PackageManager struct
func NewPackageManager() *PackageManager {
	return &PackageManager{}
}

// Add adds a package to the packages.add file
func (p *PackageManager) Add(pkg string) error {
	PrintVerbose("PackageManager:Add: running...")

	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageManager:Add:error: " + err.Error())
		return err
	}

	for _, p := range pkgs {
		if p == pkg {
			PrintVerbose("PackageManager:Add: package already added")
			return nil
		}
	}

	pkgs = append(pkgs, pkg)

	PrintVerbose("PackageManager:Add: writing packages.add")
	return p.writeAddPackages(pkgs)
}

// Remove removes a package from the packages.add file
func (p *PackageManager) Remove(pkg string) error {
	PrintVerbose("PackageManager:Remove: running...")

	pkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageManager:Remove:error: " + err.Error())
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
		PrintVerbose("PackageManager:Remove:error(2): " + err.Error())
		return err
	}

	PrintVerbose("PackageManager:Remove: writing packages.remove")
	return p.writeRemovePackages(pkg)
}

// GetAddPackages returns the packages in the packages.add file
func (p *PackageManager) GetAddPackages() ([]string, error) {
	PrintVerbose("PackageManager:GetAddPackages: running...")
	return p.getPackages(PackagesAddFile)
}

// GetRemovePackages returns the packages in the packages.remove file
func (p *PackageManager) GetRemovePackages() ([]string, error) {
	PrintVerbose("PackageManager:GetRemovePackages: running...")
	return p.getPackages(PackagesRemoveFile)
}

func (p *PackageManager) getPackages(file string) ([]string, error) {
	PrintVerbose("PackageManager:getPackages: running...")

	pkgs := []string{}
	f, err := os.Open(filepath.Join(PackagesBaseDir, file))
	if err != nil {
		PrintVerbose("PackageManager:getPackages:error: " + err.Error())
		return pkgs, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		PrintVerbose("PackageManager:getPackages:error(2): " + err.Error())
		return pkgs, err
	}

	pkgs = strings.Split(string(b), "\n")

	PrintVerbose("PackageManager:getPackages: returning packages")
	return pkgs, nil
}

func (p *PackageManager) writeAddPackages(pkgs []string) error {
	PrintVerbose("PackageManager:writeAddPackages: running...")
	return p.writePackages(PackagesAddFile, pkgs)
}

func (p *PackageManager) writeRemovePackages(pkg string) error {
	PrintVerbose("PackageManager:writeRemovePackages: running...")

	pkgs, err := p.GetRemovePackages()
	if err != nil {
		PrintVerbose("PackageManager:writeRemovePackages:error: " + err.Error())
		return err
	}

	for _, p := range pkgs {
		if p == pkg {
			PrintVerbose("PackageManager:writeRemovePackages: package already added")
			return nil
		}
	}

	pkgs = append(pkgs, pkg)

	PrintVerbose("PackageManager:writeRemovePackages: writing packages.remove")
	return p.writePackages(PackagesRemoveFile, pkgs)
}

func (p *PackageManager) writePackages(file string, pkgs []string) error {
	PrintVerbose("PackageManager:writePackages: running...")

	f, err := os.Create(filepath.Join(PackagesBaseDir, file))
	if err != nil {
		PrintVerbose("PackageManager:writePackages:error: " + err.Error())
		return err
	}
	defer f.Close()

	for _, pkg := range pkgs {
		if pkg == "" {
			continue
		}

		_, err = f.WriteString(fmt.Sprintf("%s\n", pkg))
		if err != nil {
			PrintVerbose("PackageManager:writePackages:error(2): " + err.Error())
			return err
		}
	}

	PrintVerbose("PackageManager:writePackages: packages written")
	return nil
}

func (p *PackageManager) GetFinalCmd() string {
	PrintVerbose("PackageManager:GetFinalCmd: running...")

	addPkgs, err := p.GetAddPackages()
	if err != nil {
		PrintVerbose("PackageManager:GetFinalCmd:error: " + err.Error())
		return ""
	}

	removePkgs, err := p.GetRemovePackages()
	if err != nil {
		PrintVerbose("PackageManager:GetFinalCmd:error(2): " + err.Error())
		return ""
	}

	cmd := fmt.Sprintf(
		"%s %s && %s %s",
		settings.Cnf.IPkgMngAdd, strings.Join(addPkgs, " "),
		settings.Cnf.IPkgMngRm, strings.Join(removePkgs, " "),
	)
	PrintVerbose("PackageManager:GetFinalCmd: returning cmd: " + cmd)
	return cmd
}
