package remote

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
	"govard/internal/engine"
)

var validSnapshotNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// ValidateSnapshotName ensures the snapshot name is safe and doesn't allow path traversal.
func ValidateSnapshotName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("snapshot name cannot be empty")
	}
	if strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\") || strings.Contains(trimmed, "..") {
		return fmt.Errorf("snapshot name must not contain path separators or '..'")
	}
	if !validSnapshotNamePattern.MatchString(trimmed) {
		return fmt.Errorf("snapshot name contains invalid characters: %q", trimmed)
	}
	return nil
}

// RemoteSnapshotRoot returns the remote snapshot root path for a given remote config.
func RemoteSnapshotRoot(remoteCfg engine.RemoteConfig) string {
	return strings.TrimRight(remoteCfg.Path, "/") + "/.govard/snapshots"
}

// RemoteSnapshotDir returns the full remote path for a named snapshot.
func RemoteSnapshotDir(remoteCfg engine.RemoteConfig, name string) string {
	return RemoteSnapshotRoot(remoteCfg) + "/" + name
}

// buildRemoteSnapshotCreateCommand builds the SSH command to create a snapshot on the remote.
// It creates the directory, dumps the DB to db.sql.gz, and tars media to media.tar.gz.
func buildRemoteSnapshotCreateCommand(
	remoteCfg engine.RemoteConfig,
	name string,
	framework string,
	dbDumpCommandStr string,
	remoteMediaPath string,
) string {
	snapshotDir := RemoteSnapshotDir(remoteCfg, name)
	quoted := QuoteRemotePath(snapshotDir)

	parts := []string{
		fmt.Sprintf("mkdir -p %s", quoted),
	}

	// DB dump
	if dbDumpCommandStr != "" {
		dbPath := snapshotDir + "/db.sql.gz"
		parts = append(parts,
			fmt.Sprintf("{ %s; } > %s", dbDumpCommandStr, shellQuoteRemote(dbPath)),
		)
	}

	// Media tar
	if strings.TrimSpace(remoteMediaPath) != "" {
		mediaTar := snapshotDir + "/media.tar.gz"
		parts = append(parts,
			fmt.Sprintf("if [ -d %s ]; then tar -czf %s -C %s .; fi",
				shellQuoteRemote(remoteMediaPath),
				shellQuoteRemote(mediaTar),
				shellQuoteRemote(remoteMediaPath),
			),
		)
	}

	// Write metadata
	metaPath := snapshotDir + "/metadata.yml"
	metaContent := fmt.Sprintf(
		"name: %s\\ncreated_at: $(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ)\\nframework: %s\\ndb: true\\nmedia: true",
		name, framework,
	)
	parts = append(parts,
		fmt.Sprintf("printf '%s\\n' > %s", metaContent, shellQuoteRemote(metaPath)),
	)

	return strings.Join(parts, " && ")
}

// buildRemoteSnapshotListCommand builds the SSH command to list snapshots on the remote.
func buildRemoteSnapshotListCommand(remoteCfg engine.RemoteConfig) string {
	root := RemoteSnapshotRoot(remoteCfg)
	// List snapshot directories and cat their metadata
	return fmt.Sprintf(
		"if [ -d %s ]; then for d in %s/*/; do [ -d \"$d\" ] && cat \"$d/metadata.yml\" 2>/dev/null && echo '---'; done; else echo 'EMPTY'; fi",
		shellQuoteRemote(root), shellQuoteRemote(root),
	)
}

// buildRemoteSnapshotDeleteCommand builds the SSH command to delete a snapshot on the remote.
func buildRemoteSnapshotDeleteCommand(remoteCfg engine.RemoteConfig, name string) string {
	snapshotDir := RemoteSnapshotDir(remoteCfg, name)
	return fmt.Sprintf("rm -rf %s", shellQuoteRemote(snapshotDir))
}

