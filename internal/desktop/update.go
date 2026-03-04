package desktop

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const desktopUpdateCheckLatestURLEnvVar = "GOVARD_UPDATE_CHECK_URL"
const desktopBinaryName = "govard-desktop"

var desktopUpdateHTTPClient = &http.Client{Timeout: 5 * time.Second}
var desktopExecutablePath = os.Executable
var desktopBinaryLookPath = exec.LookPath

var defaultRunDesktopSelfUpdate = func() (string, error) {
	binary, err := exec.LookPath("govard")
	if err != nil {
		return "", fmt.Errorf("govard CLI not found in PATH")
	}

	cmd := exec.Command(binary, "self-update")
	cmd.Env = append(os.Environ(), "GOVARD_SELF_UPDATE_CONFIRM=yes")

	output, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(output))
	if err != nil {
		if trimmed != "" {
			return "", fmt.Errorf("%v: %s", err, trimmed)
		}
		return "", err
	}
	return trimmed, nil
}

var runDesktopSelfUpdate = defaultRunDesktopSelfUpdate

var defaultRestartDesktopBinary = func(binaryPath string) error {
	cmd := exec.Command(binaryPath)
	cmd.Env = os.Environ()
	return cmd.Start()
}

var restartDesktopBinary = defaultRestartDesktopBinary

func (app *App) CheckForUpdates() (UpdateCheckResult, error) {
	current := normalizeDesktopVersionTag(Version)
	result := UpdateCheckResult{
		CurrentVersion: current,
	}

	latest, err := fetchDesktopLatestReleaseTag()
	if err != nil {
		result.Message = "Could not check for updates."
		return result, err
	}

	result.LatestVersion = latest
	result.Outdated = shouldDesktopNotifyUpdate(current, latest)
	if result.Outdated {
		result.Message = fmt.Sprintf("Update available: %s -> %s", current, latest)
		return result, nil
	}

	result.Message = fmt.Sprintf("Govard Desktop is up to date (%s).", current)
	return result, nil
}

func (app *App) InstallLatestUpdate() (string, error) {
	if runtime.GOOS == "windows" {
		return "", errors.New("automatic update is not supported on Windows yet; install a fresh release instead")
	}

	output, err := runDesktopSelfUpdate()
	if err != nil {
		return "", fmt.Errorf("install latest update: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return "Update completed. Restart Govard Desktop to run the new version.", nil
	}
	return output, nil
}

func (app *App) RestartDesktopApp() (string, error) {
	binaryPath, err := resolveDesktopBinaryForRestart()
	if err != nil {
		return "", fmt.Errorf("resolve desktop binary for restart: %w", err)
	}

	if err := restartDesktopBinary(binaryPath); err != nil {
		return "", fmt.Errorf("restart desktop app: %w", err)
	}

	if app != nil && app.ctx != nil {
		// Give the RPC layer a moment to flush the response before quitting.
		go func() {
			time.Sleep(250 * time.Millisecond)
			quitApplication(app.ctx)
		}()
	}

	return "Restarting Govard Desktop...", nil
}

func resolveDesktopBinaryForRestart() (string, error) {
	candidates := []string{}

	if executable, err := desktopExecutablePath(); err == nil {
		if resolved, resolveErr := filepath.EvalSymlinks(executable); resolveErr == nil {
			executable = resolved
		}
		candidates = append(candidates, executable)
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), desktopBinaryName))
	}

	if lookedUp, err := desktopBinaryLookPath(desktopBinaryName); err == nil {
		candidates = append(candidates, lookedUp)
	}

	seen := map[string]bool{}
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			continue
		}
		if resolved, resolveErr := filepath.EvalSymlinks(trimmed); resolveErr == nil {
			trimmed = resolved
		}
		if seen[trimmed] {
			continue
		}
		seen[trimmed] = true

		if stat, err := os.Stat(trimmed); err == nil && !stat.IsDir() {
			return trimmed, nil
		}
	}

	return "", fmt.Errorf("%s not found", desktopBinaryName)
}

func fetchDesktopLatestReleaseTag() (string, error) {
	url := desktopUpdateLatestURL()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("prepare update check request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "govard-desktop")

	resp, err := desktopUpdateHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request latest release failed with status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode latest release payload: %w", err)
	}

	latest := normalizeDesktopVersionTag(release.TagName)
	if latest == "" {
		return "", errors.New("latest release payload does not include tag_name")
	}

	return latest, nil
}

func desktopUpdateLatestURL() string {
	if override := strings.TrimSpace(os.Getenv(desktopUpdateCheckLatestURLEnvVar)); override != "" {
		return strings.TrimRight(override, "/")
	}
	return "https://api.github.com/repos/ddtcorex/govard/releases/latest"
}

func normalizeDesktopVersionTag(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "v") {
		return trimmed
	}
	return "v" + trimmed
}

func shouldDesktopNotifyUpdate(currentVersion, latestTag string) bool {
	latest := normalizeDesktopVersionTag(latestTag)
	if latest == "" {
		return false
	}
	current := normalizeDesktopVersionTag(currentVersion)
	if current == "" {
		return true
	}
	return latest != current
}
