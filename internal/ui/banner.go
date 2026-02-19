package ui

import (
	"fmt"
	"github.com/pterm/pterm"
)

const banner = `
   ______                         __
  / ____/___  _   ______ ________/ /
 / / __/ __ \| | / / __  / ___/ _  /
/ /_/ / /_/ /| |/ / /_/ / /  / /_/ /
\____/\____/ |___/\__,_/_/   \__,_/
`

func PrintBrand(version string) {
	fmt.Println(pterm.LightCyan(banner))
	fmt.Printf("                        v%s\n\n", version)
}
