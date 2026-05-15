package engine

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed profiles.json
var profilesJSON embed.FS

// RuntimeProfileFixture represents a single test case for version resolution.
type RuntimeProfileFixture struct {
	Name            string            `json:"name"`
	Framework       string            `json:"framework"`
	Version         string            `json:"version"`
	Source          string            `json:"source"`
	SourcePrefix    string            `json:"source_prefix"`
	WarningContains string            `json:"warning_contains"`
	Expected        map[string]string `json:"expected"`
	ExpectError     bool              `json:"expect_error"`
}

type profileStack struct {
	SearchVersion  string `json:"search_version"`
	VarnishVersion string `json:"varnish_version"`
	NginxVersion   string `json:"nginx_version"`
	QueueVersion   string `json:"queue_version"`
	CacheVersion   string `json:"cache_version"`
}

type profileRule struct {
	Min             int    `json:"min"`
	Stack           string `json:"stack"`
	PHPVersion      string `json:"php_version"`
	DBVersion       string `json:"db_version"`
	Cache           string `json:"cache,omitempty"`
	Search          string `json:"search,omitempty"`
	SearchVersion   string `json:"search_version,omitempty"`
	VarnishVersion  string `json:"varnish_version"`
	NginxVersion    string `json:"nginx_version"`
	QueueVersion    string `json:"queue_version"`
	CacheVersion    string `json:"cache_version"`
	ComposerVersion string `json:"composer_version,omitempty"`
}

type patchVariant struct {
	Patch           *int          `json:"patch,omitempty"`
	PatchMin        *int          `json:"patch_min,omitempty"`
	PatchMax        *int          `json:"patch_max,omitempty"`
	PHPVersion      string        `json:"php_version"`
	DBVersion       string        `json:"db_version"`
	Cache           string        `json:"cache,omitempty"`
	CacheVersion    string        `json:"cache_version,omitempty"`
	Search          string        `json:"search,omitempty"`
	SearchVersion   string        `json:"search_version,omitempty"`
	QueueVersion    string        `json:"queue_version"`
	VarnishVersion  string        `json:"varnish_version"`
	NginxVersion    string        `json:"nginx_version"`
	ComposerVersion string        `json:"composer_version,omitempty"`
	Rules           []profileRule `json:"rules"`
}

type versionGroup struct {
	Major    int               `json:"major"`
	Minor    int               `json:"minor"`
	Defaults map[string]string `json:"defaults"`
	Patches  []patchVariant    `json:"patches"`
}

type frameworkRule struct {
	Major           int    `json:"major"`
	MinorMin        *int   `json:"minor_min,omitempty"`
	PHPVersion      string `json:"php_version"`
	DBType          string `json:"db_type"`
	DBVersion       string `json:"db_version"`
	ComposerVersion string `json:"composer_version,omitempty"`
}

type profileRegistryData struct {
	Magento struct {
		Stacks   map[string]profileStack `json:"stacks"`
		Versions []versionGroup          `json:"versions"`
	} `json:"magento2"`
	Frameworks   map[string][]frameworkRule `json:"frameworks"`
	TestFixtures []RuntimeProfileFixture    `json:"test_fixtures"`
}

var registry profileRegistryData

func init() {
	if data, err := profilesJSON.ReadFile("profiles.json"); err == nil {
		_ = json.Unmarshal(data, &registry)
	}
}

// GetFrameworkTestFixtures returns the test cases embedded in both registries.
func GetFrameworkTestFixtures() []RuntimeProfileFixture {
	return registry.TestFixtures
}

// resolveFrameworkProfileFromRegistry looks up the technology stack for other frameworks.
func resolveFrameworkProfileFromRegistry(framework string, major int, minor int) (runtimeProfileOverride, string, bool) {
	rules, ok := registry.Frameworks[framework]
	if !ok {
		return runtimeProfileOverride{}, "", false
	}

	for _, rule := range rules {
		if rule.Major != major {
			continue
		}

		if rule.MinorMin != nil && minor < *rule.MinorMin {
			continue
		}

		override := runtimeProfileOverride{
			PHPVersion:      rule.PHPVersion,
			DBType:          rule.DBType,
			DBVersion:       rule.DBVersion,
			ComposerVersion: rule.ComposerVersion,
		}

		var source string
		if rule.MinorMin != nil {
			source = fmt.Sprintf("version-specific:%s@%d.%d", framework, major, *rule.MinorMin)
		} else {
			source = fmt.Sprintf("version-specific:%s@%d", framework, major)
		}

		return override, source, true
	}

	return runtimeProfileOverride{}, "", false
}

