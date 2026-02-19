package ui

import (
	"github.com/pterm/pterm"
)

func PrintSuccess(msg string) {
	pterm.Success.Println(msg)
}

func PrintError(msg string) {
	pterm.Error.Println(msg)
}

func PrintInfo(msg string) {
	pterm.Info.Println(msg)
}
