package tests

import (
	"testing"

	"govard/internal/engine/bootstrap"
)

func TestMagento2BootstrapFreshArgs(t *testing.T) {
	cmds := bootstrap.Magento2FreshCommands(bootstrap.Options{Version: "2.4.8"})
	if len(cmds) == 0 {
		t.Fatal("expected commands")
	}
}
