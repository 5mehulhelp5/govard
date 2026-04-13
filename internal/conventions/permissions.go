package conventions

import "os"

const (
	// DefaultFilePerm is the default permission for general files.
	DefaultFilePerm os.FileMode = 0644
	// DefaultDirPerm is the default permission for general directories.
	DefaultDirPerm os.FileMode = 0755
	// SecretFilePerm is the strict permission for sensitive files (e.g. config, secrets, logs).
	SecretFilePerm os.FileMode = 0600
	// SecretDirPerm is the strict permission for sensitive directories.
	SecretDirPerm os.FileMode = 0700
	// PublicDirPerm is the loose permission for public directories (e.g. legacy caches).
	PublicDirPerm os.FileMode = 0777
	// PublicFilePerm is the loose permission for public files (e.g. shared logs).
	PublicFilePerm os.FileMode = 0666
)
