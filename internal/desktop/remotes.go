package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"govard/internal/engine"
	engineremote "govard/internal/engine/remote"

	"gopkg.in/yaml.v3"
)

var defaultRunGovardCommandForDesktop = func(root string, args []string) (string, error) {
	binary, err := exec.LookPath("govard")
	if err != nil {
		return "", fmt.Errorf("govard CLI not found in PATH")
	}

	cmd := exec.Command(binary, args...)
	cmd.Dir = filepath.Clean(root)
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

var runGovardCommandForDesktop = defaultRunGovardCommandForDesktop

var defaultSyncSanitizeExcludePatterns = []string{
	".env",
	"*.pem",
	"*.key",
}

var defaultSyncLogExcludePatterns = []string{
	"var/log/**",
	"storage/logs/**",
}

type remoteSyncPlanOptions struct {
	Sanitize    bool
	ExcludeLogs bool
	Compress    bool
}

func defaultRemoteSyncPlanOptions() remoteSyncPlanOptions {
	return remoteSyncPlanOptions{
		Compress: true,
	}
}

func listProjectRemotes(project string) (RemoteSnapshot, error) {
	root, err := resolveProjectRootForRemotes(project)
	if err != nil {
		return RemoteSnapshot{}, err
	}
	return listProjectRemotesByPath(root)
}

func listProjectRemotesByPath(root string) (RemoteSnapshot, error) {
	cleanRoot := filepath.Clean(strings.TrimSpace(root))
	cfg, _, err := engine.LoadConfigFromDir(cleanRoot, true)
	if err != nil {
		return RemoteSnapshot{}, fmt.Errorf("load config for remotes: %w", err)
	}

	return RemoteSnapshot{
		Project:  strings.TrimSpace(cfg.ProjectName),
		Remotes:  buildRemoteEntries(cfg.Remotes),
		Warnings: []string{},
	}, nil
}

func upsertProjectRemote(project string, input RemoteUpsertInput) error {
	root, err := resolveProjectRootForRemotes(project)
	if err != nil {
		return err
	}
	return upsertProjectRemoteByPath(root, input)
}

func upsertProjectRemoteByPath(root string, input RemoteUpsertInput) error {
	cleanRoot := filepath.Clean(strings.TrimSpace(root))
	cfg, err := engine.LoadBaseConfigFromDir(cleanRoot, true)
	if err != nil {
		return fmt.Errorf("load writable config: %w", err)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return fmt.Errorf("remote name is required")
	}
	host := strings.TrimSpace(input.Host)
	if host == "" {
		return fmt.Errorf("remote host is required")
	}
	user := strings.TrimSpace(input.User)
	if user == "" {
		return fmt.Errorf("remote user is required")
	}
	path := strings.TrimSpace(input.Path)
	if path == "" {
		return fmt.Errorf("remote path is required")
	}

	port := input.Port
	if port <= 0 {
		port = 22
	}
	if port > 65535 {
		return fmt.Errorf("remote port must be between 1 and 65535")
	}

	environment := engine.NormalizeRemoteEnvironment(input.Environment)
	if !engine.IsValidRemoteEnvironment(environment) {
		return fmt.Errorf("unsupported remote environment '%s'", input.Environment)
	}

	capabilities, err := engine.ParseRemoteCapabilitiesCSV(input.Capabilities)
	if err != nil {
		return err
	}

	authMethod := engineremote.NormalizeAuthMethod(input.AuthMethod)
	if !engineremote.IsSupportedAuthMethod(authMethod) {
		return fmt.Errorf("unsupported auth method '%s'", input.AuthMethod)
	}

	if cfg.Remotes == nil {
		cfg.Remotes = map[string]engine.RemoteConfig{}
	}

	cfg.Remotes[name] = engine.RemoteConfig{
		Host:         host,
		User:         user,
		Path:         path,
		Port:         port,
		Environment:  environment,
		Protected:    input.Protected,
		Capabilities: capabilities,
		Auth: engine.RemoteAuth{
			Method: authMethod,
		},
	}

	engine.NormalizeConfig(&cfg)
	if err := engine.ValidateConfig(cfg); err != nil {
		return err
	}
	if err := writeBaseConfig(cleanRoot, cfg); err != nil {
		return err
	}
	return nil
}

func testRemote(project string, remoteName string) (string, error) {
	startedAt := time.Now()
	status := engine.OperationStatusFailure
	category := "runtime"
	message := ""
	defer func() {
		writeDesktopOperationEvent(
			"desktop.remote.test",
			status,
			project,
			remoteName,
			"",
			message,
			category,
			time.Since(startedAt),
		)
	}()

	remoteName = strings.TrimSpace(remoteName)
	if remoteName == "" {
		category = "validation"
		message = "remote name is required"
		return "", fmt.Errorf("%s", message)
	}

	root, err := resolveProjectRootForRemotes(project)
	if err != nil {
		category = "validation"
		message = err.Error()
		return "", err
	}

	snapshot, err := listProjectRemotesByPath(root)
	if err != nil {
		message = err.Error()
		return "", err
	}
	if !hasRemoteName(snapshot.Remotes, remoteName) {
		category = "validation"
		message = fmt.Sprintf("unknown remote: %s", remoteName)
		return "", fmt.Errorf("%s", message)
	}

	output, err := runGovardCommandForDesktop(root, []string{"remote", "test", remoteName})
	if err != nil {
		message = err.Error()
		return "", fmt.Errorf("remote test failed: %w", err)
	}
	if output == "" {
		output = fmt.Sprintf("Remote '%s' test completed.", remoteName)
	}

	status = engine.OperationStatusSuccess
	category = ""
	message = "remote test completed"
	return output, nil
}

func runRemoteSyncPreset(project string, remoteName string, preset string) (string, error) {
	return runRemoteSyncPresetWithOptions(
		project,
		remoteName,
		preset,
		defaultRemoteSyncPlanOptions(),
	)
}

func runRemoteSyncPresetWithOptions(
	project string,
	remoteName string,
	preset string,
	options remoteSyncPlanOptions,
) (string, error) {
	startedAt := time.Now()
	status := engine.OperationStatusFailure
	category := "runtime"
	message := ""
	defer func() {
		writeDesktopOperationEvent(
			"desktop.remote.sync.plan",
			status,
			project,
			remoteName,
			"local",
			message,
			category,
			time.Since(startedAt),
		)
	}()

	remoteName = strings.TrimSpace(remoteName)
	if remoteName == "" {
		category = "validation"
		message = "remote name is required"
		return "", fmt.Errorf("%s", message)
	}

	root, err := resolveProjectRootForRemotes(project)
	if err != nil {
		category = "validation"
		message = err.Error()
		return "", err
	}

	snapshot, err := listProjectRemotesByPath(root)
	if err != nil {
		message = err.Error()
		return "", err
	}
	if !hasRemoteName(snapshot.Remotes, remoteName) {
		category = "validation"
		message = fmt.Sprintf("unknown remote: %s", remoteName)
		return "", fmt.Errorf("%s", message)
	}

	args, err := buildRemoteSyncPlanArgsWithOptions(remoteName, preset, options)
	if err != nil {
		category = "validation"
		message = err.Error()
		return "", err
	}

	output, err := runGovardCommandForDesktop(root, args)
	if err != nil {
		message = err.Error()
		return "", fmt.Errorf("sync plan failed: %w", err)
	}
	if output == "" {
		output = "Sync plan generated."
	}

	status = engine.OperationStatusPlan
	category = ""
	message = "sync plan generated"
	return output, nil
}

func resolveProjectRootForRemotes(project string) (string, error) {
	trimmedProject := strings.TrimSpace(project)
	if trimmedProject == "" {
		return "", fmt.Errorf("project is required")
	}

	if pathHasBaseConfig(trimmedProject) {
		return filepath.Clean(trimmedProject), nil
	}

	if entries, err := engine.ReadProjectRegistryEntries(); err == nil {
		for _, entry := range entries {
			if !projectRegistryMatches(entry, trimmedProject) {
				continue
			}
			if pathHasBaseConfig(entry.Path) {
				return filepath.Clean(entry.Path), nil
			}
		}
	}

	if info, err := loadProjectInfo(trimmedProject); err == nil {
		if configPath := strings.TrimSpace(info.configPath); configPath != "" {
			root := filepath.Dir(configPath)
			if pathHasBaseConfig(root) {
				return filepath.Clean(root), nil
			}
		}
	}

	return "", fmt.Errorf("unable to resolve project path for '%s'", trimmedProject)
}

func projectRegistryMatches(entry engine.ProjectRegistryEntry, project string) bool {
	if strings.TrimSpace(entry.ProjectName) == project {
		return true
	}
	if strings.TrimSpace(entry.Domain) == project {
		return true
	}
	return filepath.Base(strings.TrimSpace(entry.Path)) == project
}

func pathHasBaseConfig(root string) bool {
	configPath := filepath.Join(filepath.Clean(strings.TrimSpace(root)), engine.BaseConfigFile)
	info, err := os.Stat(configPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func buildRemoteEntries(remotes map[string]engine.RemoteConfig) []RemoteEntry {
	if len(remotes) == 0 {
		return []RemoteEntry{}
	}

	names := make([]string, 0, len(remotes))
	for name := range remotes {
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]RemoteEntry, 0, len(names))
	for _, name := range names {
		cfg := remotes[name]
		port := cfg.Port
		if port <= 0 {
			port = 22
		}

		capabilities := engine.RemoteCapabilityList(cfg)
		if len(capabilities) == 1 && capabilities[0] == "none" {
			capabilities = []string{}
		}

		environment := engine.NormalizeRemoteEnvironment(cfg.Environment)
		if environment == "" {
			environment = engine.RemoteEnvStaging
		}

		entry := RemoteEntry{
			Name:         strings.TrimSpace(name),
			Host:         strings.TrimSpace(cfg.Host),
			User:         strings.TrimSpace(cfg.User),
			Path:         strings.TrimSpace(cfg.Path),
			Port:         port,
			Environment:  environment,
			Protected:    cfg.Protected,
			AuthMethod:   engineremote.NormalizeAuthMethod(cfg.Auth.Method),
			Capabilities: append([]string{}, capabilities...),
		}
		if entry.AuthMethod == "" {
			entry.AuthMethod = engineremote.AuthMethodKeychain
		}
		entries = append(entries, entry)
	}
	return entries
}

func buildRemoteSyncPlanArgs(remoteName string, preset string) ([]string, error) {
	return buildRemoteSyncPlanArgsWithOptions(
		remoteName,
		preset,
		defaultRemoteSyncPlanOptions(),
	)
}

func buildRemoteSyncPlanArgsWithOptions(
	remoteName string,
	preset string,
	options remoteSyncPlanOptions,
) ([]string, error) {
	normalizedPreset, err := normalizeRemoteSyncPreset(preset)
	if err != nil {
		return nil, err
	}

	args := []string{"sync", "--source", strings.TrimSpace(remoteName), "--destination", "local"}
	switch normalizedPreset {
	case "files":
		args = append(args, "--file")
	case "media":
		args = append(args, "--media")
	case "db":
		args = append(args, "--db")
	case "full":
		args = append(args, "--full")
	default:
		return nil, fmt.Errorf("unsupported sync preset '%s'", normalizedPreset)
	}

	if options.Sanitize {
		for _, pattern := range defaultSyncSanitizeExcludePatterns {
			args = append(args, "--exclude", pattern)
		}
	}

	if options.ExcludeLogs {
		for _, pattern := range defaultSyncLogExcludePatterns {
			args = append(args, "--exclude", pattern)
		}
	}

	if !options.Compress {
		args = append(args, "--no-compress")
	}

	args = append(args, "--plan")
	return args, nil
}

func normalizeRemoteSyncPreset(preset string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "", "file", "files", "code", "source":
		return "files", nil
	case "media", "assets":
		return "media", nil
	case "db", "database":
		return "db", nil
	case "full", "all":
		return "full", nil
	default:
		return "", fmt.Errorf("unsupported sync preset '%s'", preset)
	}
}

func writeBaseConfig(root string, config engine.Config) error {
	writableConfig := engine.PrepareConfigForWrite(config)
	data, err := yaml.Marshal(&writableConfig)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, engine.BaseConfigFile), data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", engine.BaseConfigFile, err)
	}
	return nil
}

func hasRemoteName(remotes []RemoteEntry, name string) bool {
	for _, remote := range remotes {
		if remote.Name == name {
			return true
		}
	}
	return false
}

func writeDesktopOperationEvent(
	operation string,
	status engine.OperationStatus,
	project string,
	source string,
	destination string,
	message string,
	category string,
	duration time.Duration,
) {
	_ = engine.WriteOperationEvent(engine.OperationEvent{
		Operation:   operation,
		Status:      status,
		Project:     strings.TrimSpace(project),
		Source:      strings.TrimSpace(source),
		Destination: strings.TrimSpace(destination),
		Message:     strings.TrimSpace(message),
		Category:    strings.TrimSpace(category),
		DurationMS:  duration.Milliseconds(),
	})
}
