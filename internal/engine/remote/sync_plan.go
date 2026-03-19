package remote

import (
	"fmt"
	"strings"
)

type SyncOptions struct {
	Source      string
	Destination string
	Files       bool
	Media       bool
	DB          bool
	Delete      bool
	Resume      bool
	NoCompress  bool
	NoNoise     bool
	NoPII       bool
	Path        string
	Include     []string
	Exclude     []string
}

type SyncPlan struct {
	Source      string
	Destination string
	Command     string
}

func BuildSyncPlan(opts SyncOptions) SyncPlan {
	cmd := "rsync -a"
	if !opts.NoCompress {
		cmd += "z"
	}
	if opts.Delete {
		cmd += " --delete"
	}
	if opts.Resume {
		cmd += " --partial --append-verify"
	}
	for _, pattern := range opts.Include {
		trimmed := strings.TrimSpace(pattern)
		if trimmed == "" {
			continue
		}
		cmd += fmt.Sprintf(" --include %q", trimmed)
	}
	for _, pattern := range opts.Exclude {
		trimmed := strings.TrimSpace(pattern)
		if trimmed == "" {
			continue
		}
		cmd += fmt.Sprintf(" --exclude %q", trimmed)
	}
	return SyncPlan{Source: opts.Source, Destination: opts.Destination, Command: cmd}
}
