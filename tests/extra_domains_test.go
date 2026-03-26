package tests

import (
	"govard/internal/engine"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAllDomains(t *testing.T) {
	tests := []struct {
		name     string
		config   engine.Config
		expected []string
	}{
		{
			name:     "Empty config",
			config:   engine.Config{},
			expected: []string{},
		},
		{
			name: "Only primary domain",
			config: engine.Config{
				Domain: "myshop.test",
			},
			expected: []string{"myshop.test"},
		},
		{
			name: "Primary and multiple extras",
			config: engine.Config{
				Domain:       "myshop.test",
				ExtraDomains: []string{"brand-b.test", "wholesale.test"},
			},
			expected: []string{"myshop.test", "brand-b.test", "wholesale.test"},
		},
		{
			name: "Deduplicate extras",
			config: engine.Config{
				Domain:       "myshop.test",
				ExtraDomains: []string{"myshop.test", "brand-b.test"},
			},
			expected: []string{"myshop.test", "brand-b.test"},
		},
		{
			name: "Trim whitespace and ignore empty",
			config: engine.Config{
				Domain:       " myshop.test ",
				ExtraDomains: []string{"  brand-b.test  ", "", "   "},
			},
			expected: []string{"myshop.test", "brand-b.test"},
		},
		{
			name: "Include store domain hostnames",
			config: engine.Config{
				Domain: "myshop.test",
				StoreDomains: engine.StoreDomainMappings{
					"brand-b.test": {
						Code: "brand_b",
					},
					"wholesale.test": {
						Code: "wholesale",
						Type: "website",
					},
				},
			},
			expected: []string{"myshop.test", "brand-b.test", "wholesale.test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.config.AllDomains()
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("AllDomains() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

func TestConfigExtraDomainsYAML(t *testing.T) {
	yamlInput := `
domain: myshop.test
extra_domains:
  - brand-b.test
  - wholesale.test
`
	var config engine.Config
	err := yaml.Unmarshal([]byte(yamlInput), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	expectedExtras := []string{"brand-b.test", "wholesale.test"}
	if !reflect.DeepEqual(config.ExtraDomains, expectedExtras) {
		t.Errorf("ExtraDomains = %v, want %v", config.ExtraDomains, expectedExtras)
	}

	// Test Round-trip
	output, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	var roundTripped engine.Config
	err = yaml.Unmarshal(output, &roundTripped)
	if err != nil {
		t.Fatalf("Failed to unmarshal Marshaled YAML: %v", err)
	}

	if !reflect.DeepEqual(roundTripped.ExtraDomains, expectedExtras) {
		t.Errorf("Round-tripped ExtraDomains = %v, want %v", roundTripped.ExtraDomains, expectedExtras)
	}
}
