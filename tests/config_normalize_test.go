package tests

import (
	"os"
	"strings"
	"testing"

	"govard/internal/engine"

	"gopkg.in/yaml.v3"
)

func TestNormalizeConfigDefaultsMagento2(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			Features: engine.Features{
				Cache:  true,
				Search: true,
			},
		},
	}

	engine.NormalizeConfig(&config, "")

	if config.Stack.DBType != "none" {
		t.Fatalf("Expected DBType none, got %s", config.Stack.DBType)
	}
	if config.Stack.DBVersion != "" {
		t.Fatalf("Expected DBVersion empty, got %s", config.Stack.DBVersion)
	}
	if config.Stack.PHPVersion != "8.5" {
		t.Fatalf("Expected PHPVersion 8.5, got %s", config.Stack.PHPVersion)
	}
	if config.Stack.NodeVersion != "24" {
		t.Fatalf("Expected NodeVersion 24, got %s", config.Stack.NodeVersion)
	}

	if config.Stack.Services.WebServer != "nginx" {
		t.Fatalf("Expected WebServer nginx, got %s", config.Stack.Services.WebServer)
	}
	if config.Stack.Services.Cache != "none" {
		t.Fatalf("Expected Cache none, got %s", config.Stack.Services.Cache)
	}

	if config.Stack.Services.Search != "none" {
		t.Fatalf("Expected Search none, got %s", config.Stack.Services.Search)
	}
	if config.Stack.Services.Queue != "none" {
		t.Fatalf("Expected Queue none, got %s", config.Stack.Services.Queue)
	}

	if config.Stack.CacheVersion != "" {
		t.Fatalf("Expected CacheVersion empty, got %s", config.Stack.CacheVersion)
	}

	if config.Stack.SearchVersion != "" {
		t.Fatalf("Expected SearchVersion empty, got %s", config.Stack.SearchVersion)
	}
	if config.Stack.QueueVersion != "" {
		t.Fatalf("Expected QueueVersion empty, got %s", config.Stack.QueueVersion)
	}
}

func TestNormalizeConfigQueueDefaults(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			Features: engine.Features{
				Queue: true,
			},
			Services: engine.Services{
				Queue: "rabbitmq",
			},
		},
	}

	engine.NormalizeConfig(&config, "")

	if config.Stack.QueueVersion != "4.2" {
		t.Fatalf("Expected QueueVersion 4.2, got %s", config.Stack.QueueVersion)
	}
}

func TestNormalizeConfigVersionAwareDefaultsMagento2(t *testing.T) {
	config := engine.Config{
		Framework:        "magento2",
		FrameworkVersion: "2.4.7-p3",
		Stack: engine.Stack{
			Services: engine.Services{
				DB:     "mariadb",
				Cache:  "redis",
				Search: "opensearch",
				Queue:  "rabbitmq",
			},
			Features: engine.Features{
				Cache:  true,
				Search: true,
				Queue:  true,
			},
		},
	}

	engine.NormalizeConfig(&config, "")

	if config.Stack.PHPVersion != "8.3" {
		t.Fatalf("Expected PHPVersion 8.3, got %s", config.Stack.PHPVersion)
	}
	if config.Stack.DBType != "mariadb" {
		t.Fatalf("Expected DBType mariadb, got %s", config.Stack.DBType)
	}
	if config.Stack.DBVersion != "10.6" {
		t.Fatalf("Expected DBVersion 10.6, got %s", config.Stack.DBVersion)
	}
	if config.Stack.Services.Cache != "redis" {
		t.Fatalf("Expected Cache redis, got %s", config.Stack.Services.Cache)
	}
	if config.Stack.CacheVersion != "7.2" {
		t.Fatalf("Expected CacheVersion 7.2, got %s", config.Stack.CacheVersion)
	}
	if config.Stack.Services.Search != "opensearch" {
		t.Fatalf("Expected Search opensearch, got %s", config.Stack.Services.Search)
	}
	if config.Stack.SearchVersion != "2.12" {
		t.Fatalf("Expected SearchVersion 2.12, got %s", config.Stack.SearchVersion)
	}
	if config.Stack.Services.Queue != "rabbitmq" {
		t.Fatalf("Expected Queue rabbitmq, got %s", config.Stack.Services.Queue)
	}
	if config.Stack.QueueVersion != "3.12" {
		t.Fatalf("Expected QueueVersion 3.12, got %s", config.Stack.QueueVersion)
	}
	if config.Stack.WebRoot != "/pub" {
		t.Fatalf("Expected WebRoot /pub, got %s", config.Stack.WebRoot)
	}
}

