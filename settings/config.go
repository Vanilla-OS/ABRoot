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
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	// Common
	AutoRepair bool `json:"autoRepair"`

	// Registry
	Registry           string `json:"registry"`
	RegistryAPIVersion string `json:"registryAPIVersion"`
	Name               string `json:"name"`
	Tag                string `json:"tag"`

	// Package manager
	IPkgMngAdd string `json:"iPkgMngAdd"`
	IPkgMngRm  string `json:"iPkgMngRm"`

	// Partitions
	PartLabelHome string `json:"partLabelHome"`
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
	// prod paths
	viper.AddConfigPath("/etc/abroot/")
	viper.AddConfigPath("/usr/share/abroot/")

	// dev paths
	viper.AddConfigPath("config/")

	// tests paths
	viper.AddConfigPath("../config/")

	viper.SetConfigName("abroot")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatal(err)
	}

	Cnf = &Config{
		// Common
		AutoRepair: viper.GetBool("autoRepair"),

		// Registry
		Registry:           viper.GetString("registry"),
		RegistryAPIVersion: viper.GetString("registryAPIVersion"),
		Name:               viper.GetString("name"),
		Tag:                viper.GetString("tag"),

		// Package manager
		IPkgMngAdd: viper.GetString("iPkgMngAdd"),
		IPkgMngRm:  viper.GetString("iPkgMngRm"),

		// Partitions
		PartLabelHome: viper.GetString("partLabelHome"),
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
