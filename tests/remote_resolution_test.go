package tests

import (
	"govard/internal/cmd"
	"govard/internal/engine"
	"strings"
	"testing"
)

func TestResolveAutoRemote(t *testing.T) {
	tests := []struct {
		name      string
		config    engine.Config
		requested string
		want      string
		wantErr   string
	}{
		{
			name: "Explicitly requested exists",
			config: engine.Config{
				Remotes: map[string]engine.RemoteConfig{
					"prod": {Host: "prod.com"},
				},
			},
			requested: "prod",
			want:      "prod",
		},
		{
			name: "Explicitly requested exists as alias",
			config: engine.Config{
				Remotes: map[string]engine.RemoteConfig{
					"development": {Host: "dev.com"},
				},
			},
			requested: "dev",
			want:      "development",
		},
		{
			name: "Explicitly requested missing",
			config: engine.Config{
				Remotes: map[string]engine.RemoteConfig{
					"staging": {Host: "stg.com"},
				},
			},
			requested: "prod",
			wantErr:   "unknown remote: prod",
		},
		{
			name: "Auto-select staging (primary focus)",
			config: engine.Config{
				Remotes: map[string]engine.RemoteConfig{
					"staging":     {Host: "stg.com"},
					"development": {Host: "dev.com"},
				},
			},
			requested: "",
			want:      "staging",
		},
		{
			name: "Auto-select staging via alias",
			config: engine.Config{
				Remotes: map[string]engine.RemoteConfig{
					"stg":         {Host: "stg.com"},
					"development": {Host: "dev.com"},
				},
			},
			requested: "",
			want:      "stg",
		},
		{
			name: "Auto-select dev when staging missing",
			config: engine.Config{
				Remotes: map[string]engine.RemoteConfig{
					"dev": {Host: "dev.com"},
				},
			},
			requested: "",
			want:      "dev",
		},
		{
			name: "Auto-select development alias when staging missing",
			config: engine.Config{
				Remotes: map[string]engine.RemoteConfig{
					"development": {Host: "dev.com"},
				},
			},
			requested: "",
			want:      "development",
		},
		{
			name: "Neither exists",
			config: engine.Config{
				Remotes: map[string]engine.RemoteConfig{
					"production": {Host: "prod.com"},
				},
			},
			requested: "",
			wantErr:   "no remote environment found (tried staging, dev)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cmd.ResolveAutoRemote(tt.config, tt.requested)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("ResolveAutoRemote() expected error %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("ResolveAutoRemote() error = %v, wantErr %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ResolveAutoRemote() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveAutoRemote() got = %q, want %q", got, tt.want)
			}
		})
	}
}