func TestNormalizeConfigCacheAndSearchVersions(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			Features: engine.Features{
				Cache:  true,
				Search: true,
			},
			Services: engine.Services{
				Cache:  "valkey",
				Search: "opensearch",
			},
		},
	}

	engine.NormalizeConfig(&config, "")

	if config.Stack.CacheVersion != "9.0" {
		t.Fatalf("Expected CacheVersion 9.0, got %s", config.Stack.CacheVersion)
	}

	if config.Stack.SearchVersion != "3.0" {
		t.Fatalf("Expected SearchVersion 3.0, got %s", config.Stack.SearchVersion)
	}
}

func TestPrepareConfigForWriteOmitsCurrentRuntimeUserIDs(t *testing.T) {
	uid := os.Getuid()
	gid := os.Getgid()
	if uid < 0 || gid < 0 {
		t.Skip("uid/gid not available on this platform")
	}

	config := engine.Config{
		Stack: engine.Stack{
			UserID:  uid,
			GroupID: gid,
		},
	}

	writable := engine.PrepareConfigForWrite(config)
	if writable.Stack.UserID != 0 {
		t.Fatalf("expected UserID to be omitted, got %d", writable.Stack.UserID)
	}
	if writable.Stack.GroupID != 0 {
		t.Fatalf("expected GroupID to be omitted, got %d", writable.Stack.GroupID)
	}

	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal writable config: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "user_id:") {
		t.Fatalf("expected serialized config to omit user_id, got:\n%s", content)
	}
	if strings.Contains(content, "group_id:") {
		t.Fatalf("expected serialized config to omit group_id, got:\n%s", content)
	}
}

func TestPrepareConfigForWriteKeepsCustomUserIDs(t *testing.T) {
	uid := os.Getuid()
	gid := os.Getgid()

	customUID := uid + 111
	customGID := gid + 222
	if uid < 0 {
		customUID = 2001
	}
	if gid < 0 {
		customGID = 2002
	}

	config := engine.Config{
		Stack: engine.Stack{
			UserID:  customUID,
			GroupID: customGID,
		},
	}

	writable := engine.PrepareConfigForWrite(config)
	if writable.Stack.UserID != customUID {
		t.Fatalf("expected custom UserID %d to persist, got %d", customUID, writable.Stack.UserID)
	}
	if writable.Stack.GroupID != customGID {
		t.Fatalf("expected custom GroupID %d to persist, got %d", customGID, writable.Stack.GroupID)
	}
}

