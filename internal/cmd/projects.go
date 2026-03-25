package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"govard/internal/engine"

	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"prj"},
	Short:   "Browse known projects from registry",
}

var projectsOpenCmd = &cobra.Command{
	Use:   "open <query>",
	Short: "Find a project by fuzzy query and print its path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectsOpen(cmd, args[0])
	},
}

func initProjectsCommands() {
	projectsCmd.AddCommand(projectsOpenCmd)
}

func runProjectsOpen(cmd *cobra.Command, query string) error {
	entries, err := engine.ReadProjectRegistryEntries()
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("project registry is empty; run govard init/up in at least one project first")
	}

	match, ok := findBestProjectMatch(entries, query)
	if !ok {
		return fmt.Errorf("no project matches query %q", strings.TrimSpace(query))
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), match.Path)
	return nil
}

func findBestProjectMatch(entries []engine.ProjectRegistryEntry, query string) (engine.ProjectRegistryEntry, bool) {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery == "" {
		return engine.ProjectRegistryEntry{}, false
	}

	bestIndex := -1
	bestScore := 0
	for i, entry := range entries {
		score, ok := scoreProjectEntry(entry, normalizedQuery)
		if !ok {
			continue
		}
		if bestIndex == -1 || score < bestScore {
			bestIndex = i
			bestScore = score
		}
	}
	if bestIndex < 0 {
		return engine.ProjectRegistryEntry{}, false
	}
	return entries[bestIndex], true
}

func scoreProjectEntry(entry engine.ProjectRegistryEntry, query string) (int, bool) {
	candidates := []struct {
		value string
		bias  int
	}{
		{value: strings.ToLower(strings.TrimSpace(entry.ProjectName)), bias: 0},
		{value: strings.ToLower(strings.TrimSpace(entry.Domain)), bias: 5},
		{value: strings.ToLower(filepath.Base(strings.TrimSpace(entry.Path))), bias: 8},
		{value: strings.ToLower(strings.TrimSpace(entry.Path)), bias: 12},
	}

	best := 0
	matched := false
	for _, candidate := range candidates {
		score, ok := scoreFieldMatch(candidate.value, query)
		if !ok {
			continue
		}
		score += candidate.bias
		if !matched || score < best {
			best = score
			matched = true
		}
	}
	if !matched {
		return 0, false
	}
	return best, true
}

func scoreFieldMatch(value string, query string) (int, bool) {
	if value == "" || query == "" {
		return 0, false
	}

	if value == query {
		return 0, true
	}
	if strings.HasPrefix(value, query) {
		return 10 + len(value) - len(query), true
	}
	if index := strings.Index(value, query); index >= 0 {
		return 30 + index, true
	}

	gap, ok := subsequenceGap(value, query)
	if !ok {
		return 0, false
	}
	return 100 + gap, true
}

func subsequenceGap(value string, query string) (int, bool) {
	if query == "" {
		return 0, false
	}

	totalGap := 0
	searchFrom := 0
	for i := 0; i < len(query); i++ {
		next := strings.IndexByte(value[searchFrom:], query[i])
		if next < 0 {
			return 0, false
		}
		totalGap += next
		searchFrom += next + 1
		if searchFrom > len(value) {
			return 0, false
		}
	}
	return totalGap, true
}
