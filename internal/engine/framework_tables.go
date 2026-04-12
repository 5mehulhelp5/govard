package engine

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed framework_tables.json
var frameworkTablesJSON []byte

type FrameworkTableConfig struct {
	Ignored   []string `json:"ignored"`
	Sensitive []string `json:"sensitive"`
}

var (
	frameworkTables     map[string]FrameworkTableConfig
	frameworkTablesOnce sync.Once
)

// GetFrameworkIgnoredTables returns the list of tables to ignore for a given framework
// based on whether noise (logs/cache) or PII (sensitive data) filters are active.
func GetFrameworkIgnoredTables(framework string, noNoise bool, noPII bool) []string {
	if !noNoise && !noPII {
		return nil
	}

	frameworkTablesOnce.Do(func() {
		_ = json.Unmarshal(frameworkTablesJSON, &frameworkTables)
	})

	fw := strings.TrimSpace(framework)
	if fw == "" {
		fw = "magento2"
	}

	config, ok := frameworkTables[fw]
	if !ok {
		// Fallback to magento2 standard if framework not recognized
		config = frameworkTables["magento2"]
	}

	tables := make([]string, 0)
	if noNoise {
		tables = append(tables, config.Ignored...)
	}
	if noPII {
		tables = append(tables, config.Sensitive...)
	}
	return tables
}
