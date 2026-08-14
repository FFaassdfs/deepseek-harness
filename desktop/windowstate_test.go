package main

import (
	"path/filepath"
	"testing"
)

func TestWindowStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "window.json")

	orig := &WindowState{Width: 1100, Height: 720, X: 50, Y: 30, Maximised: false}
	if err := orig.saveTo(path); err != nil {
		t.Fatalf("saveTo: %v", err)
	}

	got, err := loadWindowStateFrom(path)
	if err != nil {
		t.Fatalf("loadWindowStateFrom: %v", err)
	}
	if *got != *orig {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, orig)
	}
}

func TestWindowStateRoundTripMaximised(t *testing.T) {
	path := filepath.Join(t.TempDir(), "window.json")

	orig := &WindowState{Width: 1280, Height: 860, X: 0, Y: 0, Maximised: true}
	if err := orig.saveTo(path); err != nil {
		t.Fatalf("saveTo: %v", err)
	}

	got, err := loadWindowStateFrom(path)
	if err != nil {
		t.Fatalf("loadWindowStateFrom: %v", err)
	}
	if *got != *orig {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, orig)
	}
}

func TestLoadWindowStateMissingFile(t *testing.T) {
	if _, err := loadWindowStateFrom(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
