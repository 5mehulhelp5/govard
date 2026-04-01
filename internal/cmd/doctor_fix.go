package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"govard/internal/engine"

	"github.com/pterm/pterm"
)

const (
	DoctorFixStatusApplied     = "applied"
	DoctorFixStatusFailed      = "failed"
	DoctorFixStatusUnavailable = "unavailable"
)

// DoctorFixResult captures the outcome of a single doctor --fix attempt.
type DoctorFixResult struct {
	CheckID string   `json:"check_id"`
	Title   string   `json:"title"`
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Actions []string `json:"actions,omitempty"`
}

var doctorDependencies = engine.DoctorDependencies{}

var doctorRunDiagnostics = engine.RunDoctorDiagnostics

type doctorFixHandler func(check engine.DoctorCheck) DoctorFixResult

var doctorFixHandlers = map[string]doctorFixHandler{
	"host.govard.home":        fixGovardHomeDirectory,
	"host.search.index_block": unblockSearchIndex,
	"host.compose.spam":       purgeStaleComposeFiles,
	"host.govard.registry":    fixGovardRegistry,
}

func runDoctorDiagnostics() engine.DoctorReport {
	return doctorRunDiagnostics(doctorDependencies)
}

func applyDoctorSafeFixes(report engine.DoctorReport) []DoctorFixResult {
	results := make([]DoctorFixResult, 0)
	seen := map[string]bool{}

	for _, check := range report.Checks {
		if check.Status == engine.DoctorStatusPass {
			continue
		}

		checkID := strings.TrimSpace(check.ID)
		if checkID == "" || seen[checkID] {
			continue
		}
		seen[checkID] = true

		handler, ok := doctorFixHandlers[checkID]
		if !ok {
			results = append(results, DoctorFixResult{
				CheckID: checkID,
				Title:   strings.TrimSpace(check.Title),
				Status:  DoctorFixStatusUnavailable,
				Message: "No safe automatic fix available.",
			})
			continue
		}

		results = append(results, handler(check))
	}

	return results
}

func fixGovardHomeDirectory(check engine.DoctorCheck) DoctorFixResult {
	result := DoctorFixResult{
		CheckID: strings.TrimSpace(check.ID),
		Title:   strings.TrimSpace(check.Title),
		Status:  DoctorFixStatusApplied,
		Message: "Govard runtime directories prepared.",
		Actions: []string{},
	}

	homeDir := filepath.Clean(engine.GovardHomeDir())
	dirs := []string{
		homeDir,
		filepath.Join(homeDir, "compose"),
		filepath.Join(homeDir, "diagnostics"),
	}

	for _, dir := range dirs {
		result.Actions = append(result.Actions, fmt.Sprintf("mkdir -p %s", dir))
		if err := os.MkdirAll(dir, 0o700); err != nil {
			result.Status = DoctorFixStatusFailed
			result.Message = err.Error()
			return result
		}
		result.Actions = append(result.Actions, fmt.Sprintf("chmod 700 %s", dir))
		if err := os.Chmod(dir, 0o700); err != nil {
			result.Status = DoctorFixStatusFailed
			result.Message = err.Error()
			return result
		}
	}

	result.Actions = append(result.Actions, fmt.Sprintf("write probe file in %s", homeDir))
	if err := engine.CheckGovardHomeWritable(); err != nil {
		result.Status = DoctorFixStatusFailed
		result.Message = err.Error()
		return result
	}

	return result
}

func unblockSearchIndex(check engine.DoctorCheck) DoctorFixResult {
	result := DoctorFixResult{
		CheckID: strings.TrimSpace(check.ID),
		Title:   strings.TrimSpace(check.Title),
		Status:  DoctorFixStatusApplied,
		Message: "Elasticsearch/OpenSearch index unblocked.",
		Actions: []string{},
	}

	config := loadConfig()
	result.Actions = append(result.Actions, "unblock search index via docker exec curl")
	if err := engine.FixElasticsearchIndexBlock(config.ProjectName, config); err != nil {
		result.Status = DoctorFixStatusFailed
		result.Message = err.Error()
		return result
	}

	return result
}

