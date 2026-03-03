package desktop

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"govard/internal/engine"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const globalServicesComposeProjectName = "proxy"

var errDesktopGlobalServicesNotInitialized = errors.New("global services are not initialized")

type globalServiceSpec struct {
	ID             string
	Name           string
	ComposeService string
	ContainerName  string
	URLHost        string
}

var globalServiceSpecs = []globalServiceSpec{
	{
		ID:             "caddy",
		Name:           "Caddy Proxy",
		ComposeService: "caddy",
		ContainerName:  "govard-proxy-caddy",
	},
	{
		ID:             "mail",
		Name:           "Mailpit",
		ComposeService: "mail",
		ContainerName:  "govard-proxy-mail",
		URLHost:        "mail",
	},
	{
		ID:             "pma",
		Name:           "PHPMyAdmin",
		ComposeService: "pma",
		ContainerName:  "govard-proxy-pma",
		URLHost:        "pma",
	},
	{
		ID:             "portainer",
		Name:           "Portainer",
		ComposeService: "portainer",
		ContainerName:  "govard-proxy-portainer",
		URLHost:        "portainer",
	},
	{
		ID:             "dnsmasq",
		Name:           "DNSMasq",
		ComposeService: "dnsmasq",
		ContainerName:  "govard-proxy-dnsmasq",
	},
}

var defaultEnsureGlobalServicesForDesktop = func() error {
	if err := ensureGlobalComposeFileExists(); err == nil {
		return nil
	}

	if err := engine.EnsureGlobalProxy(); err != nil {
		return fmt.Errorf("ensure global proxy: %w", err)
	}
	return ensureGlobalComposeFileExists()
}

var ensureGlobalServicesForDesktop = defaultEnsureGlobalServicesForDesktop

var defaultRunGlobalServicesComposeForDesktop = func(args ...string) (string, error) {
	composeFile := globalServicesComposeFilePath()
	composeDir := globalServicesComposeDirPath()

	if err := ensureGlobalComposeFileExists(); err != nil {
		return "", err
	}

	dockerArgs := []string{
		"compose",
		"--project-directory",
		composeDir,
		"-p",
		globalServicesComposeProjectName,
		"-f",
		composeFile,
	}
	dockerArgs = append(dockerArgs, args...)

	command := exec.Command("docker", dockerArgs...)
	output, err := command.CombinedOutput()
	trimmed := strings.TrimSpace(string(output))
	if err != nil {
		if trimmed != "" {
			return "", fmt.Errorf("%w: %s", err, trimmed)
		}
		return "", err
	}
	return trimmed, nil
}

var runGlobalServicesComposeForDesktop = defaultRunGlobalServicesComposeForDesktop

func (s *GlobalServiceService) GetGlobalServices() (GlobalServicesSnapshot, error) {
	snapshot := GlobalServicesSnapshot{
		Total:    len(globalServiceSpecs),
		Services: make([]GlobalService, 0, len(globalServiceSpecs)),
	}

	containersByName := map[string]container.Summary{}
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		snapshot.Warnings = append(snapshot.Warnings, "Docker client error: "+err.Error())
	} else if containers, listErr := cli.ContainerList(ctx, container.ListOptions{All: true}); listErr != nil {
		snapshot.Warnings = append(snapshot.Warnings, "Docker unavailable: "+listErr.Error())
	} else {
		for _, c := range containers {
			for _, rawName := range c.Names {
				name := strings.TrimSpace(strings.TrimPrefix(rawName, "/"))
				if name == "" {
					continue
				}
				containersByName[name] = c
			}
		}
	}

	for _, spec := range globalServiceSpecs {
		service := GlobalService{
			ID:             spec.ID,
			Name:           spec.Name,
			ComposeService: spec.ComposeService,
			ContainerName:  spec.ContainerName,
			Status:         "missing",
			State:          "not-created",
			Health:         "unknown",
			StatusText:     "Container not created",
			Running:        false,
			Openable:       spec.URLHost != "",
		}
		if spec.URLHost != "" {
			service.URL = buildProxyURL(spec.URLHost)
		}

		if c, ok := containersByName[spec.ContainerName]; ok {
			status, health, running := deriveGlobalContainerStatus(c.State, c.Status)
			service.Status = status
			service.State = strings.TrimSpace(c.State)
			if service.State == "" {
				service.State = "unknown"
			}
			service.Health = health
			service.StatusText = strings.TrimSpace(c.Status)
			if service.StatusText == "" {
				service.StatusText = service.State
			}
			service.Running = running
		}

		if service.Running {
			snapshot.Active++
		}
		snapshot.Services = append(snapshot.Services, service)
	}

	snapshot.Summary = fmt.Sprintf("%d/%d global services running", snapshot.Active, snapshot.Total)
	return snapshot, nil
}

func (s *GlobalServiceService) StartGlobalServices() (string, error) {
	if err := ensureGlobalServicesForDesktop(); err != nil {
		return "", err
	}
	out, err := runGlobalServicesComposeForDesktop("up", "-d")
	if err != nil {
		return "", fmt.Errorf("start global services: %w", err)
	}
	return withCommandOutput("Global services started.", out), nil
}

