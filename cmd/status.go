package cmd

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
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/google/uuid"
	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/abroot/settings"
	"github.com/vanilla-os/orchid/cmdr"
)

func NewStatusCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"status",
		abroot.Trans("status.long"),
		abroot.Trans("status.short"),
		status,
	)

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"json",
			"j",
			abroot.Trans("status.jsonFlag"),
			false))

	cmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"dump",
			"d",
			abroot.Trans("status.dumpFlag"),
			false))

	cmd.Example = "abroot status"

	return cmd
}

func status(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("status.rootRequired"))
		return nil
	}

	jsonFlag, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	dumpFlag, err := cmd.Flags().GetBool("dump")
	if err != nil {
		return err
	}

	a := core.NewABRootManager()
	present, err := a.GetPresent()
	if err != nil {
		return err
	}

	future, err := a.GetFuture()
	if err != nil {
		return err
	}

	specs := core.GetPCSpecs()
	abImage, err := core.NewABImageFromRoot()
	if err != nil {
		return err
	}

	kargs, err := core.KargsRead()
	if err != nil {
		return err
	}

	pkgMng := core.NewPackageManager(false)
	pkgsAdd, err := pkgMng.GetAddPackages()
	if err != nil {
		return err
	}
	pkgsRm, err := pkgMng.GetRemovePackages()
	if err != nil {
		return err
	}
	pkgsUnstg, err := pkgMng.GetUnstagedPackagesPlain()
	if err != nil {
		return err
	}

	if jsonFlag || dumpFlag {
		type status struct {
			Present   string       `json:"present"`
			Future    string       `json:"future"`
			CnfFile   string       `json:"cnfFile"`
			CPU       string       `json:"cpu"`
			GPU       []string     `json:"gpu"`
			Memory    string       `json:"memory"`
			ABImage   core.ABImage `json:"abimage"`
			Kargs     string       `json:"kargs"`
			PkgsAdd   []string     `json:"pkgsAdd"`
			PkgsRm    []string     `json:"pkgsRm"`
			PkgsUnstg []string     `json:"pkgsUnstg"`
		}

		s := status{
			Present:   present.Label,
			Future:    future.Label,
			CnfFile:   settings.CnfFileUsed,
			CPU:       specs.CPU,
			GPU:       specs.GPU,
			Memory:    specs.Memory,
			ABImage:   *abImage,
			Kargs:     kargs,
			PkgsAdd:   pkgsAdd,
			PkgsRm:    pkgsRm,
			PkgsUnstg: pkgsUnstg,
		}

		b, err := json.Marshal(s)
		if err != nil {
			return err
		}

		if jsonFlag {
			fmt.Println(string(b))
			return nil
		}

		tarballPath := fmt.Sprintf("/tmp/abroot-status-%s.tar.gz", uuid.New().String())
		tarballFile, err := os.Create(tarballPath)
		if err != nil {
			return err
		}
		defer tarballFile.Close()

		gzipWriter := gzip.NewWriter(tarballFile)
		defer gzipWriter.Close()

		tarWriter := tar.NewWriter(gzipWriter)
		defer tarWriter.Close()

		tarHeader := &tar.Header{
			Name: "status.json",
			Mode: 0644,
			Size: int64(len(b)),
		}
		err = tarWriter.WriteHeader(tarHeader)
		if err != nil {
			return err
		}
		_, err = tarWriter.Write(b)
		if err != nil {
			return err
		}

		err = filepath.Walk("/var/log/", func(path string, info os.FileInfo, err error) error {
			if strings.Contains(path, "abroot.log") {
				relPath, err := filepath.Rel("/var/log/", path)
				if err != nil {
					return err
				}
				tarHeader := &tar.Header{
					Name: filepath.Join("logs", relPath),
					Mode: 0644,
					Size: info.Size(),
				}
				err = tarWriter.WriteHeader(tarHeader)
				if err != nil {
					return err
				}
				logFile, err := os.Open(path)
				if err != nil {
					return err
				}
				defer logFile.Close()
				_, err = io.Copy(tarWriter, logFile)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		cmdr.Info.Printf(abroot.Trans("status.dumpMsg"), tarballPath)
		return nil
	}

	formattedGPU := ""
	for _, gpu := range specs.GPU {
		formattedGPU += fmt.Sprintf("\n\t\t- %s", gpu)
	}

	unstagedAlert := ""
	if len(pkgsUnstg) > 0 {
		unstagedAlert = fmt.Sprintf(abroot.Trans("status.unstagedFoundMsg"), len(pkgsUnstg))
	}

	cmdr.Info.Printf(
		abroot.Trans("status.infoMsg"),
		present.Label, future.Label,
		settings.CnfFileUsed,
		specs.CPU, formattedGPU, specs.Memory,
		abImage.Digest, abImage.Timestamp.Format("2006-01-02 15:04:05"), abImage.Image,
		kargs,
		strings.Join(pkgsAdd, ", "), strings.Join(pkgsRm, ", "), strings.Join(pkgsUnstg, ", "),
		unstagedAlert,
	)

	return nil
}
