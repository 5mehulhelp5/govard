package cmd

import (
	"fmt"
	"os"
	"strings"

	"govard/internal/engine"
	"govard/internal/engine/remote"
)

type SyncEndpoint struct {
	Name      string
	IsLocal   bool
	RemoteCfg engine.RemoteConfig
	RootPath  string
	MediaPath string
}

type ResolvedSyncEndpoints struct {
	Source      SyncEndpoint
	Destination SyncEndpoint
}

func resolveSyncEndpoints(config engine.Config, sourceName string, destinationName string) (ResolvedSyncEndpoints, error) {
	cwd, _ := os.Getwd()

	source, err := resolveSyncEndpoint(config, sourceName, cwd)
	if err != nil {
		return ResolvedSyncEndpoints{}, err
	}

	destination, err := resolveSyncEndpoint(config, destinationName, cwd)
	if err != nil {
		return ResolvedSyncEndpoints{}, err
	}

	return ResolvedSyncEndpoints{
		Source:      source,
		Destination: destination,
	}, nil
}

func resolveSyncEndpoint(config engine.Config, name string, cwd string) (SyncEndpoint, error) {
	if name == "local" {
		return SyncEndpoint{
			Name:      name,
			IsLocal:   true,
			RootPath:  cwd,
			MediaPath: engine.ResolveLocalMediaPath(config, cwd),
		}, nil
	}

	resolvedName, remoteCfg, err := ensureRemoteKnown(config, name)
	if err != nil {
		return SyncEndpoint{}, err
	}

	root, media := engine.ResolveRemotePathsForConfig(config.Framework, remoteCfg)
	if strings.TrimSpace(root) == "" {
		return SyncEndpoint{}, fmt.Errorf("the remote environment '%s' does not have a configured project path", name)
	}

	return SyncEndpoint{
		Name:      resolvedName,
		IsLocal:   false,
		RemoteCfg: remoteCfg,
		RootPath:  root,
		MediaPath: media,
	}, nil
}

func describeSyncEndpoint(endpoint SyncEndpoint) string {
	if endpoint.IsLocal {
		return fmt.Sprintf("%s (local project: %s)", endpoint.Name, endpoint.RootPath)
	}
	writePolicy := "Write-allowed"
	if blocked, reason := engine.RemoteWriteBlocked(endpoint.Name, endpoint.RemoteCfg); blocked {
		writePolicy = "Write-blocked (" + reason + ")"
	}
	return fmt.Sprintf(
		"%s (Target: %s, Path: %s, Policy: %s)",
		endpoint.Name,
		remote.RemoteTarget(endpoint.RemoteCfg),
		endpoint.RootPath,
		writePolicy,
	)
}

func ensureSyncCapability(endpoints ResolvedSyncEndpoints, capability string) error {
	if err := ensureEndpointCapability(endpoints.Source, "source", capability); err != nil {
		return err
	}
	if err := ensureEndpointCapability(endpoints.Destination, "destination", capability); err != nil {
		return err
	}
	return nil
}

func ensureEndpointCapability(endpoint SyncEndpoint, position string, capability string) error {
	if endpoint.IsLocal {
		return nil
	}
	if engine.RemoteCapabilityEnabled(endpoint.RemoteCfg, capability) {
		return nil
	}
	capDisplay := strings.ToUpper(capability[0:1]) + capability[1:]
	return fmt.Errorf(
		"the %s environment '%s' does not support %s synchronization (supported capabilities: %s)",
		position,
		endpoint.Name,
		capDisplay,
		strings.Join(engine.RemoteCapabilityList(endpoint.RemoteCfg), ", "),
	)
}

// ResolveSyncEndpointsForTest exposes resolveSyncEndpoints for tests in /tests.
func ResolveSyncEndpointsForTest(source, destination SyncEndpoint) ResolvedSyncEndpoints {
	return ResolvedSyncEndpoints{
		Source:      source,
		Destination: destination,
	}
}
