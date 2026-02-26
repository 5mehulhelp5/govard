package tests

import (
	"govard/internal/engine"
	"testing"
)

func TestAddDomainLogic(t *testing.T) {
	config := engine.Config{
		Domain: "main.test",
	}

	// Logic to add a domain
	newDomain := "extra.test"
	found := false
	for _, d := range config.ExtraDomains {
		if d == newDomain {
			found = true
			break
		}
	}
	if !found {
		config.ExtraDomains = append(config.ExtraDomains, newDomain)
	}

	if len(config.ExtraDomains) != 1 || config.ExtraDomains[0] != "extra.test" {
		t.Errorf("Expected 1 extra domain, got %v", config.ExtraDomains)
	}
}

func TestRemoveDomainLogic(t *testing.T) {
	config := engine.Config{
		Domain:       "main.test",
		ExtraDomains: []string{"extra.test", "another.test"},
	}

	// Logic to remove a domain
	toRemove := "extra.test"
	var updated []string
	for _, d := range config.ExtraDomains {
		if d != toRemove {
			updated = append(updated, d)
		}
	}
	config.ExtraDomains = updated

	if len(config.ExtraDomains) != 1 || config.ExtraDomains[0] != "another.test" {
		t.Errorf("Expected 1 extra domain left, got %v", config.ExtraDomains)
	}
}
