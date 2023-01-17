package cmdr

import "github.com/pterm/pterm"

var (
	Info, Warning, Success, Fatal, Debug, Description, Error pterm.PrefixPrinter
	Spinner                                                  pterm.SpinnerPrinter
	ProgressBar                                              pterm.ProgressbarPrinter
)

func init() {
	Error = pterm.Error
	Info = pterm.Info
	Warning = pterm.Warning
	Success = pterm.Success
	Fatal = pterm.Fatal
	Debug = pterm.Debug
	Description = pterm.Description
	Spinner = pterm.DefaultSpinner
	ProgressBar = pterm.DefaultProgressbar
}
