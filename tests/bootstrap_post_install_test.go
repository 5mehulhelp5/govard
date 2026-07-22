package tests

import (
	"reflect"
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

func TestRunBootstrapMagentoSetupInstallForTestUsesConfigTablePrefix(t *testing.T) {
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
		TablePrefix: "demo_",
	}, "remote", "")
	if err != nil {
		t.Fatalf("RunBootstrapMagentoSetupInstallForTest() error = %v", err)
	}

	joined := strings.Join(flattenArgs(capturedArgs), " ")
	if !strings.Contains(joined, "--db-prefix=demo_") {
		t.Fatalf("expected setup args to contain table prefix, got %q", joined)
	}
}

func TestRunBootstrapMagentoSetupInstallForTestUsesMagentoDBCredentialsForMagento2(t *testing.T) {
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

	joined := strings.Join(flattenArgs(capturedArgs), " ")
	if !strings.Contains(joined, "--db-name=magento") || !strings.Contains(joined, "--db-user=magento") || !strings.Contains(joined, "--db-password=magento") {
		t.Fatalf("expected magento db credentials in setup args, got %q", joined)
	}
}

func TestRunBootstrapMagentoSetupInstallForTestUsesMageOSDBCredentials(t *testing.T) {
	var capturedArgs [][]string
	restore := cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		copied := append([]string(nil), args...)
		capturedArgs = append(capturedArgs, copied)
		return nil
	})
	defer restore()

	err := cmd.RunBootstrapMagentoSetupInstallForTest(&cobra.Command{}, engine.Config{
		ProjectName: "sample-project",
		Framework:   conventions.FrameworkMageOS,
	}, "remote", "")
	if err != nil {
		t.Fatalf("RunBootstrapMagentoSetupInstallForTest() error = %v", err)
	}

	joined := strings.Join(flattenArgs(capturedArgs), " ")
	if !strings.Contains(joined, "--db-name=mageos") || !strings.Contains(joined, "--db-user=mageos") || !strings.Contains(joined, "--db-password=mageos") {
		t.Fatalf("expected mageos db credentials in setup args, got %q", joined)
	}
}

// TestRunBootstrapMagentoSetupInstallForTestUsesMageOSDBCredentialsForLegacyVersion covers the
// second (version-comparison) setupArgs block, which independently hardcoded magento DB
// credentials before this fix.
func TestRunBootstrapMagentoSetupInstallForTestUsesMageOSDBCredentialsForLegacyVersion(t *testing.T) {
	var capturedArgs [][]string
	restore := cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		copied := append([]string(nil), args...)
		capturedArgs = append(capturedArgs, copied)
		return nil
	})
	defer restore()

	err := cmd.RunBootstrapMagentoSetupInstallForTest(&cobra.Command{}, engine.Config{
		ProjectName: "sample-project",
		Framework:   conventions.FrameworkMageOS,
	}, "remote", "2.4.7")
	if err != nil {
		t.Fatalf("RunBootstrapMagentoSetupInstallForTest() error = %v", err)
	}

	joined := strings.Join(flattenArgs(capturedArgs), " ")
	if !strings.Contains(joined, "--db-name=mageos") || !strings.Contains(joined, "--db-user=mageos") || !strings.Contains(joined, "--db-password=mageos") {
		t.Fatalf("expected mageos db credentials in legacy-version setup args, got %q", joined)
	}
	if !strings.Contains(joined, "--search-engine=opensearch") {
		t.Fatalf("expected Mage-OS 1.x to use OpenSearch, got %q", joined)
	}
}

func flattenArgs(groups [][]string) []string {
	result := make([]string, 0)
	for _, group := range groups {
		result = append(result, group...)
	}
	return result
}

func TestRunBootstrapHyvaInstallForTestRunsExpectedComposerCalls(t *testing.T) {
	calls := make([][]string, 0, 3)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()

	err := cmd.RunBootstrapHyvaInstallForTest(&cobra.Command{}, "token-123")
	if err != nil {
		t.Fatalf("RunBootstrapHyvaInstallForTest() error = %v", err)
	}

	want := [][]string{
		{"tool", "composer", "config", "http-basic.hyva-themes.repo.packagist.com", "token", "token-123"},
		{"tool", "composer", "config", "repositories.hyva-themes", "composer", "https://hyva-themes.repo.packagist.com/app-hyva-test-dv1dgx/"},
		{"tool", "composer", "require", "-n", "hyva-themes/magento2-default-theme"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("composer calls = %#v, want %#v", calls, want)
	}
}

func TestRunBootstrapMagentoSetupInstallForTestUsesElasticsearch7ForLegacyVersion(t *testing.T) {
	calls := make([][]string, 0, 1)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()

	err := cmd.RunBootstrapMagentoSetupInstallForTest(
		&cobra.Command{},
		engine.Config{Framework: "magento2", Domain: "sample.test"},
		"staging",
		"2.4.7",
	)
	if err != nil {
		t.Fatalf("RunBootstrapMagentoSetupInstallForTest() error = %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected one setup call, got %d", len(calls))
	}
	joined := strings.Join(calls[0], " ")
	if !strings.Contains(joined, "--search-engine=elasticsearch7") {
		t.Fatalf("expected elasticsearch7 engine for legacy versions, args: %s", joined)
	}
	if strings.Contains(joined, "--search-engine=opensearch") {
		t.Fatalf("did not expect opensearch args for legacy version: %s", joined)
	}
}

func TestRunBootstrapMagentoSetupInstallForTestUsesOpenSearchForRecentVersion(t *testing.T) {
	calls := make([][]string, 0, 1)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()

	err := cmd.RunBootstrapMagentoSetupInstallForTest(
		&cobra.Command{},
		engine.Config{Domain: "sample.test"},
		"staging",
		"2.4.8",
	)
	if err != nil {
		t.Fatalf("RunBootstrapMagentoSetupInstallForTest() error = %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected one setup call, got %d", len(calls))
	}
	joined := strings.Join(calls[0], " ")
	if !strings.Contains(joined, "--search-engine=opensearch") {
		t.Fatalf("expected opensearch args for 2.4.8+, got: %s", joined)
	}
}

func TestRunBootstrapSampleDataForTestRunsAllSteps(t *testing.T) {
	calls := make([][]string, 0, 4)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()

	err := cmd.RunBootstrapSampleDataForTest(&cobra.Command{})
	if err != nil {
		t.Fatalf("RunBootstrapSampleDataForTest() error = %v", err)
	}

	want := [][]string{
		{"tool", "magento", "sample:deploy"},
		{"tool", "magento", "setup:upgrade"},
		{"tool", "magento", "indexer:reindex"},
		{"tool", "magento", "cache:flush"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("sample data calls = %#v, want %#v", calls, want)
	}
}
