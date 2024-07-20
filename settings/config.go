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
	IPkgMngPre    string `json:"iPkgMngPre"`
	IPkgMngPost   string `json:"iPkgMngPost"`
	IPkgMngAdd    string `json:"iPkgMngAdd"`
	IPkgMngRm     string `json:"iPkgMngRm"`
	IPkgMngApi    string `json:"iPkgMngApi"`
	IPkgMngStatus int    `json:"iPkgMngStatus"`

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

	// Lib
	LibPathStates string `json:"libPathStates"`

	// Virtual
	FullImageName string
}

var Cnf *Config
var CnfFileUsed string

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

	CnfFileUsed = viper.ConfigFileUsed()

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
		IPkgMngPre:    viper.GetString("iPkgMngPre"),
		IPkgMngPost:   viper.GetString("iPkgMngPost"),
		IPkgMngAdd:    viper.GetString("iPkgMngAdd"),
		IPkgMngRm:     viper.GetString("iPkgMngRm"),
		IPkgMngApi:    viper.GetString("iPkgMngApi"),
		IPkgMngStatus: viper.GetInt("iPkgMngStatus"),

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

		// Lib
		LibPathStates: viper.GetString("libPathStates"),

		// Virtual
		FullImageName: "",
	}

	Cnf.FullImageName = fmt.Sprintf("%s/%s:%s", Cnf.Registry, Cnf.Name, Cnf.Tag)
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
