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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
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

	a := core.NewABRootManager()
	present, err := a.GetPresent()
	if err != nil {
		return err
	}

	future, err := a.GetFuture()
	if err != nil {
		return err
	}

	if jsonFlag {
		type status struct {
			Present string `json:"present"`
			Future  string `json:"future"`
		}

		s := status{
			Present: present.Label,
			Future:  future.Label,
		}

		b, err := json.Marshal(s)
		if err != nil {
			return err
		}

		fmt.Println(string(b))
		return nil
	}

	cmdr.Info.Printf(abroot.Trans("status.infoMsg"), present.Label, future.Label)

	return nil
}
