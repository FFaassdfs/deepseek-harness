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
	if cfg.Port != 3080 || cfg.Command != "dsh web" || !cfg.AutoUpdateHarness {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if got := cfg.URL(); got != "http://127.0.0.1:3080" {
		t.Fatalf("unexpected URL: %q", got)
	}
}

func TestLoadConfigAutoUpdateFalse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"autoUpdateHarness": false}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("loadConfigFrom: %v", err)
	}
	if cfg.AutoUpdateHarness {
		t.Fatal("expected autoUpdateHarness=false, got true")
	}
}

func TestLoadConfigWorkDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"workdir": "D:/dev/deepseek-harness"}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := loadConfigFrom(path)
	if err != nil {
		t.Fatalf("loadConfigFrom: %v", err)
	}
	if cfg.WorkDir != "D:/dev/deepseek-harness" {
		t.Fatalf("unexpected workdir: %q", cfg.WorkDir)
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
