package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readDocFile(t *testing.T, relativePath string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", relativePath))
	if err != nil {
		t.Fatalf("read doc %s: %v", relativePath, err)
	}
	return string(data)
}

func extractHeadingBlock(t *testing.T, content string, heading string) string {
	t.Helper()

	start := strings.Index(content, heading)
	if start < 0 {
		t.Fatalf("heading %q not found", heading)
	}
	rest := content[start+len(heading):]

	// Blocks in docs files are delimited by the next H2/H3 heading.
	nextH3 := strings.Index(rest, "\n### ")
	nextH2 := strings.Index(rest, "\n## ")
	end := len(rest)

	if nextH3 >= 0 && nextH3 < end {
		end = nextH3
	}
	if nextH2 >= 0 && nextH2 < end {
		end = nextH2
	}

	return rest[:end]
}

func TestDocsFrameworkCommandsMatchSupportedShortcuts(t *testing.T) {
	content := readDocFile(t, "docs/user/commands.md")

	magento2Block := extractHeadingBlock(t, content, "### Magento 2")
	if strings.Contains(magento2Block, "govard magerun") {
		t.Fatalf("docs/user/commands.md Magento 2 block must not advertise govard magerun")
	}

	for _, required := range []string{
		"### Drupal",
		"govard tool drush [command]",
		"### Symfony",
		"govard tool symfony [command]",
		"### Shopware",
		"govard tool shopware [command]",
		"### CakePHP",
		"govard tool cake [command]",
		"### WordPress",
		"govard tool wp [command]",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("framework command docs missing %q", required)
		}
	}
}

func TestDocsMagento2GuideUsesSupportedCLIReferences(t *testing.T) {
	content := readDocFile(t, "docs/frameworks/magento2.md")

	for _, banned := range []string{
		"govard db export",
		"govard magerun [command]",
		"Commands run as `magento` user",
	} {
		if strings.Contains(content, banned) {
			t.Fatalf("docs/frameworks/magento2.md contains outdated command/runtime claim %q", banned)
		}
	}

	if !strings.Contains(content, "govard db dump > dump.sql") {
		t.Fatalf("docs/frameworks/magento2.md should use supported db dump command examples")
	}
}

func TestDocsContributingFixturePathsMatchRepository(t *testing.T) {
	content := readDocFile(t, "docs/dev/contributing.md")

	for _, banned := range []string{
		"`init-projects/[framework]-init/`",
		"`framework-projects/[framework]/`",
		"Create `tests/init-projects/[framework]-init/`",
	} {
		if strings.Contains(content, banned) {
			t.Fatalf("docs/dev/contributing.md contains outdated fixture path %q", banned)
		}
	}

	for _, required := range []string{
		"`tests/fixtures/`",
		"`tests/integration/projects/`",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("docs/dev/contributing.md missing current fixture path %q", required)
		}
	}
}

func TestDocsUserCommandsGlobalSectionMatchesRootCLI(t *testing.T) {
	content := readDocFile(t, "docs/user/commands.md")

	if strings.Contains(content, "--verbose") {
		t.Fatalf("docs/user/commands.md must not advertise unsupported global flag --verbose")
	}

	if strings.Contains(content, "### `govard completion`") {
		t.Fatal("docs/user/commands.md must not advertise removed completion command")
	}
}
