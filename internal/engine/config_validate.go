package engine

import (
	"fmt"
	"strings"
)

var (
	validWebServers = map[string]struct{}{
		"nginx":  {},
		"apache": {},
		"hybrid": {},
	}
	validSearchServices = map[string]struct{}{
		"none":          {},
		"opensearch":    {},
		"elasticsearch": {},
	}
	validCacheServices = map[string]struct{}{
		"none":   {},
		"redis":  {},
		"valkey": {},
	}
	validQueueServices = map[string]struct{}{
		"none":     {},
		"rabbitmq": {},
	}
)

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.ProjectName) == "" {
		return fmt.Errorf("project_name is required")
	}
	if strings.TrimSpace(cfg.Domain) == "" {
		return fmt.Errorf("domain is required")
	}
	if strings.ContainsAny(cfg.Domain, " \t\r\n") {
		return fmt.Errorf("domain cannot contain whitespace")
	}
	if err := validateBlueprintRegistryConfig(cfg.BlueprintRegistry); err != nil {
		return err
	}

	if err := validateService("stack.services.web_server", cfg.Stack.Services.WebServer, validWebServers); err != nil {
		return err
	}
	if err := validateService("stack.services.search", cfg.Stack.Services.Search, validSearchServices); err != nil {
		return err
	}
	if err := validateService("stack.services.cache", cfg.Stack.Services.Cache, validCacheServices); err != nil {
		return err
	}
	if err := validateService("stack.services.queue", cfg.Stack.Services.Queue, validQueueServices); err != nil {
		return err
	}

	for name, remote := range cfg.Remotes {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("remote name cannot be empty")
		}
		if strings.TrimSpace(remote.Host) == "" {
			return fmt.Errorf("remote '%s' is missing host", name)
		}
		if strings.TrimSpace(remote.User) == "" {
			return fmt.Errorf("remote '%s' is missing user", name)
		}
		if strings.TrimSpace(remote.Path) == "" {
			return fmt.Errorf("remote '%s' is missing path", name)
		}
		if remote.Port < 1 || remote.Port > 65535 {
			return fmt.Errorf("remote '%s' has invalid port %d", name, remote.Port)
		}
		if !IsValidRemoteEnvironment(remote.Environment) {
			return fmt.Errorf("remote '%s' has unsupported environment '%s' (allowed: dev, staging, prod)", name, remote.Environment)
		}
		if !IsSupportedRemoteAuthMethod(remote.Auth.Method) {
			return fmt.Errorf("remote '%s' has unsupported auth method '%s' (allowed: ssh-agent, keychain, keyfile)", name, remote.Auth.Method)
		}
	}

	for event, steps := range cfg.Hooks {
		if _, ok := allowedHookEvents[event]; !ok {
			return fmt.Errorf("unsupported hook event: %s", event)
		}
		for idx, step := range steps {
			if strings.TrimSpace(step.Run) == "" {
				return fmt.Errorf("hook %s has empty run command at index %d", event, idx)
			}
		}
	}

	return nil
}

func validateService(field, value string, allowed map[string]struct{}) error {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if _, ok := allowed[value]; !ok {
		return fmt.Errorf("unsupported value for %s: %s", field, value)
	}
	return nil
}
