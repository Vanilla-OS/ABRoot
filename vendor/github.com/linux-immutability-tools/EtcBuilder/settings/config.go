package settings

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

var OverwriteFiles []string
var SpecialFiles []string

type Config struct {
	OverwriteFiles []string
	SpecialFiles   []string
}

func GatherConfigFiles() error {
	configFiles, err := os.ReadDir("/usr/share/etcbuilder/")
	if err != nil {
		return err
	}

	for _, config := range configFiles {
		if !strings.Contains(config.Name(), ".toml") {
			continue
		}
		parseConfigFile("/usr/share/etcbuilder/" + config.Name())
	}

	return nil
}

func parseConfigFile(file string) {
	var conf Config
	configData, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	_, err = toml.Decode(string(configData), &conf)
	if err != nil {
		_ = fmt.Errorf("ERROR: Failed to parse configuration file %s", file)
		fmt.Printf("err: %v\n", err)
	}
	OverwriteFiles = append(OverwriteFiles, conf.OverwriteFiles...)
	SpecialFiles = append(SpecialFiles, conf.SpecialFiles...)
}
