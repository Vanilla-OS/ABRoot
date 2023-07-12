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

func fileHandler(f string, n string, fileInfo fs.FileInfo, newFileInfo fs.FileInfo, args []string) error {
	if fileInfo.IsDir() || newFileInfo.IsDir() || fileInfo.Name() != newFileInfo.Name() {
		return nil
	}

	if slices.Contains(settings.SpecialFiles, fileInfo.Name()) {
		fmt.Printf("Special merging file %s\n", fileInfo.Name())
		err := core.MergeSpecialFile(f, args[0]+"/"+strings.ReplaceAll(f, args[0], ""), n, args[3]+"/"+strings.ReplaceAll(f, args[0], ""))
		if err != nil {
			return err
		}
	} else if slices.Contains(settings.OverwriteFiles, fileInfo.Name()) {
		fmt.Printf("Overwriting User %[1]s with New %[1]s!\n", fileInfo.Name()) // Don't have to do anything when overwriting
	} else {
		keep, err := core.KeepUserFile(f, n)
		if err != nil {
			return err
		}
		if keep {
			fmt.Printf("Keeping User file %s\n", f)
			dirInfo, err := os.Stat(strings.TrimRight(f, fileInfo.Name()))
			if err != nil {
				return err
			}
			destFilePath := args[3] + "/" + strings.ReplaceAll(f, args[0], "")
			os.Mkdir(strings.TrimRight(destFilePath, fileInfo.Name()), dirInfo.Mode())
			copyFile(f, destFilePath)
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

	destFiles, err := os.ReadDir(args[2])
	if err != nil {
		return err
	}

	err = clearDirectory(destFiles, args[2])
	if err != nil {
		return err
	}

	err = filepath.Walk(args[0], func(userPath string, userInfo os.FileInfo, e error) error {
		err := filepath.Walk(args[1], func(newPath string, newInfo os.FileInfo, err error) error {
			return fileHandler(userPath, newPath, userInfo, newInfo, args)
		})
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func ExtBuildCommand(currentEtc string, newSystem string, newUser string) error {
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

	args := []string{currentEtc, newSystem, newUser}
	err = filepath.Walk(currentEtc, func(userPath string, userInfo os.FileInfo, e error) error {
		err := filepath.Walk(newSystem, func(newPath string, newInfo os.FileInfo, err error) error {
			return fileHandler(userPath, newPath, userInfo, newInfo, args)
		})
		return err
	})
	if err != nil {
		return err
	}
	return nil
}
