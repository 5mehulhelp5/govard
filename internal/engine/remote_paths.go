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

	mediaSubpath := ResolveFrameworkRemoteMediaSubpath(framework)
	if root == "" {
		return root, "/" + strings.TrimLeft(mediaSubpath, "/")
	}
	return root, strings.TrimRight(root, "/") + "/" + strings.TrimLeft(mediaSubpath, "/")
}
