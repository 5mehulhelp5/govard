package tests

import (
	"testing"

	"govard/internal/engine"

	"gopkg.in/yaml.v3"
)

func TestRemoteConfigDefaults(t *testing.T) {
	yamlInput := `project_name: test
remotes:
  staging:
    host: example.com
    user: deploy
    path: /var/www/html
    auth:
      method: keychain
`
	var cfg engine.Config
	if err := yaml.Unmarshal([]byte(yamlInput), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	engine.NormalizeConfig(&cfg)

	remote := cfg.Remotes["staging"]
	if remote.Port != 22 {
		t.Fatalf("expected port 22, got %d", remote.Port)
	}
	if engine.NormalizeRemoteEnvironment("staging") != "staging" {
		t.Fatalf("expected environment staging, got %s", engine.NormalizeRemoteEnvironment("staging"))
	}
	if remote.Auth.Method != "keychain" {
		t.Fatalf("expected keychain, got %s", remote.Auth.Method)
	}
	if remote.Auth.StrictHostKey {
		t.Fatalf("expected strict host key default false")
	}
	if remote.Auth.KnownHostsFile != "" {
		t.Fatalf("expected empty known hosts file default, got %s", remote.Auth.KnownHostsFile)
	}
	if !remote.Capabilities.Files || !remote.Capabilities.Media || !remote.Capabilities.DB || !remote.Capabilities.Deploy {
		t.Fatalf("expected default capabilities to be enabled, got %+v", remote.Capabilities)
	}
	if remote.Path == "" {
		t.Fatalf("expected path set")
	}
}

func TestRemoteConfigDefaultsAuthMethodWhenOmitted(t *testing.T) {
	yamlInput := `project_name: test
domain: test.test
stack:
  services:
    web_server: nginx
    search: none
    cache: none
    queue: none
remotes:
  dev:
    host: example.com
    user: deploy
    path: /var/www/html
`
	var cfg engine.Config
	if err := yaml.Unmarshal([]byte(yamlInput), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	engine.NormalizeConfig(&cfg)

	remote := cfg.Remotes["dev"]
	if remote.Auth.Method != engine.RemoteAuthMethodKeychain {
		t.Fatalf("expected default auth method keychain, got %s", remote.Auth.Method)
	}
	if remote.Auth.StrictHostKey {
		t.Fatal("expected strict host key default false")
	}
	if remote.Auth.KnownHostsFile != "" {
		t.Fatalf("expected empty known_hosts_file by default, got %q", remote.Auth.KnownHostsFile)
	}
}

func TestRemoteConfigNormalizesAuthMethodAndKnownHosts(t *testing.T) {
	yamlInput := `project_name: test
domain: test.test
stack:
  services:
    web_server: nginx
    search: none
    cache: none
    queue: none
remotes:
  staging:
    host: example.com
    user: deploy
    path: /var/www/html
    auth:
      method: SSH_AGENT
      known_hosts_file: " ~/.ssh/known_hosts "
`
	var cfg engine.Config
	if err := yaml.Unmarshal([]byte(yamlInput), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	engine.NormalizeConfig(&cfg)

	remote := cfg.Remotes["staging"]
	if remote.Auth.Method != engine.RemoteAuthMethodSSHAgent {
		t.Fatalf("expected normalized ssh-agent method, got %s", remote.Auth.Method)
	}
	if remote.Auth.KnownHostsFile != "~/.ssh/known_hosts" {
		t.Fatalf("expected trimmed known hosts path, got %q", remote.Auth.KnownHostsFile)
	}
	if !remote.Auth.StrictHostKey {
		t.Fatal("expected strict host key enabled when known_hosts_file is set")
	}
}

func TestRemoteConfigRejectsUnsupportedAuthMethod(t *testing.T) {
	cfg := engine.Config{
		ProjectName: "test",
		Domain:      "test.test",
		Stack: engine.Stack{
			Services: engine.Services{
				WebServer: "nginx",
				Search:    "none",
				Cache:     "none",
				Queue:     "none",
			},
		},
		Remotes: map[string]engine.RemoteConfig{
			"staging": {
				Host: "example.com",
				User: "deploy",
				Port: 22,
				Path: "/var/www/html",
				Auth: engine.RemoteAuth{
					Method: "password",
				},
			},
		},
	}
	engine.NormalizeConfig(&cfg)
	if err := engine.ValidateConfig(cfg); err == nil {
		t.Fatal("expected unsupported auth method validation error")
	}
}
