package engine

import "strings"

func ResolveRemotePaths(cfg Config, name string) (string, string) {
	remote, ok := cfg.Remotes[name]
	if !ok {
		return "", ""
	}
	return ResolveRemotePathsForConfig(cfg.Framework, remote)
}

func ResolveRemotePathsForConfig(framework string, remote RemoteConfig) (string, string) {
	root := remote.Path
	if strings.TrimSpace(remote.Paths.Media) != "" {
		return root, remote.Paths.Media
	}
	switch framework {
	case "magento2", "magento1":
		return root, root + "/pub/media"
	case "wordpress":
		return root, root + "/wp-content/uploads"
	case "drupal":
		return root, root + "/sites/default/files"
	case "shopware":
		return root, root + "/public/media"
	default:
		return root, root + "/public/media"
	}
}
