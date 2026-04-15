package tests

import (
	"strings"
	"testing"

	"govard/internal/cmd"
	"govard/internal/conventions"
	"govard/internal/engine"

	"github.com/spf13/cobra"
)

func TestRunBootstrapMagentoSetupInstallForTestUsesDefaultAdminEmailWhenDomainMissing(t *testing.T) {
	var capturedArgs [][]string
	restore := cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		copied := append([]string(nil), args...)
		capturedArgs = append(capturedArgs, copied)
		return nil
	})
	defer restore()

	err := cmd.RunBootstrapMagentoSetupInstallForTest(&cobra.Command{}, engine.Config{
		ProjectName: "sample-project",
		Framework:   conventions.FrameworkMagento2,
	}, "remote", "")
	if err != nil {
		t.Fatalf("RunBootstrapMagentoSetupInstallForTest() error = %v", err)
	}

	var setupInstallArgs []string
	for _, args := range capturedArgs {
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "setup:install") {
			setupInstallArgs = args
			break
		}
	}
	if len(setupInstallArgs) == 0 {
		t.Fatalf("expected magento setup:install subcommand, got %#v", capturedArgs)
	}

	joined := strings.Join(setupInstallArgs, " ")
	expected := "--admin-email=" + conventions.DefaultAdminEmail
	if !strings.Contains(joined, expected) {
		t.Fatalf("expected %q in setup args, got %q", expected, joined)
	}
}