// buildRemoteSnapshotRestoreCommand builds the SSH command to restore a snapshot on the remote.
func buildRemoteSnapshotRestoreCommand(
	remoteCfg engine.RemoteConfig,
	name string,
	framework string,
	dbImportCommandStr string,
	remoteMediaPath string,
	dbOnly bool,
	mediaOnly bool,
) string {
	snapshotDir := RemoteSnapshotDir(remoteCfg, name)
	parts := []string{
		fmt.Sprintf("test -d %s", shellQuoteRemote(snapshotDir)),
	}

	// Restore DB
	if !mediaOnly && dbImportCommandStr != "" {
		dbPath := snapshotDir + "/db.sql.gz"
		parts = append(parts,
			fmt.Sprintf("if [ -f %s ]; then zcat %s | %s; fi",
				shellQuoteRemote(dbPath), shellQuoteRemote(dbPath), dbImportCommandStr,
			),
		)
	}

	// Restore media
	if !dbOnly && strings.TrimSpace(remoteMediaPath) != "" {
		mediaTar := snapshotDir + "/media.tar.gz"
		parts = append(parts,
			fmt.Sprintf("if [ -f %s ]; then mkdir -p %s && tar -xzf %s -C %s; fi",
				shellQuoteRemote(mediaTar),
				shellQuoteRemote(remoteMediaPath),
				shellQuoteRemote(mediaTar),
				shellQuoteRemote(remoteMediaPath),
			),
		)
	}

	return strings.Join(parts, " && ")
}

// BuildRemoteSnapshotPullCommand builds the rsync command to download a snapshot from remote to local.
func BuildRemoteSnapshotPullCommand(
	remoteName string,
	remoteCfg engine.RemoteConfig,
	name string,
	localSnapshotDir string,
) *exec.Cmd {
	remoteSnapshotDir := RemoteSnapshotDir(remoteCfg, name) + "/"
	source := fmt.Sprintf("%s:%s", RemoteTarget(remoteCfg), remoteSnapshotDir)
	return BuildRsyncCommand(remoteName, source, localSnapshotDir+"/", remoteCfg, false, true, false, nil, nil)
}

// BuildRemoteSnapshotPushCommand builds the rsync command to upload a local snapshot to remote.
func BuildRemoteSnapshotPushCommand(
	remoteName string,
	remoteCfg engine.RemoteConfig,
	name string,
	localSnapshotDir string,
) *exec.Cmd {
	remoteSnapshotDir := RemoteSnapshotDir(remoteCfg, name) + "/"
	destination := fmt.Sprintf("%s:%s", RemoteTarget(remoteCfg), remoteSnapshotDir)
	return BuildRsyncCommand(remoteName, localSnapshotDir+"/", destination, remoteCfg, false, true, false, nil, nil)
}

// ParseRemoteSnapshotList parses the output of the remote listing command.
func ParseRemoteSnapshotList(raw string) ([]engine.SnapshotMetadata, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "EMPTY" {
		return []engine.SnapshotMetadata{}, nil
	}

	documents := strings.Split(trimmed, "---")
	snapshots := make([]engine.SnapshotMetadata, 0, len(documents))
	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}
		var meta engine.SnapshotMetadata
		if err := yaml.Unmarshal([]byte(doc), &meta); err != nil {
			continue
		}
		if meta.Name == "" {
			continue
		}
		snapshots = append(snapshots, meta)
	}
	return snapshots, nil
}


// ForTest wrappers

func ValidateSnapshotNameForTest(name string) error {
	return ValidateSnapshotName(name)
}

func BuildRemoteSnapshotCreateCommandForTest(remoteName string, remoteCfg engine.RemoteConfig, name string, framework string) string {
	_, mediaPath := engine.ResolveRemotePathsForConfig(framework, remoteCfg)
	dbDump := "echo 'mock-dump'"
	return buildRemoteSnapshotCreateCommand(remoteCfg, name, framework, dbDump, mediaPath)
}

func BuildRemoteSnapshotListCommandForTest(remoteCfg engine.RemoteConfig) string {
	return buildRemoteSnapshotListCommand(remoteCfg)
}

func BuildRemoteSnapshotDeleteCommandForTest(remoteCfg engine.RemoteConfig, name string) string {
	return buildRemoteSnapshotDeleteCommand(remoteCfg, name)
}

func BuildRemoteSnapshotRestoreCommandForTest(remoteCfg engine.RemoteConfig, name string, framework string, dbOnly bool, mediaOnly bool) string {
	_, mediaPath := engine.ResolveRemotePathsForConfig(framework, remoteCfg)
	dbImport := "mysql -u magento magento"
	return buildRemoteSnapshotRestoreCommand(remoteCfg, name, framework, dbImport, mediaPath, dbOnly, mediaOnly)
}

func ParseRemoteSnapshotListForTest(raw string) ([]engine.SnapshotMetadata, error) {
	return ParseRemoteSnapshotList(raw)
}
