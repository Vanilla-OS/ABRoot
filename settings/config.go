package settings

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2024
	Description:
		ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	// Common
	MaxParallelDownloads uint `json:"maxParallelDownloads"`

	// Registry
	Registry           string `json:"registry"`
	RegistryAPIVersion string `json:"registryAPIVersion"`
	RegistryService    string `json:"registryService"`
	Name               string `json:"name"`
	Tag                string `json:"tag"`

	// Package manager
	IPkgMngPre    string `json:"iPkgMngPre"`
	IPkgMngPost   string `json:"iPkgMngPost"`
	IPkgMngAdd    string `json:"iPkgMngAdd"`
	IPkgMngRm     string `json:"iPkgMngRm"`
	IPkgMngApi    string `json:"iPkgMngApi"`
	IPkgMngStatus int    `json:"iPkgMngStatus"`

	// Boot configuration commands
	UpdateInitramfsCmd string `json:"updateInitramfsCmd"`
	UpdateGrubCmd      string `json:"updateGrubCmd"`

	// Package diff API (Differ)
	DifferURL string `json:"differURL"`

	// Partitions
	PartLabelVar  string `json:"partLabelVar"`
	PartLabelA    string `json:"partLabelA"`
	PartLabelB    string `json:"partLabelB"`
	PartLabelBoot string `json:"partLabelBoot"`
	PartLabelEfi  string `json:"partLabelEfivar"`
	PartCryptVar  string `json:"PartCryptVar"`

	// Structure
	ThinProvisioning bool   `json:"thinProvisioning"`
	ThinInitVolume   string `json:"thinInitVolume"`
}

var Cnf *Config

const CnfFileName = "abroot.json"

var (
	CnfPathSystem = filepath.Join("/usr/share/abroot/", CnfFileName)
	CnfPathAdmin  = filepath.Join("/etc/abroot/", CnfFileName)
	CnfPathDev1   = filepath.Join("../config/", CnfFileName)
	CnfPathDev2   = filepath.Join("./config/", CnfFileName)
)

func init() {

	// system path
	viper.SetConfigFile(CnfPathSystem)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not read config file in /usr", err)
	}

	// admin path
	viper.SetConfigFile(CnfPathAdmin)
	err = viper.MergeInConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not read config file in /usr", err)
	}

	// dev paths
	viper.SetConfigFile(CnfPathDev1)
	viper.MergeInConfig()
	viper.SetConfigFile(CnfPathDev2)
	viper.MergeInConfig()

	// VanillaOS specific defaults for backwards compatibility
	viper.SetDefault("updateInitramfsCmd", "lpkg --unlock && /usr/sbin/update-initramfs -u && lpkg --lock")
	viper.SetDefault("updateGrubCmd", "/usr/sbin/grub-mkconfig -o '%s'")

	Cnf = &Config{
		// Common
		MaxParallelDownloads: viper.GetUint("maxParallelDownloads"),

		// Registry
		Registry:           viper.GetString("registry"),
		RegistryAPIVersion: viper.GetString("registryAPIVersion"),
		RegistryService:    viper.GetString("registryService"),
		Name:               viper.GetString("name"),
		Tag:                viper.GetString("tag"),

		// Package manager
		IPkgMngPre:    viper.GetString("iPkgMngPre"),
		IPkgMngPost:   viper.GetString("iPkgMngPost"),
		IPkgMngAdd:    viper.GetString("iPkgMngAdd"),
		IPkgMngRm:     viper.GetString("iPkgMngRm"),
		IPkgMngApi:    viper.GetString("iPkgMngApi"),
		IPkgMngStatus: viper.GetInt("iPkgMngStatus"),

		// Boot configuration commands
		UpdateInitramfsCmd: viper.GetString("updateInitramfsCmd"),
		UpdateGrubCmd:      viper.GetString("updateGrubCmd"),

		// Package diff API (Differ)
		DifferURL: viper.GetString("differURL"),

		// Partitions
		PartLabelVar:  viper.GetString("partLabelVar"),
		PartLabelA:    viper.GetString("partLabelA"),
		PartLabelB:    viper.GetString("partLabelB"),
		PartLabelBoot: viper.GetString("partLabelBoot"),
		PartLabelEfi:  viper.GetString("partLabelEfi"),
		PartCryptVar:  viper.GetString("PartCryptVar"),

		// Structure
		ThinProvisioning: viper.GetBool("thinProvisioning"),
		ThinInitVolume:   viper.GetString("thinInitVolume"),
	}
}

func GetFullImageName() string {
	return fmt.Sprintf("%s/%s", Cnf.Registry, Cnf.Name)
}

func GetFullImageNameWithTag() string {
	return fmt.Sprintf("%s:%s", GetFullImageName(), Cnf.Tag)
}

// WriteConfigToFile writes the current configuration to a file
func WriteConfigToFile(file string) error {
	jsonOutput, err := json.MarshalIndent(Cnf, "", "    ")
	if err != nil {
		return err
	}

	outputFile, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	_, err = outputFile.Write(jsonOutput)

	return err
}
