package engine

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

type SnapshotMetadata struct {
	Name      string    `yaml:"name"`
	CreatedAt time.Time `yaml:"created_at"`
	Recipe    string    `yaml:"recipe"`
	Domain    string    `yaml:"domain"`
	DB        bool      `yaml:"db"`
	Media     bool      `yaml:"media"`
}

func SnapshotRoot(projectRoot string) string {
	return filepath.Join(projectRoot, ".govard", "snapshots")
}

func CreateSnapshot(projectRoot string, config Config, name string) (string, error) {
	if name == "" {
		name = time.Now().Format("20060102-150405")
	}

	root := SnapshotRoot(projectRoot)
	snapshotDir := filepath.Join(root, name)
	if _, err := os.Stat(snapshotDir); err == nil {
		return "", fmt.Errorf("snapshot %s already exists", name)
	}

	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return "", err
	}

	meta := SnapshotMetadata{
		Name:      name,
		CreatedAt: time.Now(),
		Recipe:    config.Recipe,
		Domain:    config.Domain,
		DB:        false,
		Media:     false,
	}

	containerName := fmt.Sprintf("%s-db-1", config.ProjectName)
	dbPath := filepath.Join(snapshotDir, "db.sql")
	dbFile, err := os.Create(dbPath)
	if err == nil {
		dumpCmd := exec.Command("docker", "exec", "-i", containerName, "mysqldump", "-u", "magento", "-pmagento", "magento")
		dumpCmd.Stdout = dbFile
		dumpCmd.Stderr = os.Stderr
		if err := dumpCmd.Run(); err == nil {
			meta.DB = true
		}
		_ = dbFile.Close()
	}

	mediaSource := ResolveLocalMediaPath(config, projectRoot)
	mediaDest := filepath.Join(snapshotDir, "media")
	if info, err := os.Stat(mediaSource); err == nil && info.IsDir() {
		if err := copyDir(mediaSource, mediaDest); err == nil {
			meta.Media = true
		}
	}

	payload, err := yaml.Marshal(meta)
	if err == nil {
		_ = os.WriteFile(filepath.Join(snapshotDir, "metadata.yml"), payload, 0644)
	}

	return snapshotDir, nil
}

func ListSnapshots(projectRoot string) ([]SnapshotMetadata, error) {
	root := SnapshotRoot(projectRoot)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []SnapshotMetadata{}, nil
		}
		return nil, err
	}

	snapshots := make([]SnapshotMetadata, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		metaPath := filepath.Join(root, name, "metadata.yml")
		payload, err := os.ReadFile(metaPath)
		if err != nil {
			snapshots = append(snapshots, SnapshotMetadata{Name: name})
			continue
		}

		var meta SnapshotMetadata
		if err := yaml.Unmarshal(payload, &meta); err != nil {
			snapshots = append(snapshots, SnapshotMetadata{Name: name})
			continue
		}
		if meta.Name == "" {
			meta.Name = name
		}
		snapshots = append(snapshots, meta)
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].CreatedAt.After(snapshots[j].CreatedAt)
	})

	return snapshots, nil
}

func RestoreSnapshot(projectRoot string, config Config, name string, dbOnly bool, mediaOnly bool) error {
	snapshotDir := filepath.Join(SnapshotRoot(projectRoot), name)
	if _, err := os.Stat(snapshotDir); err != nil {
		return fmt.Errorf("snapshot %s not found", name)
	}

	if !mediaOnly {
		dbPath := filepath.Join(snapshotDir, "db.sql")
		if file, err := os.Open(dbPath); err == nil {
			containerName := fmt.Sprintf("%s-db-1", config.ProjectName)
			importCmd := exec.Command("docker", "exec", "-i", containerName, "mysql", "-u", "magento", "-pmagento", "magento")
			importCmd.Stdin = file
			importCmd.Stdout = os.Stdout
			importCmd.Stderr = os.Stderr
			if err := importCmd.Run(); err != nil {
				_ = file.Close()
				return err
			}
			_ = file.Close()
		}
	}

	if !dbOnly {
		mediaSnapshot := filepath.Join(snapshotDir, "media")
		if info, err := os.Stat(mediaSnapshot); err == nil && info.IsDir() {
			targetMedia := ResolveLocalMediaPath(config, projectRoot)
			if err := os.RemoveAll(targetMedia); err != nil {
				return err
			}
			if err := copyDir(mediaSnapshot, targetMedia); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := copyFileWithMode(srcPath, dstPath, info.Mode()); err != nil {
			return err
		}
	}

	return nil
}

func copyFileWithMode(src string, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
