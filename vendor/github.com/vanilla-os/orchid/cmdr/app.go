package cmdr

import (
	"embed"
	"errors"
	"log"
	"os"
	"path"

	"github.com/fitv/go-i18n"
	"github.com/spf13/viper"
	"github.com/vanilla-os/orchid"
)

// The App struct represents the cli application
// with supporting functionality like internationalization
// and logging.
type App struct {
	Name        string
	RootCommand *Command
	Logger      *log.Logger
	logFile     *os.File
	*i18n.I18n
}

// NewApp creates a new command line application.
// It requires an embed.FS with a top level directory
// named 'locales'.
func NewApp(name string, locales embed.FS) *App {
	// for application logs
	orchid.InitLog(name+" : 	", log.LstdFlags)

	viper.SetEnvPrefix(name)
	viper.AutomaticEnv()

	i18n, err := i18n.New(locales, "locales")
	if err != nil {
		Error.Println(err)
		os.Exit(1)
	}
	i18n.SetDefaultLocale(orchid.Locale())
	a := &App{
		Name:   name,
		Logger: log.Default(),
		I18n:   i18n,
	}
	err = a.logSetup()
	if err != nil {
		log.Printf("error setting up logging: %v", err)
	}
	return a

}
func (a *App) logSetup() error {
	err := a.ensureLogDir()
	if err != nil {
		return err
	}
	logDir, err := getLogDir(a.Name)
	if err != nil {
		return err
	}
	logFile := path.Join(logDir, a.Name+".log")
	//create your file with desired read/write permissions
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	a.logFile = f

	//set output of logs to f
	log.SetOutput(a.logFile)
	return nil

}
func (a *App) CreateRootCommand(c *Command) {
	a.RootCommand = c
	c.DisableAutoGenTag = true
	manCmd := NewManCommand(a.Name)
	manC := &Command{
		Command: manCmd,
	}
	a.RootCommand.AddCommand(manC)
}

func (a *App) Run() error {
	if a.logFile != nil {
		defer a.logFile.Close()
	}
	if a.RootCommand != nil {
		return a.RootCommand.Execute()
	}
	return errors.New("no root command defined")
}

func (a *App) ensureLogDir() error {
	logPath, err := getLogDir(a.Name)
	if err != nil {
		return err
	}
	return os.MkdirAll(logPath, 0755)
}

func getLogDir(app string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(home, ".local", "share", app), nil
}
