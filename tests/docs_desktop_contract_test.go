package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readDoc(t *testing.T, relativePath string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", relativePath))
	if err != nil {
		t.Fatalf("read doc %s: %v", relativePath, err)
	}
	return string(data)
}

func TestDocsDesktopCommandReferenceTracksLightweightSurface(t *testing.T) {
	content := readDoc(t, "docs/desktop.md")

	for _, banned := range []string{
		"Role mode switch (`developer` and `pm_tester`)",
		"role-based UI visibility",
		"`pm_tester`",
	} {
		if strings.Contains(content, banned) {
			t.Fatalf("desktop command docs contain removed desktop feature claim %q", banned)
		}
	}

	for _, required := range []string{
		"Project workspace layout (environments, quick actions, onboarding)",
		"Quick actions (PHPMyAdmin, Xdebug toggle, health)",
		"Shell launcher (service, user, shell)",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("desktop docs missing lightweight desktop capability %q", required)
		}
	}
}

func TestDocsArchitectureDesktopSectionTracksLightweightCore(t *testing.T) {
	content := readDoc(t, "docs/architecture.md")

	for _, banned := range []string{
		"workflow snapshots",
		"operations panel",
		"Onboarding readiness",
		"setup wizard panel",
		"doctor pack export",
		"db dump/import",
	} {
		if strings.Contains(content, banned) {
			t.Fatalf("architecture docs contain removed desktop capability claim %q", banned)
		}
	}

	for _, required := range []string{
		"quick actions (start/stop/open, PHPMyAdmin, Xdebug toggle, health)",
		"workspace grouping environment list, quick actions, and onboarding",
		"logs with service filtering and live streaming",
		"shell launcher with project/service/user/shell selection",
		"modular frontend split across feature modules and bridge/state services",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("docs/architecture.md missing lightweight desktop architecture claim %q", required)
		}
	}
}
