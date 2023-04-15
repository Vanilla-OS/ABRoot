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
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Registry  string `json:"registry"`
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	HooksPath string `json:"hooksPath"`
}

var Cnf *Config

func init() {
	viper.AddConfigPath("/etc/abroot/")
	viper.AddConfigPath("/usr/share/abroot/")
	viper.AddConfigPath("config/")
	viper.SetConfigName("abroot")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatal(err)
	}

	Cnf = &Config{
		Registry:  viper.GetString("registry"),
		Name:      viper.GetString("name"),
		Tag:       viper.GetString("tag"),
		HooksPath: viper.GetString("hooksPath"),
	}
}
