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
	logoStyle := pterm.NewStyle(pterm.FgLightBlue)
	taglineStyle := pterm.NewStyle(pterm.Bold)
	versionStyle := pterm.NewStyle(pterm.FgLightBlue)

	fmt.Println(logoStyle.Sprint(banner))
	fmt.Println(taglineStyle.Sprint("Go-based Versatile Runtime & Development"))
	fmt.Println(taglineStyle.Sprint("========================================"))
	fmt.Println(versionStyle.Sprint(fmt.Sprintf("v%s", version)))
}