func TestPrepareConfigForWriteKeepsRuntimeProfileDefaults(t *testing.T) {
	profileResult, err := engine.ResolveRuntimeProfile("magento2", "2.4.7-p3")
	if err != nil {
		t.Fatalf("resolve runtime profile: %v", err)
	}
	profile := profileResult.Profile

	config := engine.Config{
		ProjectName:      "demo",
		Framework:        "magento2",
		FrameworkVersion: "2.4.7-p3",
		Domain:           "demo.test",
		Stack: engine.Stack{
			PHPVersion:    profile.PHPVersion,
			NodeVersion:   profile.NodeVersion,
			DBType:        profile.DBType,
			DBVersion:     profile.DBVersion,
			WebRoot:       profile.WebRoot,
			CacheVersion:  profile.CacheVersion,
			SearchVersion: profile.SearchVersion,
			QueueVersion:  profile.QueueVersion,
			XdebugSession: profile.XdebugSession,
			WebServer:     profile.WebServer,
			Services: engine.Services{
				WebServer: profile.WebServer,
				Search:    profile.Search,
				Cache:     profile.Cache,
				Queue:     profile.Queue,
				DB:        profile.DBType,
			},
			Features: engine.Features{
				Xdebug: true,
				Cache:  true,
				Search: true,
				Queue:  true,
			},
		},
		Remotes: map[string]engine.RemoteConfig{},
		Hooks:   map[string][]engine.HookStep{},
	}

	writable := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal writable config: %v", err)
	}
	content := string(data)

	for _, key := range []string{
		"php_version:",
		"node_version:",
		"db_version:",
		"web_root:",
		"cache_version:",
		"search_version:",
		"queue_version:",
	} {
		if !strings.Contains(content, key) {
			t.Fatalf("expected explicit %q to be preserved in serialized config, got:\n%s", key, content)
		}
	}
	for _, key := range []string{
		"search:",
		"cache:",
		"queue:",
		"db:",
		"web_server:",
	} {
		if !strings.Contains(content, key) {
			t.Fatalf("expected service %q to be present in serialized config, got:\n%s", key, content)
		}
	}
	if strings.Contains(content, "hooks:") {
		t.Fatalf("expected empty hooks to be omitted, got:\n%s", content)
	}
	if strings.Contains(content, "xdebug_session:") {
		t.Fatalf("expected default xdebug_session to be omitted, got:\n%s", content)
	}
	if strings.Contains(content, "\n  web_server:") {
		t.Fatalf("expected duplicated top-level web_server to be omitted, got:\n%s", content)
	}
	if !strings.Contains(content, "    web_server:") {
		t.Fatalf("expected services.web_server to remain serialized, got:\n%s", content)
	}
	if !strings.Contains(content, "xdebug: true") {
		t.Fatalf("expected xdebug feature to be serialized, got:\n%s", content)
	}
	if strings.Contains(content, "cache: true") {
		t.Fatalf("expected feature cache to be OMITTED from YAML, got:\n%s", content)
	}
	if strings.Contains(content, "search: true") {
		t.Fatalf("expected feature search to be OMITTED from YAML, got:\n%s", content)
	}
}

func TestPrepareConfigForWriteKeepsNonDefaultRuntimeOverrides(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			PHPVersion: "8.1",
			Features: engine.Features{
				Cache: true,
			},
			Services: engine.Services{
				Cache: "valkey",
			},
		},
	}

	writable := engine.PrepareConfigForWrite(config)
	if writable.Stack.PHPVersion != "8.1" {
		t.Fatalf("expected non-default php_version to remain, got %q", writable.Stack.PHPVersion)
	}
	if writable.Stack.Services.Cache != "valkey" {
		t.Fatalf("expected non-default cache service to remain, got %q", writable.Stack.Services.Cache)
	}
}

func TestPrepareConfigForWriteKeepsNonDefaultXdebugSession(t *testing.T) {
	config := engine.Config{
		Framework:        "magento2",
		FrameworkVersion: "2.4.7-p3",
		Stack: engine.Stack{
			XdebugSession: "VSCODE",
		},
	}

	writable := engine.PrepareConfigForWrite(config)
	if writable.Stack.XdebugSession != "VSCODE" {
		t.Fatalf("expected non-default xdebug_session to remain, got %q", writable.Stack.XdebugSession)
	}

	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal writable config: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "xdebug_session: VSCODE") {
		t.Fatalf("expected serialized non-default xdebug_session, got:\n%s", content)
	}
}