// resolveMagentoProfileFromRegistry looks up the technology stack for a given Magento version using the JSON registry.
func resolveMagentoProfileFromRegistry(major, minor, patch, pPatch int) (runtimeProfileOverride, string, bool) {
	for _, group := range registry.Magento.Versions {
		if group.Major != major || group.Minor != minor {
			continue
		}

		for _, v := range group.Patches {
			if v.Patch != nil && *v.Patch != patch {
				continue
			}
			if v.PatchMin != nil && patch < *v.PatchMin {
				continue
			}
			if v.PatchMax != nil && patch > *v.PatchMax {
				continue
			}

			// Found the specific patch group (e.g., 2.4.7)
			override := runtimeProfileOverride{
				DBType:     group.Defaults["db_type"],
				Cache:      group.Defaults["cache"],
				Search:     group.Defaults["search"],
				Queue:      group.Defaults["queue"],
				WebRoot:    group.Defaults["web_root"],
				PHPVersion: v.PHPVersion,
				DBVersion:  v.DBVersion,
			}

			// Apply baseline patch versions
			applyPatchBaselines(&override, v)

			// Resolve rule (find the highest min pPatch that matches)
			var activeRule *profileRule
			for i := range v.Rules {
				if pPatch >= v.Rules[i].Min {
					activeRule = &v.Rules[i]
					break
				}
			}

			if activeRule != nil {
				// Apply stack if defined
				if activeRule.Stack != "" {
					if stack, ok := registry.Magento.Stacks[activeRule.Stack]; ok {
						applyStackToOverride(&override, stack)
					}
				}
				// Apply rule-specific overrides
				applyRuleOverrides(&override, *activeRule)
			}

			return override, fmt.Sprintf("version-specific:magento2@%d.%d.%d-p%d", major, minor, patch, pPatch), true
		}
	}

	return runtimeProfileOverride{}, "", false
}

func applyPatchBaselines(o *runtimeProfileOverride, v patchVariant) {
	if v.Cache != "" {
		o.Cache = v.Cache
	}
	if v.Search != "" {
		o.Search = v.Search
	}
	if v.CacheVersion != "" {
		o.CacheVersion = v.CacheVersion
	}
	if v.SearchVersion != "" {
		o.SearchVersion = v.SearchVersion
	}
	if v.QueueVersion != "" {
		o.QueueVersion = v.QueueVersion
	}
	if v.VarnishVersion != "" {
		o.VarnishVersion = v.VarnishVersion
	}
	if v.NginxVersion != "" {
		o.NginxVersion = v.NginxVersion
	}
	if v.ComposerVersion != "" {
		o.ComposerVersion = v.ComposerVersion
	}
}

func applyStackToOverride(o *runtimeProfileOverride, stack profileStack) {
	if stack.SearchVersion != "" {
		o.SearchVersion = stack.SearchVersion
	}
	if stack.VarnishVersion != "" {
		o.VarnishVersion = stack.VarnishVersion
	}
	if stack.NginxVersion != "" {
		o.NginxVersion = stack.NginxVersion
	}
	if stack.QueueVersion != "" {
		o.QueueVersion = stack.QueueVersion
	}
	if stack.CacheVersion != "" {
		o.CacheVersion = stack.CacheVersion
	}
}

func applyRuleOverrides(o *runtimeProfileOverride, rule profileRule) {
	if rule.PHPVersion != "" {
		o.PHPVersion = rule.PHPVersion
	}
	if rule.DBVersion != "" {
		o.DBVersion = rule.DBVersion
	}
	if rule.Cache != "" {
		o.Cache = rule.Cache
	}
	if rule.Search != "" {
		o.Search = rule.Search
	}
	if rule.SearchVersion != "" {
		o.SearchVersion = rule.SearchVersion
	}
	if rule.ComposerVersion != "" {
		o.ComposerVersion = rule.ComposerVersion
	}
	if rule.VarnishVersion != "" {
		o.VarnishVersion = rule.VarnishVersion
	}
	if rule.NginxVersion != "" {
		o.NginxVersion = rule.NginxVersion
	}
	if rule.QueueVersion != "" {
		o.QueueVersion = rule.QueueVersion
	}
	if rule.CacheVersion != "" {
		o.CacheVersion = rule.CacheVersion
	}
}
