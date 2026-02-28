package ui

import (
	"fmt"

	"github.com/pterm/pterm"
)

const banner = `
  ____  ______     ___    ____  ____  
 / ___|/ _ \ \   / / \  |  _ \|  _ \ 
| |  _| | | \ \ / / _ \ | |_) | | | |
| |_| | |_| |\ V / ___ \|  _ <| |_| |
 \____|\___/  \_/_/   \_\_| \_\____/ 
`

func PrintBrand(version string) {
	header := pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgWhite))
	header.Println("Go-based Versatile Runtime & Development")

	fmt.Println(pterm.Blue(banner))
	pterm.Info.Printf("Govard Version: v%s\n", version)
	fmt.Println("========================================")
}