func TestPrepareConfigForWritePrunesDefaultRemoteAuthAndPaths(t *testing.T) {
	config := engine.Config{
		ProjectName: "demo",
		Framework:   "magento2",
		Domain:      "demo.test",
		Remotes: map[string]engine.RemoteConfig{
			"dev": {
				Host: "example.com",
				User: "deploy",
				Port: 22,
				Path: "/var/www/html",
				Capabilities: &engine.RemoteCapabilities{
					Files: engine.BoolPtr(true),
					Media: engine.BoolPtr(true),
					DB:    engine.BoolPtr(true),
				},
				Auth: engine.RemoteAuth{
					Method: engine.RemoteAuthMethodKeychain,
				},
				Paths: engine.RemotePaths{
					Media: "",
				},
			},
		},
	}

	writable := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal writable config: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "method: keychain") {
		t.Fatalf("expected default remote auth method to be omitted, got:\n%s", content)
	}
	if strings.Contains(content, "key_path:") {
		t.Fatalf("expected empty auth.key_path to be omitted, got:\n%s", content)
	}
	if strings.Contains(content, "strict_host_key:") {
		t.Fatalf("expected default strict_host_key to be omitted, got:\n%s", content)
	}
	if strings.Contains(content, "known_hosts_file:") {
		t.Fatalf("expected empty known_hosts_file to be omitted, got:\n%s", content)
	}
	if strings.Contains(content, "paths:") {
		t.Fatalf("expected empty paths block to be omitted, got:\n%s", content)
	}
}

func TestPrepareConfigForWriteKeepsNonDefaultRemoteAuthAndPaths(t *testing.T) {
	config := engine.Config{
		ProjectName: "demo",
		Framework:   "magento2",
		Domain:      "demo.test",
		Remotes: map[string]engine.RemoteConfig{
			"staging": {
				Host: "staging.example.com",
				User: "deploy",
				Port: 22,
				Path: "/srv/www/staging",
				Capabilities: &engine.RemoteCapabilities{
					Files: engine.BoolPtr(true),
					Media: engine.BoolPtr(true),
					DB:    engine.BoolPtr(true),
				},
				Auth: engine.RemoteAuth{
					Method:         engine.RemoteAuthMethodKeyfile,
					KeyPath:        "~/.ssh/id_ed25519",
					StrictHostKey:  true,
					KnownHostsFile: "~/.ssh/known_hosts",
				},
				Paths: engine.RemotePaths{
					Media: "/srv/media",
				},
			},
		},
	}

	writable := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal writable config: %v", err)
	}
	content := string(data)
	for _, key := range []string{
		"method: keyfile",
		"key_path: ~/.ssh/id_ed25519",
		"strict_host_key: true",
		"known_hosts_file: ~/.ssh/known_hosts",
		"paths:",
		"media: /srv/media",
	} {
		if !strings.Contains(content, key) {
			t.Fatalf("expected %q to be serialized, got:\n%s", key, content)
		}
	}
}

func TestPrepareConfigForWriteOmitsEmptyFrameworkVersion(t *testing.T) {
	config := engine.Config{
		ProjectName:      "demo",
		Framework:        "magento2",
		FrameworkVersion: "",
		Domain:           "demo.test",
		Stack: engine.Stack{
			PHPVersion:  "8.3",
			NodeVersion: "24",
			DBType:      "mariadb",
			DBVersion:   "10.6",
			WebRoot:     "/pub",
			Services: engine.Services{
				WebServer: "nginx",
				Search:    "opensearch",
				Cache:     "redis",
				Queue:     "rabbitmq",
			},
			Features: engine.Features{
				Xdebug: true,
			},
		},
		Remotes: map[string]engine.RemoteConfig{},
		Hooks:   map[string][]engine.HookStep{},
	}

	writable := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal writable config: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "framework_version:") {
		t.Fatalf("expected empty framework_version to be omitted, got:\n%s", content)
	}
}

