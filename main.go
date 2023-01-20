package main

import (
	"embed"

	"github.com/vanilla-os/abroot/cmd"
	"github.com/vanilla-os/orchid/cmdr"
)

var (
	Version = "1.3.0"
)

//go:embed locales/*.yml
var fs embed.FS
var abroot *cmdr.App

func main() {

	abroot = cmd.New(Version, fs)

	// root command
	root := cmd.NewRootCommand(Version)
	abroot.CreateRootCommand(root)

	// update-boot command
	updateBoot := cmd.NewUpdateBootCommand()
	root.AddCommand(updateBoot)

	get := cmd.NewGetCommand()
	root.AddCommand(get)

	execCmd := cmd.NewExecCommand()
	root.AddCommand(execCmd)

	shellCmd := cmd.NewShellCommand()
	root.AddCommand(shellCmd)

	kargsCmd := cmd.NewKargsCommand()
	root.AddCommand(kargsCmd)

	diffCmd := cmd.NewDiffCommand()
	root.AddCommand(diffCmd)

	// run the app
	err := abroot.Run()
	if err != nil {
		cmdr.Error.Println(err)

	}

}

/*
func main() {
	rootCmd := newABRootCommand()

	rootCmd.AddCommand(cmd.NewUpdateBootCommand())
	rootCmd.AddCommand(cmd.NewGetCommand())
	rootCmd.AddCommand(cmd.NewKargsCommand())
	rootCmd.AddCommand(cmd.NewShellCommand())
	rootCmd.AddCommand(cmd.NewExecCommand())
	rootCmd.SetHelpFunc(help)
	rootCmd.Execute()

	core.CheckABRequirements()
}
*/
