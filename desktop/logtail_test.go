package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadLogTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dsh.log")
	content := strings.Repeat("line\n", 100)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	const max = 20
	tail := readLogTail(path, max)
	if tail != content[len(content)-max:] {
		t.Fatalf("tail mismatch: got %q", tail)
	}
}

func TestReadLogTailMissingFile(t *testing.T) {
	if got := readLogTail(filepath.Join(t.TempDir(), "nope.log"), 100); got != "" {
		t.Fatalf("expected empty for missing file, got %q", got)
	}
}

func TestReadLogTailSmallFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dsh.log")
	if err := os.WriteFile(path, []byte("abc"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := readLogTail(path, 100); got != "abc" {
		t.Fatalf("expected whole file, got %q", got)
	}
}
