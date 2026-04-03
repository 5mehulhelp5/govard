package tests

import (
	"testing"

	"govard/internal/engine"
)

func TestFrameworkDefaultsMagento2(t *testing.T) {
	config, ok := engine.GetFrameworkConfig("magento2")
	if !ok {
		t.Fatal("Expected magento2 framework config")
	}

	if config.DefaultCache != "redis" {
		t.Fatalf("Expected DefaultCache redis, got %s", config.DefaultCache)
	}

	if config.DefaultSearch != "opensearch" {
		t.Fatalf("Expected DefaultSearch opensearch, got %s", config.DefaultSearch)
	}
	if config.DefaultQueue != "none" {
		t.Fatalf("Expected DefaultQueue none, got %s", config.DefaultQueue)
	}
	if config.DefaultNodeVer != "24" {
		t.Fatalf("Expected DefaultNodeVer 24, got %s", config.DefaultNodeVer)
	}
	if config.DefaultDBVer != "11.4" {
		t.Fatalf("Expected DefaultDBVer 11.4, got %s", config.DefaultDBVer)
	}
	if config.DefaultMySQLVer != "8.4" {
		t.Fatalf("Expected DefaultMySQLVer 8.4, got %s", config.DefaultMySQLVer)
	}
	if config.DefaultCacheVer != "7.4" {
		t.Fatalf("Expected DefaultCacheVer 7.4, got %s", config.DefaultCacheVer)
	}

	if config.DefaultSearchVer != "2.19" {
		t.Fatalf("Expected DefaultSearchVer 2.19, got %s", config.DefaultSearchVer)
	}
	if config.DefaultQueueVer != "3.13.7" {
		t.Fatalf("Expected DefaultQueueVer 3.13.7, got %s", config.DefaultQueueVer)
	}
}

func TestFrameworkDefaultsNonMagento2DisableCacheAndSearch(t *testing.T) {
	frameworks := []string{
		"laravel",
		"nextjs",
		"drupal",
		"symfony",
		"magento1",
		"openmage",
		"shopware",
		"cakephp",
		"wordpress",
		"custom",
	}

	for _, framework := range frameworks {
		config, ok := engine.GetFrameworkConfig(framework)
		if !ok {
			t.Fatalf("expected %s framework config", framework)
		}
		if config.DefaultCache != "none" {
			t.Fatalf("expected %s DefaultCache none, got %s", framework, config.DefaultCache)
		}
		if config.DefaultSearch != "none" {
			t.Fatalf("expected %s DefaultSearch none, got %s", framework, config.DefaultSearch)
		}
	}
}
