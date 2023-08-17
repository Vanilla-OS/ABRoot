package settings

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
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	// Common
	AutoRepair           bool `json:"autoRepair"`
	MaxParallelDownloads uint `json:"maxParallelDownloads"`

	// Registry
	Registry           string `json:"registry"`
	RegistryAPIVersion string `json:"registryAPIVersion"`
	RegistryService    string `json:"registryService"`
	Name               string `json:"name"`
	Tag                string `json:"tag"`

	// Package manager
	IPkgMngPre  string `json:"iPkgMngPre"`
	IPkgMngPost string `json:"iPkgMngPost"`
	IPkgMngAdd  string `json:"iPkgMngAdd"`
	IPkgMngRm   string `json:"iPkgMngRm"`
	IPkgMngApi  string `json:"iPkgMngApi"`

	// Partitions
	PartLabelVar  string `json:"partLabelVar"`
	PartLabelA    string `json:"partLabelA"`
	PartLabelB    string `json:"partLabelB"`
	PartLabelBoot string `json:"partLabelBoot"`
	PartLabelEfi  string `json:"partLabelEfivar"`

	// Lib
	LibPathStates string `json:"libPathStates"`

	// Virtual
	FullImageName string
}

var Cnf *Config

func init() {
	// user paths
	homedir, _ := os.UserHomeDir()
	viper.AddConfigPath(homedir + "/.config/abroot/")

	// dev paths
	viper.AddConfigPath("config/")
	viper.AddConfigPath("../config/")

	// prod paths
	viper.AddConfigPath("/etc/abroot/")
	viper.AddConfigPath("/usr/share/abroot/")

	viper.SetConfigName("abroot")
	viper.SetConfigType("json")

	err := viper.ReadInConfig()
	if err != nil {
		return
	}

	Cnf = &Config{
		// Common
		AutoRepair:           viper.GetBool("autoRepair"),
		MaxParallelDownloads: viper.GetUint("maxParallelDownloads"),

		// Registry
		Registry:           viper.GetString("registry"),
		RegistryAPIVersion: viper.GetString("registryAPIVersion"),
		RegistryService:    viper.GetString("registryService"),
		Name:               viper.GetString("name"),
		Tag:                viper.GetString("tag"),

		// Package manager
		IPkgMngPre:  viper.GetString("iPkgMngPre"),
		IPkgMngPost: viper.GetString("iPkgMngPost"),
		IPkgMngAdd:  viper.GetString("iPkgMngAdd"),
		IPkgMngRm:   viper.GetString("iPkgMngRm"),
		IPkgMngApi:  viper.GetString("iPkgMngApi"),

		// Partitions
		PartLabelVar:  viper.GetString("partLabelVar"),
		PartLabelA:    viper.GetString("partLabelA"),
		PartLabelB:    viper.GetString("partLabelB"),
		PartLabelBoot: viper.GetString("partLabelBoot"),
		PartLabelEfi:  viper.GetString("partLabelEfi"),

		// Lib
		LibPathStates: viper.GetString("libPathStates"),

		// Virtual
		FullImageName: "",
	}

	Cnf.FullImageName = fmt.Sprintf("%s/%s:%s", Cnf.Registry, Cnf.Name, Cnf.Tag)
}
