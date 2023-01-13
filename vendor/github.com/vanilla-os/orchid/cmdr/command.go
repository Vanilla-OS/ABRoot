package cmdr

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Command represents a cli command which
// may have flags, arguments, and children commands.
type Command struct {
	*cobra.Command
	children []*Command
}

// AddCommand adds a command to the slice and to the underlying
// cobra command.
func (c *Command) AddCommand(commands ...*Command) {
	c.children = append(c.children, commands...)
	for _, cmd := range commands {
		c.Command.AddCommand(cmd.Command)
	}
}

// Children returns the children commands.
func (c *Command) Children() []*Command {
	return c.children
}

// WithBoolFlag adds a boolean flag to the command and
// registers the flag with environment variable injection
func (c *Command) WithBoolFlag(f BoolFlag) *Command {
	c.Command.Flags().BoolP(f.Name, f.Shorthand, f.Value, f.Usage)
	viper.BindPFlag(f.Name, c.Command.Flags().Lookup(f.Name))
	return c
}

// WithPersistentBoolFlag adds a persistent boolean flag to the command and
// registers the flag with environment variable injection
func (c *Command) WithPersistentBoolFlag(f BoolFlag) *Command {
	c.Command.PersistentFlags().BoolP(f.Name, f.Shorthand, f.Value, f.Usage)
	viper.BindPFlag(f.Name, c.Command.PersistentFlags().Lookup(f.Name))
	return c
}

// WithStringFlag adds a string flag to the command and registers
// the command with the environment variable injection
func (c *Command) WithStringFlag(f StringFlag) *Command {
	c.Command.Flags().StringP(f.Name, f.Shorthand, f.Value, f.Usage)
	viper.BindPFlag(f.Name, c.Command.Flags().Lookup(f.Name))
	return c
}

// WithPersistentStringFlag adds a persistent string flag to the command and registers
// the command with the environment variable injection
func (c *Command) WithPersistentStringFlag(f BoolFlag) *Command {
	c.Command.PersistentFlags().BoolP(f.Name, f.Shorthand, f.Value, f.Usage)
	viper.BindPFlag(f.Name, c.Command.PersistentFlags().Lookup(f.Name))
	return c
}

// NewCommand returns a new Command with the provided inputs
func NewCommand(use, long, short string, runE func(cmd *cobra.Command, args []string) error) *Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		RunE:  runE,
	}
	return &Command{
		Command:  cmd,
		children: make([]*Command, 0),
	}
}

// NewCustomCommand returns a Command created from
// the provided cobra.Command
func NewCommandCustom(cmd *cobra.Command) *Command {
	return &Command{
		Command:  cmd,
		children: make([]*Command, 0),
	}
}
