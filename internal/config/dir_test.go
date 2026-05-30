package config

import (
	"path/filepath"
	"testing"
)

func TestDirHonorsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdghome")
	got, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	want := filepath.Join("/tmp/xdghome", "openydt-cli")
	if got != want {
		t.Fatalf("Dir() = %s, want %s", got, want)
	}
}

func TestPathUsesDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdghome")
	got, err := Path()
	if err != nil {
		t.Fatalf("Path() error: %v", err)
	}
	want := filepath.Join("/tmp/xdghome", "openydt-cli", "config.json")
	if got != want {
		t.Fatalf("Path() = %s, want %s", got, want)
	}
}
