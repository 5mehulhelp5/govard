package updater

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

const updateCheckLatestURLEnvVar = "GOVARD_UPDATE_CHECK_URL"

var (
	updateCheckHTTPClient = &http.Client{Timeout: 2 * time.Second}
	updateCheckNotifier   = func(latestTag, currentVersion string) {
		pterm.Warning.Printf("A new version of Govard is available: %s (current: %s)\n", latestTag, currentVersion)
		pterm.Info.Println("Run 'govard self-update' to upgrade.")
	}
)

func CheckForUpdates(current string) {
	resp, err := updateCheckHTTPClient.Get(updateCheckLatestURL())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	if shouldNotifyUpdate(current, release.TagName) {
		updateCheckNotifier(release.TagName, current)
	}
}

func updateCheckLatestURL() string {
	if override := strings.TrimSpace(os.Getenv(updateCheckLatestURLEnvVar)); override != "" {
		return override
	}
	return "https://api.github.com/repos/ddtcorex/govard/releases/latest"
}

func shouldNotifyUpdate(currentVersion, latestTag string) bool {
	latest := strings.TrimSpace(latestTag)
	if latest == "" {
		return false
	}
	current := strings.TrimSpace(currentVersion)
	if current == "" {
		return true
	}

	// If current version is a development build (e.g. 1.31.0-2-gf2a0be7),
	// check if the base version matches the latest tag to avoid redundant warnings.
	if strings.Contains(current, "-") {
		base := strings.SplitN(current, "-", 2)[0]
		if latest == "v"+base {
			return false
		}
	}

	return latest != "v"+current
}

// SetUpdateCheckHTTPClientForTest overrides the HTTP client used by update checks.
func SetUpdateCheckHTTPClientForTest(client *http.Client) func() {
	previous := updateCheckHTTPClient
	if client != nil {
		updateCheckHTTPClient = client
	}
	return func() {
		updateCheckHTTPClient = previous
	}
}

// SetUpdateCheckNotifierForTest overrides update notification side effects.
func SetUpdateCheckNotifierForTest(fn func(latestTag, currentVersion string)) func() {
	previous := updateCheckNotifier
	if fn != nil {
		updateCheckNotifier = fn
	}
	return func() {
		updateCheckNotifier = previous
	}
}

// ShouldNotifyUpdateForTest exposes update comparison logic for tests in /tests.
func ShouldNotifyUpdateForTest(currentVersion, latestTag string) bool {
	return shouldNotifyUpdate(currentVersion, latestTag)
}
