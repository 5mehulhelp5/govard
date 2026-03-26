package engine

import "path/filepath"

func ResolveLocalMediaPath(cfg Config, root string) string {
	switch cfg.Framework {
	case "magento2":
		return filepath.Join(root, "pub", "media")
	case "magento1", "openmage":
		return filepath.Join(root, "media")
	case "wordpress":
		return filepath.Join(root, "wp-content", "uploads")
	case "drupal":
		return filepath.Join(root, "sites", "default", "files")
	case "shopware":
		return filepath.Join(root, "public", "media")
	default:
		return filepath.Join(root, "public", "media")
	}
}
