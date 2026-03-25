package tests

import (
	"os"
	"testing"

	"github.com/pterm/pterm"
)

func TestMain(m *testing.M) {
	pterm.DisableColor()
	originalGovardHome, hadGovardHome := os.LookupEnv("GOVARD_HOME_DIR")

	tempGovardHome, err := os.MkdirTemp("", "govard-tests-home-*")
	if err != nil {
		panic(err)
	}

	if err := os.Setenv("GOVARD_HOME_DIR", tempGovardHome); err != nil {
		panic(err)
	}

	code := m.Run()

	_ = os.RemoveAll(tempGovardHome)
	if hadGovardHome {
		_ = os.Setenv("GOVARD_HOME_DIR", originalGovardHome)
	} else {
		_ = os.Unsetenv("GOVARD_HOME_DIR")
	}

	os.Exit(code)
}
