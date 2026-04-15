package engine

import "path/filepath"

func ResolveLocalMediaPath(cfg Config, root string) string {
	return filepath.Join(root, filepath.FromSlash(ResolveFrameworkLocalMediaSubpath(cfg.Framework)))
}
