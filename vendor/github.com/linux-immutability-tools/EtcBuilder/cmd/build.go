package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/linux-immutability-tools/EtcBuilder/core"
	"github.com/linux-immutability-tools/EtcBuilder/settings"
	"golang.org/x/exp/slices"

	"github.com/spf13/cobra"
)

func NewBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "build",
		Short:        "Build a etc overlay based on the given System and User etc",
		RunE:         buildCommand,
		SilenceUsage: true,
	}

	return cmd
}

func copyFile(source string, target string) error {

	fin, err := os.Open(source)
	if err != nil {
		return err
	}
	defer fin.Close()

	fout, err := os.Create(target)
	if err != nil {
		return err
	}
	defer fout.Close()

	_, err = io.Copy(fout, fin)

	if err != nil {
		return err
	}

	return nil
}

func clearDirectory(fileList []fs.DirEntry, root string) error {
	for _, file := range fileList {
		if file.IsDir() {
			files, err := os.ReadDir(root + "/" + file.Name())
			if err != nil {
				return err
			}

			err = clearDirectory(files, root+"/"+file.Name())
			if err != nil {
				return err
			}
		}
		err := os.Remove(root + "/" + file.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

func fileHandler(userFile string, newSysFile string, fileInfo fs.FileInfo, newFileInfo fs.FileInfo, newSys string, oldSys string, newUser string, oldUser string) error {
	if fileInfo.IsDir() || newFileInfo.IsDir() || strings.ReplaceAll(userFile, oldUser, "") != strings.ReplaceAll(newSysFile, newSys, "") {
		return nil
	}

	if slices.Contains(settings.SpecialFiles, strings.ReplaceAll(newSysFile, strings.TrimRight(newSys, "/")+"/", "")) {
		fmt.Printf("Special merging file %s\n", fileInfo.Name())
		err := core.MergeSpecialFile(userFile, oldSys+"/"+strings.ReplaceAll(userFile, oldUser, ""), newSysFile, newUser+"/"+strings.ReplaceAll(userFile, oldUser, ""))
		if err != nil {
			return err
		}
	} else if slices.Contains(settings.OverwriteFiles, fileInfo.Name()) {
		fmt.Printf("Overwriting User %[1]s with New %[1]s!\n", fileInfo.Name()) // Don't have to do anything when overwriting
	} else {
		keep, err := core.KeepUserFile(userFile, newSysFile)
		if err != nil {
			return err
		}
		if keep {
			fmt.Printf("Keeping User file %s\n", userFile)
			dirInfo, err := os.Stat(strings.TrimRight(userFile, fileInfo.Name()))
			if err != nil {
				return err
			}
			destFilePath := newUser + "/" + strings.ReplaceAll(userFile, oldUser, "")
			os.Mkdir(strings.TrimRight(destFilePath, fileInfo.Name()), dirInfo.Mode())
			copyFile(userFile, destFilePath)
		}
	}
	return nil
}

func buildCommand(_ *cobra.Command, args []string) error {
	if len(args) <= 0 {
		return fmt.Errorf("no etc directories specified")
	} else if len(args) <= 3 {
		return fmt.Errorf("not enough directories specified")
	}

	err := settings.GatherConfigFiles()
	if err != nil {
		return err
	}

	oldSys := args[0]
	newSys := args[1]
	oldUser := args[2]
	newUser := args[3]

	newUserFiles, err := os.ReadDir(newUser)
	if err != nil {
		return err
	}

	err = clearDirectory(newUserFiles, newUser)
	if err != nil {
		return err
	}

	err = filepath.Walk(oldUser, func(userPath string, userInfo os.FileInfo, e error) error {
		err := filepath.Walk(newSys, func(newPath string, newInfo os.FileInfo, err error) error {
			return fileHandler(userPath, newPath, userInfo, newInfo, newSys, oldSys, newUser, oldUser)
		})
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func ExtBuildCommand(oldSys string, newSys string, oldUser string, newUser string) error {
	err := settings.GatherConfigFiles()
	if err != nil {
		return err
	}

	destFiles, err := os.ReadDir(newUser)
	if err != nil {
		return err
	}

	err = clearDirectory(destFiles, newUser)
	if err != nil {
		return err
	}

	err = filepath.Walk(oldUser, func(userPath string, userInfo os.FileInfo, e error) error {
		err := filepath.Walk(newSys, func(newPath string, newInfo os.FileInfo, err error) error {
			return fileHandler(userPath, newPath, userInfo, newInfo, newSys, oldSys, newUser, oldUser)
		})
		return err
	})
	if err != nil {
		return err
	}
	return nil
}
