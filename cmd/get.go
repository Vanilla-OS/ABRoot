package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
	"github.com/vanilla-os/orchid/cmdr"
)

var validArgs = []string{"present", "future"}

func NewGetCommand() *cmdr.Command {
	cmd := cmdr.NewCommand(
		"get <present|future>",
		abroot.Trans("get.long"),
		abroot.Trans("get.short"),
		get,
	)
	cmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	cmd.Example = "abroot get present\nabroot get future"
	cmd.ValidArgs = validArgs

	return cmd
}

func get(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(false) {
		cmdr.Error.Println(abroot.Trans("get.rootRequired"))
		return nil
	}

	template := "%s root partition: %s\n"

	switch args[0] {
	case "present":
		presentLabel, err := core.GetPresentRootLabel()
		if err != nil {
			cmdr.Error.Println("Error getting present root partition.")
			return err
		}

		cmdr.Info.Printf(template, "Present", presentLabel)

	case "future":
		futureLabel, err := core.GetFutureRootLabel()
		if err != nil {
			cmdr.Error.Println("Error getting future root partition.")
			return err
		}

		cmdr.Info.Printf(template, "Future", futureLabel)
	default:
		cmdr.Error.Printf("Unknown state: %s\n", args[0])
	}

	return nil
}