func purgeStaleComposeFiles(check engine.DoctorCheck) DoctorFixResult {
	result := DoctorFixResult{
		CheckID: strings.TrimSpace(check.ID),
		Title:   strings.TrimSpace(check.Title),
		Status:  DoctorFixStatusApplied,
		Message: "Stale govard compose files purged.",
		Actions: []string{},
	}

	result.Actions = append(result.Actions, "Purging compose files older than 7 days")
	count, err := engine.CleanupStaleComposeFiles(7 * 24 * time.Hour)
	if err != nil {
		result.Status = DoctorFixStatusFailed
		result.Message = err.Error()
		return result
	}

	result.Message = fmt.Sprintf("Removed %d stale compose file(s).", count)
	return result
}

func fixGovardRegistry(check engine.DoctorCheck) DoctorFixResult {
	result := DoctorFixResult{
		CheckID: strings.TrimSpace(check.ID),
		Title:   strings.TrimSpace(check.Title),
		Status:  DoctorFixStatusApplied,
		Message: "Project registry path restored.",
		Actions: []string{},
	}

	path := engine.ProjectRegistryPath()
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return result
		}
		result.Status = DoctorFixStatusFailed
		result.Message = err.Error()
		return result
	}

	if !info.IsDir() {
		return result
	}

	result.Actions = append(result.Actions, fmt.Sprintf("rm -rf %s", path))
	if err := os.RemoveAll(path); err != nil {
		result.Status = DoctorFixStatusFailed
		result.Message = err.Error()
		return result
	}

	return result
}

func renderDoctorFixResults(results []DoctorFixResult) {
	if len(results) == 0 {
		pterm.Info.Println("No safe fixes were available.")
		return
	}

	applied := 0
	failed := 0
	unavailable := 0

	for _, result := range results {
		line := fmt.Sprintf("%s (%s): %s", result.Title, result.CheckID, result.Message)
		switch result.Status {
		case DoctorFixStatusApplied:
			applied++
			pterm.Success.Println(line)
		case DoctorFixStatusFailed:
			failed++
			pterm.Error.Println(line)
		default:
			unavailable++
			pterm.Warning.Println(line)
		}

		for _, action := range result.Actions {
			pterm.Info.Printf("Action: %s\n", action)
		}
	}

	pterm.Info.Printf(
		"Fix summary: applied=%d failed=%d unavailable=%d\n",
		applied,
		failed,
		unavailable,
	)
}

func summarizeDoctorFixResults(results []DoctorFixResult) []string {
	if len(results) == 0 {
		return []string{"Doctor --fix: no safe fixes were available."}
	}

	lines := make([]string, 0, len(results)+1)
	applied := 0
	failed := 0
	unavailable := 0

	for _, result := range results {
		switch result.Status {
		case DoctorFixStatusApplied:
			applied++
		case DoctorFixStatusFailed:
			failed++
		default:
			unavailable++
		}
		lines = append(lines, fmt.Sprintf("Doctor --fix: %s (%s): %s", result.Title, result.CheckID, result.Message))
		for _, action := range result.Actions {
			lines = append(lines, fmt.Sprintf("Doctor --fix action: %s", action))
		}
	}

	lines = append(lines, fmt.Sprintf("Doctor --fix summary: applied=%d failed=%d unavailable=%d", applied, failed, unavailable))
	return lines
}

// SetDoctorDependenciesForTest overrides doctor diagnostics dependencies for tests.
func SetDoctorDependenciesForTest(dependencies engine.DoctorDependencies) func() {
	previous := doctorDependencies
	doctorDependencies = dependencies
	return func() {
		doctorDependencies = previous
	}
}

// ApplyDoctorSafeFixesForTest exposes doctor fix planning/execution for tests.
func ApplyDoctorSafeFixesForTest(report engine.DoctorReport) []DoctorFixResult {
	return applyDoctorSafeFixes(report)
}
