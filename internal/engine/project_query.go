package engine

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FindProjectByQuery finds the best project match for a given query (name, domain, or path).
func FindProjectByQuery(query string) (ProjectRegistryEntry, error) {
	entries, err := ReadProjectRegistryEntries()
	if err != nil {
		return ProjectRegistryEntry{}, err
	}
	if len(entries) == 0 {
		return ProjectRegistryEntry{}, fmt.Errorf("project registry is empty")
	}

	match, ok := findBestProjectMatch(entries, query)
	if !ok {
		return ProjectRegistryEntry{}, fmt.Errorf("no project matches query %q", query)
	}

	return match, nil
}

func findBestProjectMatch(entries []ProjectRegistryEntry, query string) (ProjectRegistryEntry, bool) {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery == "" {
		return ProjectRegistryEntry{}, false
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
		return ProjectRegistryEntry{}, false
	}
	return entries[bestIndex], true
}

func scoreProjectEntry(entry ProjectRegistryEntry, query string) (int, bool) {
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