func TestPrepareConfigForWriteKeepsNonEmptyFrameworkVersion(t *testing.T) {
	config := engine.Config{
		ProjectName:      "demo",
		Framework:        "magento2",
		FrameworkVersion: "2.4.7-p3",
		Domain:           "demo.test",
		Stack: engine.Stack{
			PHPVersion:  "8.3",
			NodeVersion: "24",
			DBType:      "mariadb",
			DBVersion:   "10.6",
			WebRoot:     "/pub",
			Services: engine.Services{
				WebServer: "nginx",
				Search:    "opensearch",
				Cache:     "redis",
				Queue:     "rabbitmq",
			},
			Features: engine.Features{
				Xdebug: true,
			},
		},
		Remotes: map[string]engine.RemoteConfig{},
		Hooks:   map[string][]engine.HookStep{},
	}

	writable := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal writable config: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "framework_version: 2.4.7-p3") {
		t.Fatalf("expected non-empty framework_version to be serialized, got:\n%s", content)
	}
}

func TestPrepareConfigForWritePreservesNoneServicesRoundTrip(t *testing.T) {
	// Simulate user config with queue: none (no queue service desired)
	config := engine.Config{
		ProjectName:      "demo",
		Framework:        "magento2",
		FrameworkVersion: "2.4.7-p3",
		Domain:           "demo.test",
		Stack: engine.Stack{
			PHPVersion:    "8.3",
			NodeVersion:   "20",
			DBType:        "mariadb",
			DBVersion:     "10.6",
			WebRoot:       "/pub",
			CacheVersion:  "7.2",
			SearchVersion: "2.12",
			Services: engine.Services{
				WebServer: "apache",
				DB:        "mariadb",
				Cache:     "redis",
				Search:    "elasticsearch",
				Queue:     "none",
			},
			Features: engine.Features{
				Xdebug: true,
			},
		},
	}

	// Step 1: PrepareConfigForWrite should strip queue: none
	writable := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal writable config: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "queue_version:") {
		t.Fatalf("expected queue_version to be omitted when queue is none, got:\n%s", content)
	}
	if strings.Contains(content, "queue: none") {
		t.Fatalf("expected queue: none to be stripped from YAML, got:\n%s", content)
	}

	// Step 2: Simulate re-loading: unmarshal and normalize
	var reloaded engine.Config
	if err := yaml.Unmarshal(data, &reloaded); err != nil {
		t.Fatalf("unmarshal roundtrip config: %v", err)
	}
	engine.NormalizeConfig(&reloaded, "")

	if reloaded.Stack.Services.Queue != "none" {
		t.Fatalf("expected queue to remain 'none' (absent in YAML) after roundtrip, got %q", reloaded.Stack.Services.Queue)
	}
	if reloaded.Stack.QueueVersion != "" {
		t.Fatalf("expected queue_version to remain empty after roundtrip, got %q", reloaded.Stack.QueueVersion)
	}
	if reloaded.Stack.Features.Queue {
		t.Fatalf("expected queue feature to remain false after roundtrip")
	}
}

func TestNormalizeConfigAbsentServiceTreatedAsNone(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			Services: engine.Services{
				DB: "mysql", // Explicitly provided
			},
		},
	}

	engine.NormalizeConfig(&config, "")

	if config.Stack.Services.DB != "mysql" {
		t.Fatalf("expected db: mysql to be preserved, got %q", config.Stack.Services.DB)
	}
	if config.Stack.Services.Cache != "none" {
		t.Fatalf("expected cache: none (absent in input), got %q", config.Stack.Services.Cache)
	}
	if config.Stack.Services.Search != "none" {
		t.Fatalf("expected search: none (absent in input), got %q", config.Stack.Services.Search)
	}
	if config.Stack.Services.Queue != "none" {
		t.Fatalf("expected queue: none (absent in input), got %q", config.Stack.Services.Queue)
	}
}

