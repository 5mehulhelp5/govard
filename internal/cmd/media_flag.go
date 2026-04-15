package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func resolveMediaModeFlagValue(cmd *cobra.Command, current string, args []string) string {
	if cmd == nil || !cmd.Flags().Changed("media") {
		return current
	}
	if normalized, ok := normalizeExplicitMediaModeArg(args); ok {
		return normalized
	}
	return current
}

func normalizeExplicitMediaModeArg(args []string) (string, bool) {
	if len(args) == 0 {
		return "", false
	}

	candidate := strings.ToLower(strings.TrimSpace(args[0]))
	switch candidate {
	case MediaSyncNone, MediaSyncMinimal, MediaSyncOptimized, MediaSyncAll, "catalog":
		return candidate, true
	default:
		return "", false
	}
}
