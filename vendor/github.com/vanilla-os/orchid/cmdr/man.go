package cmdr

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Pietro di Caprio <pietro@fabricators.ltd>
	Copyright: 2023
	Description: Orchid is a cli helper for VanillaOS projects
*/

import (
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/vanilla-os/orchid"
)

func NewManCommand(title string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "man",
		Short:                 "Generates manpages",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Args:                  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			header := &doc.GenManHeader{
				Title:   title,
				Source:  "VanillaOS/orchid",
				Manual:  title + " Manual",
				Section: "1",
			}
			manpath := path.Join("man", orchid.Locale())
			err := os.MkdirAll(manpath, 0755)
			if err != nil {
				return err
			}
			return doc.GenManTree(cmd.Root(), header, manpath)

		},
	}
	return cmd
}
