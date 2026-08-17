package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := loadConfigFrom(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("loadConfigFrom: %v", err)
	}
	if cfg.Port != 3080 || cfg.Command != "dsh web" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if got := cfg.URL(); got != "http://127.0.0.1:3080" {
		t.Fatalf("unexpected URL: %q", got)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"port": 3999, "command": "pnpm dsh web"}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("loadConfigFrom: %v", err)
	}
	if cfg.Port != 3999 || cfg.Command != "pnpm dsh web" {
		t.Fatalf("unexpected overrides: %+v", cfg)
	}
	if got := cfg.URL(); got != "http://127.0.0.1:3999" {
		t.Fatalf("unexpected URL: %q", got)
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{not json`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := loadConfigFrom(path); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
