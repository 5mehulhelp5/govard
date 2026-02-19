package desktop

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"govard/internal/engine"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

func quickAction(ctx context.Context, action string, project string) (string, error) {
	switch action {
	case "open-mail":
		return openDestination(ctx, buildProxyURL("mail"), "Opening Mailpit...")
	case "open-pma":
		return openDestination(ctx, buildProxyURL("pma"), "Opening PHPMyAdmin...")
	case "toggle-xdebug":
		return toggleXdebug(project)
	case "check-health":
		return checkHealth()
	default:
		return "", fmt.Errorf("unknown action")
	}
}

func toggleXdebug(project string) (string, error) {
	info, err := selectProject(project)
	if err != nil {
		return "", err
	}
	if !info.configLoaded {
		return "", fmt.Errorf("govard.yml not found for %s", info.name)
	}
	config := info.config
	if config.ProjectName == "" {
		config.ProjectName = info.name
	}
	config.Stack.Features.Xdebug = !config.Stack.Features.Xdebug

	if info.configPath == "" {
		return "", fmt.Errorf("govard.yml path unavailable")
	}
	writableConfig := engine.PrepareConfigForWrite(config)

	data, err := yaml.Marshal(&writableConfig)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(info.configPath, data, 0644); err != nil {
		return "", err
	}

	if err := engine.RenderBlueprint(info.workingDir, config); err != nil {
		return "", err
	}

	composePath := engine.ComposeFilePath(info.workingDir, config.ProjectName)
	if err := runCompose(info.workingDir, config.ProjectName, composePath); err != nil {
		return "", err
	}

	state := "enabled"
	if !config.Stack.Features.Xdebug {
		state = "disabled"
	}
	return "Xdebug " + state + " for " + info.name, nil
}

func startEnvironment(project string) (string, error) {
	info, err := loadProjectInfo(project)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	args := filters.NewArgs(filters.Arg("label", "com.docker.compose.project="+info.name))
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil {
		return "", err
	}
	if len(containers) == 0 {
		return "", fmt.Errorf("no containers found")
	}
	for _, c := range containers {
		if c.State == "running" {
			continue
		}
		if err := cli.ContainerStart(ctx, c.ID, container.StartOptions{}); err != nil {
			return "", err
		}
	}
	return "Started environment " + info.name, nil
}

func stopEnvironment(project string) (string, error) {
	info, err := loadProjectInfo(project)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	args := filters.NewArgs(filters.Arg("label", "com.docker.compose.project="+info.name))
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil {
		return "", err
	}
	if len(containers) == 0 {
		return "", fmt.Errorf("no containers found")
	}
	timeout := 10
	for _, c := range containers {
		if c.State != "running" {
			continue
		}
		if err := cli.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
			return "", err
		}
	}
	return "Stopped environment " + info.name, nil
}

func checkHealth() (string, error) {
	dashboard, err := buildDashboard()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Docker OK. %d environments, %d services running.", len(dashboard.Environments), dashboard.RunningServices), nil
}

func selectProject(project string) (*projectInfo, error) {
	if project != "" {
		return loadProjectInfo(project)
	}

	dashboard, err := buildDashboard()
	if err != nil {
		return nil, err
	}
	if len(dashboard.Environments) == 0 {
		return nil, fmt.Errorf("no environments found")
	}

	selected := dashboard.Environments[0].Project
	for _, env := range dashboard.Environments {
		if env.Status == "running" {
			selected = env.Project
			break
		}
	}

	return loadProjectInfo(selected)
}

