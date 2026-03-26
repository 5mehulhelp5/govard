package engine

import (
	"sort"
	"strings"
)

type Features struct {
	Xdebug        bool `yaml:"xdebug"`
	Varnish       bool `yaml:"varnish"`
	Redis         bool `yaml:"redis,omitempty"`
	Elasticsearch bool `yaml:"elasticsearch,omitempty"`
	Isolated      bool `yaml:"isolated,omitempty"`
	MFTF          bool `yaml:"mftf,omitempty"`
	LiveReload    bool `yaml:"livereload,omitempty"`
}

type Services struct {
	WebServer string `yaml:"web_server"`
	Search    string `yaml:"search"`
	Cache     string `yaml:"cache"`
	Queue     string `yaml:"queue"`
}

type Stack struct {
	PHPVersion      string   `yaml:"php_version"`
	NodeVersion     string   `yaml:"node_version"`
	DBType          string   `yaml:"db_type"`
	DBVersion       string   `yaml:"db_version"`
	WebRoot         string   `yaml:"web_root"`
	NginxVersion    string   `yaml:"nginx_version,omitempty"`
	ApacheVersion   string   `yaml:"apache_version,omitempty"`
	CacheVersion    string   `yaml:"cache_version"`
	SearchVersion   string   `yaml:"search_version"`
	VarnishVersion  string   `yaml:"varnish_version,omitempty"`
	QueueVersion    string   `yaml:"queue_version,omitempty"`
	ComposerVersion string   `yaml:"composer_version,omitempty"`
	XdebugSession   string   `yaml:"xdebug_session,omitempty"`
	WebServer       string   `yaml:"web_server,omitempty"`
	UserID          int      `yaml:"user_id,omitempty"`
	GroupID         int      `yaml:"group_id,omitempty"`
	Services        Services `yaml:"services"`
	Features        Features `yaml:"features"`
	ChownDirList    []string `yaml:"chown_dir_list,omitempty"`
}

type Config struct {
	ProjectName      string              `yaml:"project_name"`
	Profile          string              `yaml:"profile,omitempty"`
	Framework        string              `yaml:"framework"`
	FrameworkVersion string              `yaml:"framework_version,omitempty"`
	Domain           string              `yaml:"domain"`
	ExtraDomains     []string            `yaml:"extra_domains,omitempty"`
	StoreDomains     StoreDomainMappings `yaml:"store_domains,omitempty"`

	Lock              LockConfig              `yaml:"lock,omitempty"`
	BlueprintRegistry BlueprintRegistryConfig `yaml:"blueprint_registry,omitempty"`
	Stack             Stack                   `yaml:"stack"`
	Remotes           map[string]RemoteConfig `yaml:"remotes"`
	Hooks             map[string][]HookStep   `yaml:"hooks,omitempty"`
}

type LockConfig struct {
	Strict bool `yaml:"strict,omitempty"`
}

type BlueprintRegistryConfig struct {
	Provider string `yaml:"provider,omitempty"`
	URL      string `yaml:"url,omitempty"`
	Ref      string `yaml:"ref,omitempty"`
	Checksum string `yaml:"checksum,omitempty"`
	Trusted  bool   `yaml:"trusted,omitempty"`
}

type HookStep struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

// AllDomains returns the primary Domain followed by any non-duplicate
// ExtraDomains and StoreDomains. It trims whitespace from each domain,
// skips empty strings, and keeps StoreDomains in sorted order so
// downstream config rendering stays deterministic.
func (c Config) AllDomains() []string {
	seen := make(map[string]bool)
	domains := []string{}

	primary := strings.TrimSpace(c.Domain)
	if primary != "" {
		domains = append(domains, primary)
		seen[primary] = true
	}

	for _, domain := range c.ExtraDomains {
		trimmed := strings.TrimSpace(domain)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		domains = append(domains, trimmed)
		seen[trimmed] = true
	}

	storeDomains := make([]string, 0, len(c.StoreDomains))
	for domain := range c.StoreDomains {
		storeDomains = append(storeDomains, domain)
	}
	sort.Strings(storeDomains)

	for _, domain := range storeDomains {
		trimmed := strings.TrimSpace(domain)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		domains = append(domains, trimmed)
		seen[trimmed] = true
	}

	return domains
}
