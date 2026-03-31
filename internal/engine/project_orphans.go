package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type OrphanProject struct {
	Name        string `json:"Name"`
	Status      string `json:"Status"`
	ConfigFiles string `json:"ConfigFiles"`
}

// GetOrphanedComposeProjects returns a list of Docker Compose projects that are not in the Govard registry.
func GetOrphanedComposeProjects(ctx context.Context) ([]OrphanProject, error) {
	entries, err := ReadProjectRegistryEntries()
	if err != nil {
		return nil, err
	}

	registeredNames := make(map[string]bool)
	for _, entry := range entries {
		registeredNames[entry.ProjectName] = true
	}
	// Always ignore the Govard proxy
	registeredNames["proxy"] = true

	cmd := exec.CommandContext(ctx, "docker", "compose", "ls", "-a", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker compose ls: %w", err)
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		return []OrphanProject{}, nil
	}

	var allProjects []OrphanProject
	if err := json.Unmarshal(output, &allProjects); err != nil {
		return nil, fmt.Errorf("parse docker compose ls: %w", err)
	}

	orphans := make([]OrphanProject, 0)
	for _, p := range allProjects {
		if !registeredNames[p.Name] {
			orphans = append(orphans, p)
		}
	}

	return orphans, nil
}

// DeleteOrphanProject performs a cleanup of an unregistered Docker Compose project.
func DeleteOrphanProject(ctx context.Context, name string, stdout, stderr io.Writer) error {
	// We use -p to target the project by name.
	// Since it's not in the registry, we don't have a project directory,
	// so we run it from a neutral directory (like current or /tmp).
	cmd := exec.CommandContext(ctx, "docker", "compose", "-p", name, "down", "-v", "--remove-orphans")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