func (s *GlobalServiceService) StopGlobalServices() (string, error) {
	if err := ensureGlobalComposeFileExists(); err != nil {
		return "", err
	}
	out, err := runGlobalServicesComposeForDesktop("stop")
	if err != nil {
		return "", fmt.Errorf("stop global services: %w", err)
	}
	return withCommandOutput("Global services stopped.", out), nil
}

func (s *GlobalServiceService) RestartGlobalServices() (string, error) {
	if err := ensureGlobalServicesForDesktop(); err != nil {
		return "", err
	}
	out, err := runGlobalServicesComposeForDesktop("restart")
	if err != nil {
		return "", fmt.Errorf("restart global services: %w", err)
	}
	return withCommandOutput("Global services restarted.", out), nil
}

func (s *GlobalServiceService) PullGlobalServices() (string, error) {
	if err := ensureGlobalServicesForDesktop(); err != nil {
		return "", err
	}
	out, err := runGlobalServicesComposeForDesktop("pull")
	if err != nil {
		return "", fmt.Errorf("pull global services: %w", err)
	}
	return withCommandOutput("Global services images pulled.", out), nil
}

func (s *GlobalServiceService) StartGlobalService(serviceID string) (string, error) {
	spec, err := resolveGlobalServiceSpec(serviceID)
	if err != nil {
		return "", err
	}
	if err := ensureGlobalServicesForDesktop(); err != nil {
		return "", err
	}
	out, err := runGlobalServicesComposeForDesktop("up", "-d", spec.ComposeService)
	if err != nil {
		return "", fmt.Errorf("start %s: %w", spec.Name, err)
	}
	return withCommandOutput(fmt.Sprintf("%s started.", spec.Name), out), nil
}

func (s *GlobalServiceService) StopGlobalService(serviceID string) (string, error) {
	spec, err := resolveGlobalServiceSpec(serviceID)
	if err != nil {
		return "", err
	}
	if err := ensureGlobalComposeFileExists(); err != nil {
		return "", err
	}
	out, err := runGlobalServicesComposeForDesktop("stop", spec.ComposeService)
	if err != nil {
		return "", fmt.Errorf("stop %s: %w", spec.Name, err)
	}
	return withCommandOutput(fmt.Sprintf("%s stopped.", spec.Name), out), nil
}

func (s *GlobalServiceService) RestartGlobalService(serviceID string) (string, error) {
	spec, err := resolveGlobalServiceSpec(serviceID)
	if err != nil {
		return "", err
	}
	if err := ensureGlobalServicesForDesktop(); err != nil {
		return "", err
	}
	out, err := runGlobalServicesComposeForDesktop("restart", spec.ComposeService)
	if err != nil {
		return "", fmt.Errorf("restart %s: %w", spec.Name, err)
	}
	return withCommandOutput(fmt.Sprintf("%s restarted.", spec.Name), out), nil
}

func (s *GlobalServiceService) OpenGlobalService(serviceID string) (string, error) {
	spec, err := resolveGlobalServiceSpec(serviceID)
	if err != nil {
		return "", err
	}
	if spec.URLHost == "" {
		return "", fmt.Errorf("%s has no web interface", spec.Name)
	}

	url := buildProxyURL(spec.URLHost)
	if err := openURLWithPreferences(s.ctx, url); err != nil {
		return "Open manually: " + url, nil
	}
	return "Opening " + url + "...", nil
}

func withCommandOutput(base string, commandOutput string) string {
	trimmed := strings.TrimSpace(commandOutput)
	if trimmed == "" {
		return base
	}
	return base + "\n" + trimmed
}

func resolveGlobalServiceSpec(serviceID string) (globalServiceSpec, error) {
	normalized := strings.ToLower(strings.TrimSpace(serviceID))
	for _, spec := range globalServiceSpecs {
		if spec.ID == normalized {
			return spec, nil
		}
	}
	return globalServiceSpec{}, fmt.Errorf("unknown global service: %s", serviceID)
}

func deriveGlobalContainerStatus(state string, statusText string) (string, string, bool) {
	normalizedState := strings.ToLower(strings.TrimSpace(state))
	status := "stopped"
	running := false

	switch normalizedState {
	case "running":
		status = "running"
		running = true
	case "restarting":
		status = "restarting"
	case "paused":
		status = "paused"
	case "created":
		status = "created"
	case "dead":
		status = "dead"
	case "exited":
		status = "stopped"
	default:
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(statusText)), "up ") {
			status = "running"
			running = true
		}
	}

	return status, deriveGlobalContainerHealth(statusText), running
}

func deriveGlobalContainerHealth(statusText string) string {
	normalized := strings.ToLower(strings.TrimSpace(statusText))
	switch {
	case strings.Contains(normalized, "(healthy)"):
		return "healthy"
	case strings.Contains(normalized, "(unhealthy)"):
		return "unhealthy"
	case strings.Contains(normalized, "health: starting"):
		return "starting"
	default:
		return "unknown"
	}
}

func globalServicesComposeDirPath() string {
	return filepath.Join(os.Getenv("HOME"), ".govard", "proxy")
}

func globalServicesComposeFilePath() string {
	return filepath.Join(globalServicesComposeDirPath(), "docker-compose.yml")
}

func ensureGlobalComposeFileExists() error {
	composeFile := globalServicesComposeFilePath()
	if _, err := os.Stat(composeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", errDesktopGlobalServicesNotInitialized, composeFile)
		}
		return fmt.Errorf("stat global compose file: %w", err)
	}
	return nil
}
