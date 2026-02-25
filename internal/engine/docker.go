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
	"github.com/docker/docker/client"
)

func CheckDockerStatus() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer func() {
		_ = cli.Close()
	}()

	_, err = cli.Ping(ctx)
	return err
}

func CheckPort(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// CheckPortForGovardProxy returns true when the port is available OR already
// bound by the Govard proxy container (proxy-caddy), which is an expected state.
func CheckPortForGovardProxy(port string) bool {
	if CheckPort(port) {
		return true
	}
	return isPortBoundByGovardProxy(port)
}

func isPortBoundByGovardProxy(port string) bool {
	targetPort, err := strconv.Atoi(strings.TrimSpace(port))
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false
	}
	defer func() {
		_ = cli.Close()
	}()

	containers, err := cli.ContainerList(ctx, container.ListOptions{})
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

func CheckDockerComposePlugin() error {
	command := exec.Command("docker", "compose", "version")
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

func GetRunningProjectNames() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cli.Close()
	}()

	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}

	projectMap := make(map[string]bool)
	for _, c := range containers {
		for _, name := range c.Names {
			cleanName := strings.TrimPrefix(name, "/")
			// Standard Govard naming pattern: projectname-service-1
			parts := strings.Split(cleanName, "-")
			if len(parts) >= 3 {
				projectName := strings.Join(parts[:len(parts)-2], "-")
				// Basic filtering - only consider it a govard project if it has standard services
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