func TestPrepareConfigForWriteStripsNoneServices(t *testing.T) {
	config := engine.Config{
		Stack: engine.Stack{
			Services: engine.Services{
				DB:     "none",
				Cache:  "none",
				Search: "none",
				Queue:  "none",
			},
		},
	}

	writable := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	content := string(data)
	for _, key := range []string{"db:", "cache:", "search:", "queue:"} {
		if strings.Contains(content, key) {
			t.Fatalf("expected %q to be stripped when set to 'none', got:\n%s", key, content)
		}
	}
}

// TestNormalizeConfigComposerVersionMagento1Default verifies that magento1 gets
// "2.2" as the default composer version (legacy PHP runtime).
func TestNormalizeConfigComposerVersionMagento1Default(t *testing.T) {
	config := engine.Config{
		Framework: "magento1",
	}
	engine.NormalizeConfig(&config, "")

	if config.Stack.ComposerVersion != "2.2" {
		t.Fatalf("Expected ComposerVersion 2.2 for magento1, got %q", config.Stack.ComposerVersion)
	}
}

// TestNormalizeConfigComposerVersionMagento2Default verifies that magento2 gets
// "latest" as the default composer version (modern PHP runtime).
func TestNormalizeConfigComposerVersionMagento2Default(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
	}
	engine.NormalizeConfig(&config, "")

	if config.Stack.ComposerVersion != "latest" {
		t.Fatalf("Expected ComposerVersion latest for magento2, got %q", config.Stack.ComposerVersion)
	}
}

// TestNormalizeConfigComposerVersionSafetyOverride verifies that when PHP < 7.2.5
// and composer_version is "latest", the safety check forces it to "2.2".
func TestNormalizeConfigComposerVersionSafetyOverride(t *testing.T) {
	config := engine.Config{
		Framework: "custom",
		Stack: engine.Stack{
			PHPVersion:      "7.1",
			ComposerVersion: "latest", // user explicitly set latest
		},
	}
	engine.NormalizeConfig(&config, "")

	if config.Stack.ComposerVersion != "2.2" {
		t.Fatalf("Expected ComposerVersion 2.2 when PHP=7.1 (safety override), got %q", config.Stack.ComposerVersion)
	}
}

// TestNormalizeConfigComposerVersionExplicitPreserved verifies that an explicit
// user-defined composer_version is not overwritten by framework defaults.
func TestNormalizeConfigComposerVersionExplicitPreserved(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			ComposerVersion: "2.5.1",
		},
	}
	engine.NormalizeConfig(&config, "")

	if config.Stack.ComposerVersion != "2.5.1" {
		t.Fatalf("Expected ComposerVersion 2.5.1 to be preserved, got %q", config.Stack.ComposerVersion)
	}
}

// TestPrepareConfigForWriteStripsDefaultComposer verifies that the default
// composer_version ("latest" for magento2) is stripped from the YAML output.
func TestPrepareConfigForWriteStripsDefaultComposer(t *testing.T) {
	config := engine.Config{
		ProjectName: "test-project",
		Framework:   "magento2",
		Domain:      "test.test",
		Stack: engine.Stack{
			ComposerVersion: "latest", // same as magento2 default
		},
	}

	writable := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	if strings.Contains(string(data), "composer_version") {
		t.Fatalf("Expected composer_version to be stripped (matches default), got:\n%s", string(data))
	}
}

// TestPrepareConfigForWriteKeepsCustomComposer verifies that a non-default
// composer_version is preserved in the YAML output.
func TestPrepareConfigForWriteKeepsCustomComposer(t *testing.T) {
	config := engine.Config{
		ProjectName: "test-project",
		Framework:   "magento2",
		Domain:      "test.test",
		Stack: engine.Stack{
			ComposerVersion: "2.2", // non-default for magento2
		},
	}

	writable := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writable)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	if !strings.Contains(string(data), "composer_version") || !strings.Contains(string(data), "2.2") {
		t.Fatalf("Expected composer_version: 2.2 in YAML output, got:\n%s", string(data))
	}
}
