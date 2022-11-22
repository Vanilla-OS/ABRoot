package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vanilla-os/abroot/core"
)

func getUsage(*cobra.Command) error {
	fmt.Print(`Description:
	Get the present or future root partition.

Usage:
	get [state]

Options:
	--help/-h		show this message

States:
	present			get the present root partition
	future			get the future root partition

Examples:
	abroot get present
	abroot get future
`)

	return nil
}

func NewGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get the present or future root partition",
		RunE:  get,
	}
	cmd.SetUsageFunc(getUsage)

	return cmd
}

func get(cmd *cobra.Command, args []string) error {
	if !core.RootCheck(true) {
		return nil
	}

	template := "%s root partition: %s\n"

	if len(args) == 0 {
		fmt.Println("Please specify a state (present or future)")
	}

	switch args[0] {
	case "present":
		presentLabel, err := core.GetPresentRootLabel()
		if err != nil {
			fmt.Println("Error getting present root partition.")
			return err
		}
		fmt.Printf(template, "Present", presentLabel)
	case "future":
		futureLabel, err := core.GetFutureRootLabel()
		if err != nil {
			fmt.Println("Error getting future root partition.")
			return err
		}
		fmt.Printf(template, "Future", futureLabel)
	default:
		fmt.Printf("Unknown state: %s\n", args[0])
	}

	return nil
}
