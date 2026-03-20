package tunnel

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"govard/internal/engine"
)

type BaseURLManager interface {
	Backup(projectRoot string, config engine.Config) error
	Update(projectRoot string, config engine.Config, tunnelURL string) error
	Revert(projectRoot string, config engine.Config) error
}

func NewBaseURLManager(framework string) BaseURLManager {
	switch strings.ToLower(framework) {
	case "magento2":
		return &Magento2Manager{}
	case "magento1":
		return &Magento1Manager{}
	case "laravel":
		return &LaravelManager{}
	case "wordpress":
		return &WordPressManager{}
	case "symfony":
		return &SymfonyManager{}
	default:
		return &NoopManager{}
	}
}

type NoopManager struct{}

func (m *NoopManager) Backup(projectRoot string, config engine.Config) error { return nil }
func (m *NoopManager) Update(projectRoot string, config engine.Config, tunnelURL string) error {
	return nil
}
func (m *NoopManager) Revert(projectRoot string, config engine.Config) error { return nil }

type Magento2Manager struct {
	Executor func(name string, args ...string) ([]byte, error)
}

func (m *Magento2Manager) Backup(projectRoot string, config engine.Config) error {
	return nil
}

func (m *Magento2Manager) Update(projectRoot string, config engine.Config, tunnelURL string) error {
	containerName := fmt.Sprintf("%s-db-1", config.ProjectName)
	// We update both secure and unsecure base URLs
	sql := fmt.Sprintf("UPDATE core_config_data SET value='%s/' WHERE path IN ('web/unsecure/base_url', 'web/secure/base_url')", tunnelURL)
	_, err := m.executeDockerMysql(containerName, sql)
	return err
}

func (m *Magento2Manager) Revert(projectRoot string, config engine.Config) error {
	containerName := fmt.Sprintf("%s-db-1", config.ProjectName)
	localURL := fmt.Sprintf("https://%s/", config.Domain)
	sql := fmt.Sprintf("UPDATE core_config_data SET value='%s' WHERE path IN ('web/unsecure/base_url', 'web/secure/base_url')", localURL)
	_, err := m.executeDockerMysql(containerName, sql)
	return err
}

func (m *Magento2Manager) executeDockerMysql(container string, sql string) ([]byte, error) {
	executor := m.Executor
	if executor == nil {
		executor = func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).CombinedOutput()
		}
	}
	return executor("docker", "exec", container, "mysql", "-umagento", "-pmagento", "magento", "-e", sql)
}

type LaravelManager struct {
	ReadFile  func(string) ([]byte, error)
	WriteFile func(string, []byte, os.FileMode) error
}

func (m *LaravelManager) Backup(projectRoot string, config engine.Config) error {
	return nil
}

func (m *LaravelManager) Update(projectRoot string, config engine.Config, tunnelURL string) error {
	return m.updateEnv(projectRoot, "APP_URL", tunnelURL)
}

func (m *LaravelManager) Revert(projectRoot string, config engine.Config) error {
	return m.updateEnv(projectRoot, "APP_URL", fmt.Sprintf("https://%s", config.Domain))
}

func (m *LaravelManager) updateEnv(projectRoot string, key string, value string) error {
	envPath := filepath.Join(projectRoot, ".env")
	read := m.ReadFile
	if read == nil {
		read = os.ReadFile
	}
	content, err := read(envPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = fmt.Sprintf("%s=%s", key, value)
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	write := m.WriteFile
	if write == nil {
		write = os.WriteFile
	}
	return write(envPath, []byte(strings.Join(lines, "\n")), 0644)
}

// Similar structs for other frameworks...

type Magento1Manager struct {
	Executor func(name string, args ...string) ([]byte, error)
}

func (m *Magento1Manager) Backup(projectRoot string, config engine.Config) error { return nil }
func (m *Magento1Manager) Update(projectRoot string, config engine.Config, tunnelURL string) error {
	containerName := fmt.Sprintf("%s-db-1", config.ProjectName)
	sql := fmt.Sprintf("UPDATE core_config_data SET value='%s/' WHERE path IN ('web/unsecure/base_url', 'web/secure/base_url')", tunnelURL)
	_, err := m.executeDockerMysql(containerName, sql)
	return err
}
func (m *Magento1Manager) Revert(projectRoot string, config engine.Config) error {
	containerName := fmt.Sprintf("%s-db-1", config.ProjectName)
	localURL := fmt.Sprintf("https://%s/", config.Domain)
	sql := fmt.Sprintf("UPDATE core_config_data SET value='%s' WHERE path IN ('web/unsecure/base_url', 'web/secure/base_url')", localURL)
	_, err := m.executeDockerMysql(containerName, sql)
	return err
}
func (m *Magento1Manager) executeDockerMysql(container string, sql string) ([]byte, error) {
	executor := m.Executor
	if executor == nil {
		executor = func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).CombinedOutput()
		}
	}
	return executor("docker", "exec", container, "mysql", "-umagento", "-pmagento", "magento", "-e", sql)
}

type WordPressManager struct {
	Executor func(name string, args ...string) ([]byte, error)
}

func (m *WordPressManager) Backup(projectRoot string, config engine.Config) error { return nil }
func (m *WordPressManager) Update(projectRoot string, config engine.Config, tunnelURL string) error {
	containerName := fmt.Sprintf("%s-php-fpm-1", config.ProjectName)
	// wp option update siteurl <url> && wp option update home <url>
	_, err := m.executeWP(containerName, "option", "update", "siteurl", tunnelURL)
	if err != nil {
		return err
	}
	_, err = m.executeWP(containerName, "option", "update", "home", tunnelURL)
	return err
}
func (m *WordPressManager) Revert(projectRoot string, config engine.Config) error {
	containerName := fmt.Sprintf("%s-php-fpm-1", config.ProjectName)
	localURL := fmt.Sprintf("https://%s", config.Domain)
	_, err := m.executeWP(containerName, "option", "update", "siteurl", localURL)
	if err != nil {
		return err
	}
	_, err = m.executeWP(containerName, "option", "update", "home", localURL)
	return err
}
func (m *WordPressManager) executeWP(container string, args ...string) ([]byte, error) {
	executor := m.Executor
	if executor == nil {
		executor = func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).CombinedOutput()
		}
	}
	fullArgs := append([]string{"exec", "-u", "www-data", container, "wp"}, args...)
	return executor("docker", fullArgs...)
}

type SymfonyManager struct {
	LaravelManager
}

func (m *SymfonyManager) Update(projectRoot string, config engine.Config, tunnelURL string) error {
	// Symfony often uses APP_URL or similar, but frequently SITE_URL
	// Let's support both or just APP_URL for consistency if it's there
	return m.updateEnv(projectRoot, "APP_URL", tunnelURL)
}
func (m *SymfonyManager) Revert(projectRoot string, config engine.Config) error {
	return m.updateEnv(projectRoot, "APP_URL", fmt.Sprintf("https://%s", config.Domain))
}