func runCompose(dir, project, composeFile string) error {
	args := []string{"compose", "--project-directory", filepath.Clean(dir)}
	if project != "" {
		args = append(args, "-p", project)
	}
	if composeFile != "" {
		args = append(args, "-f", composeFile)
	}
	args = append(args, "up", "-d")

	cmd := exec.Command("docker", args...)
	cmd.Dir = filepath.Clean(dir)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func openDestination(ctx context.Context, url string, message string) (string, error) {
	if err := openURLWithPreferences(ctx, url); err != nil {
		return message + " Open manually: " + url, nil
	}
	return message, nil
}

func openURLWithPreferences(ctx context.Context, url string) error {
	settings, err := getSettings()
	if err == nil && settings.PreferredBrowser != "" {
		cmd := exec.Command(settings.PreferredBrowser, url)
		if err := cmd.Start(); err == nil {
			return nil
		}
	}
	return openURL(ctx, url)
}

func buildProxyURL(host string) string {
	domain := resolveProxyDomain()
	return "https://" + host + "." + domain
}

func resolveProxyDomain() string {
	settings, err := getSettings()
	if err != nil {
		return "govard.test"
	}
	target := strings.TrimSpace(settings.ProxyTarget)
	if target == "" {
		return "govard.test"
	}
	if strings.Contains(target, ".") {
		return target
	}
	return target + ".test"
}

func openDocs(ctx context.Context, docPath string) error {
	root, err := FindRepoRoot()
	if err != nil {
		return err
	}
	fullPath := filepath.Join(root, docPath)
	_, err = os.Stat(fullPath)
	if err != nil {
		return err
	}
	return openURLWithPreferences(ctx, "file://"+fullPath)
}

func getLogs(project string, lines int) (string, error) {
	return getLogsForService(project, "", lines)
}

func getLogsForService(project string, service string, lines int) (string, error) {
	info, err := loadProjectInfo(project)
	if err != nil {
		return "", err
	}

	containerName := resolveLogContainer(info, service)

	args := []string{"logs", "--tail", fmt.Sprintf("%d", lines), containerName}
	cmd := exec.Command("docker", args...)
	cmd.Dir = filepath.Clean(info.workingDir)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), nil
	}

	containerName = resolveLogContainer(info, "php")
	args = []string{"logs", "--tail", fmt.Sprintf("%d", lines), containerName}
	cmd = exec.Command("docker", args...)
	cmd.Dir = filepath.Clean(info.workingDir)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", string(out))
	}
	return string(out), nil
}

func openShell(project string) error {
	return openShellForService(project, "", "", "bash")
}

func openShellForService(project string, service string, user string, shell string) error {
	info, err := loadProjectInfo(project)
	if err != nil {
		return err
	}

	containerName := resolveShellContainer(info, service)
	chosenShell := normalizeShell(shell)
	chosenUser := normalizeShellUser(info, service, user)

	if err := execShell(info, containerName, chosenUser, chosenShell); err == nil {
		return nil
	}

	if chosenShell != "sh" {
		if err := execShell(info, containerName, chosenUser, "sh"); err == nil {
			return nil
		}
	}

	if chosenUser != "" {
		if err := execShell(info, containerName, "", chosenShell); err == nil {
			return nil
		}
		if chosenShell != "sh" {
			return execShell(info, containerName, "", "sh")
		}
	}
	return fmt.Errorf("failed to open shell for %s", project)
}

func execShell(info *projectInfo, containerName string, user string, shell string) error {
	args := []string{"exec", "-it"}
	if user != "" {
		args = append(args, "-u", user)
	}
	args = append(args, containerName, shell)
	cmd := exec.Command("docker", args...)
	cmd.Dir = filepath.Clean(info.workingDir)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

func normalizeShell(shell string) string {
	if shell == "sh" {
		return "sh"
	}
	return "bash"
}

func normalizeShellUser(info *projectInfo, service string, user string) string {
	if user != "" {
		return user
	}
	if info != nil {
		if saved, err := getShellUser(info.name); err == nil && saved != "" {
			return saved
		}
	}
	if service == "php" || service == "" {
		return "www-data"
	}
	if service == "apache" {
		return "www-data"
	}
	if info != nil && info.configLoaded && info.config.Stack.Services.WebServer == "apache" && service == "web" {
		return "www-data"
	}
	return ""
}

func resolveLogContainer(info *projectInfo, service string) string {
	target := resolveServiceName(info, service, "web")
	return fmt.Sprintf("%s-%s-1", info.name, target)
}

func resolveShellContainer(info *projectInfo, service string) string {
	target := resolveServiceName(info, service, "php")
	return fmt.Sprintf("%s-%s-1", info.name, target)
}

func resolveServiceName(info *projectInfo, service string, fallback string) string {
	candidate := strings.TrimSpace(service)
	if candidate != "" {
		if info == nil || info.services[candidate] {
			return candidate
		}
	}
	if info != nil {
		if info.services[fallback] {
			return fallback
		}
		for name := range info.services {
			return name
		}
	}
	return fallback
}
