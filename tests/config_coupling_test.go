package tests

import (
	"govard/internal/engine"
	"testing"
)

func TestFeatureServiceCoupling(t *testing.T) {
	tests := []struct {
		name          string
		initialConfig engine.Config
		wantCache     string
		wantSearch    string
		wantQueue     string
		wantRedis     bool
		wantES        bool
		wantMQ        bool
	}{
		{
			name: "Missing features should NOT disable services (Service is Master)",
			initialConfig: engine.Config{
				Stack: engine.Stack{
					Services: engine.Services{
						Cache:  "redis",
						Search: "elasticsearch",
						Queue:  "rabbitmq",
					},
					Features: engine.Features{}, // All false initially
				},
			},
			wantCache:  "redis",
			wantSearch: "elasticsearch",
			wantQueue:  "rabbitmq",
			wantRedis:  true,
			wantES:     true,
			wantMQ:     true,
		},
		{
			name: "Explicit true features should preserve services",
			initialConfig: engine.Config{
				Stack: engine.Stack{
					Services: engine.Services{
						Cache:  "redis",
						Search: "elasticsearch",
						Queue:  "rabbitmq",
					},
					Features: engine.Features{
						Cache:  true,
						Search: true,
						Queue:  true,
					},
				},
			},
			wantCache:  "redis",
			wantSearch: "elasticsearch",
			wantQueue:  "rabbitmq",
			wantRedis:  true,
			wantES:     true,
			wantMQ:     true,
		},
		{
			name: "Setting service to none should disable feature",
			initialConfig: engine.Config{
				Stack: engine.Stack{
					Services: engine.Services{
						Cache:  "none",
						Search: "none",
						Queue:  "none",
					},
					Features: engine.Features{
						Cache:  true,
						Search: true,
						Queue:  true,
					},
				},
			},
			wantCache:  "none",
			wantSearch: "none",
			wantQueue:  "none",
			wantRedis:  false,
			wantES:     false,
			wantMQ:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.initialConfig
			engine.NormalizeConfig(&cfg, "")

			if cfg.Stack.Services.Cache != tt.wantCache {
				t.Errorf("%s: expected cache %q, got %q", tt.name, tt.wantCache, cfg.Stack.Services.Cache)
			}
			if cfg.Stack.Services.Search != tt.wantSearch {
				t.Errorf("%s: expected search %q, got %q", tt.name, tt.wantSearch, cfg.Stack.Services.Search)
			}
			if cfg.Stack.Services.Queue != tt.wantQueue {
				t.Errorf("%s: expected queue %q, got %q", tt.name, tt.wantQueue, cfg.Stack.Services.Queue)
			}
			if cfg.Stack.Features.Cache != tt.wantRedis {
				t.Errorf("%s: expected feature cache %v, got %v", tt.name, tt.wantRedis, cfg.Stack.Features.Cache)
			}
			if cfg.Stack.Features.Search != tt.wantES {
				t.Errorf("%s: expected feature search %v, got %v", tt.name, tt.wantES, cfg.Stack.Features.Search)
			}
			if cfg.Stack.Features.Queue != tt.wantMQ {
				t.Errorf("%s: expected feature queue %v, got %v", tt.name, tt.wantMQ, cfg.Stack.Features.Queue)
			}
		})
	}
}
