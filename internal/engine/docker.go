package engine

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
)

func CheckDockerStatus(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cli, err := GetDockerClient()
	if err != nil {
		return err
	}

	_, err = cli.Ping(ctx)
	return err
}

func CheckPort(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	_ = ln.Close()
	return nil
}

func CheckPortForGovardProxy(ctx context.Context, port string) bool {
	maxRetries := 10
	// Optimization: List containers once to check both our proxy and others
	containers, _ := getRunningContainers(ctx)

	for i := 0; i < maxRetries; i++ {
		err := CheckPort(port)
		if err == nil {
			return true
		}

		if strings.Contains(err.Error(), "permission denied") {
			if isPortBoundByGovardProxyFromList(containers, port) {
				return true
			}
			if isPortBoundByOtherContainerFromList(containers, port) {
				return false
			}
			if i == maxRetries-1 {
				return true
			}
			select {
			case <-ctx.Done():
				return false
			case <-time.After(200 * time.Millisecond):
			}
			// Refresh list for next attempt
			containers, _ = getRunningContainers(ctx)
			continue
		}

		if isPortBoundByGovardProxyFromList(containers, port) {
			return true
		}

		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(500 * time.Millisecond):
			}
			// Refresh list for next attempt
			containers, _ = getRunningContainers(ctx)
		}
	}
	return false
}

func getRunningContainers(ctx context.Context) ([]container.Summary, error) {
	cli, err := GetDockerClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return cli.ContainerList(ctx, container.ListOptions{})
}

func isPortBoundByOtherContainerFromList(containers []container.Summary, port string) bool {
	targetPort, err := strconv.Atoi(strings.TrimSpace(port))
	if err != nil {
		return false
	}
	for _, c := range containers {
		if isGovardProxyContainer(c.Names) {
			continue
		}
		for _, published := range c.Ports {
			if published.Type == "tcp" && int(published.PublicPort) == targetPort {
				return true
			}
		}
	}
	return false
}

func isPortBoundByGovardProxyFromList(containers []container.Summary, port string) bool {
	targetPort, err := strconv.Atoi(strings.TrimSpace(port))
	if err != nil {
		return false
	}
	for _, c := range containers {
		if !isGovardProxyContainer(c.Names) {
			continue
		}
		for _, published := range c.Ports {
			if published.Type == "tcp" && int(published.PublicPort) == targetPort {
				return true
			}
		}
	}
	return false
}

func isGovardProxyContainer(names []string) bool {
	for _, name := range names {
		clean := strings.TrimPrefix(name, "/")
		if clean == "proxy-caddy-1" || clean == "govard-proxy-caddy" {
			return true
		}
	}
	return false
}

func IsContainerRunning(ctx context.Context, name string) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cli, err := GetDockerClient()
	if err != nil {
		return false
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return false
	}

	for _, c := range containers {
		for _, cname := range c.Names {
			if strings.TrimPrefix(cname, "/") == name {
				return true
			}
		}
	}

	return false
}

func CheckDockerComposePlugin(ctx context.Context) error {
	command := exec.CommandContext(ctx, "docker", "compose", "version")
	output, err := command.CombinedOutput()
	if err == nil {
		return nil
	}
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return fmt.Errorf("docker compose plugin is not available: %w", err)
	}
	return fmt.Errorf("docker compose plugin is not available: %w (%s)", err, trimmed)
}

func GetRunningProjectNames(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cli, err := GetDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	projectMap := make(map[string]bool)
	for _, c := range containers {
		// Prefer Docker Compose project label for high fidelity detection
		if projectName, ok := c.Labels["com.docker.compose.project"]; ok && projectName != "" {
			projectMap[projectName] = true
			continue
		}

		// Fallback to name parsing if labels are missing
		for _, name := range c.Names {
			cleanName := strings.TrimPrefix(name, "/")
			parts := strings.Split(cleanName, "-")
			if len(parts) >= 3 {
				projectName := strings.Join(parts[:len(parts)-2], "-")
				serviceName := parts[len(parts)-2]
				if serviceName == "web" || serviceName == "php" || serviceName == "db" {
					projectMap[projectName] = true
				}
			}
		}
	}

	projects := make([]string, 0, len(projectMap))
	for p := range projectMap {
		projects = append(projects, p)
	}
	return projects, nil
}
