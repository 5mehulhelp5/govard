package tests

import (
	"testing"

	"govard/internal/ui"
)

func TestUIPkgPrintHelpersDoNotPanic(t *testing.T) {
	ui.PrintSuccess("ok")
	ui.PrintError("error")
	ui.PrintInfo("info")
}
