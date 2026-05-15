package tests

import (
	"os"
	"path/filepath"
	"testing"

	"govard/internal/engine"
)

func TestNormalizeConfigDetectsMagento2TablePrefix(t *testing.T) {
	root := t.TempDir()
	etcDir := filepath.Join(root, "app", "etc")
	if err := os.MkdirAll(etcDir, 0o755); err != nil {
		t.Fatalf("mkdir app/etc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(etcDir, "env.php"), []byte(`<?php
return [
    'db' => [
        'table_prefix' => 'magspas_',
    ],
];
`), 0o644); err != nil {
		t.Fatalf("write env.php: %v", err)
	}

	config := engine.Config{Framework: "magento2"}
	engine.NormalizeConfig(&config, root)

	if config.TablePrefix != "magspas_" {
		t.Fatalf("expected table prefix magspas_, got %q", config.TablePrefix)
	}
}

func TestNormalizeConfigDetectsMagento1TablePrefix(t *testing.T) {
	root := t.TempDir()
	etcDir := filepath.Join(root, "app", "etc")
	if err := os.MkdirAll(etcDir, 0o755); err != nil {
		t.Fatalf("mkdir app/etc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(etcDir, "local.xml"), []byte(`<?xml version="1.0"?>
<config>
    <global>
        <resources>
            <db>
                <table_prefix><![CDATA[magspas_]]></table_prefix>
            </db>
        </resources>
    </global>
</config>
`), 0o644); err != nil {
		t.Fatalf("write local.xml: %v", err)
	}

	config := engine.Config{Framework: "openmage"}
	engine.NormalizeConfig(&config, root)

	if config.TablePrefix != "magspas_" {
		t.Fatalf("expected table prefix magspas_, got %q", config.TablePrefix)
	}
}

func TestValidateConfigRejectsUnsafeTablePrefix(t *testing.T) {
	config := engine.Config{
		ProjectName: "sample-project",
		Framework:   "magento2",
		Domain:      "sample.test",
		TablePrefix: "magspas_;DROP",
	}
	engine.NormalizeConfig(&config, "")

	if err := engine.ValidateConfig(config); err == nil {
		t.Fatal("expected unsafe table_prefix to be rejected")
	}
}
