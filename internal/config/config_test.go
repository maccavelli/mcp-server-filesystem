package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	// Isolate UserConfigDir: HOME covers macOS, XDG_CONFIG_HOME covers Linux.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := New()
	if cfg == nil {
		t.Fatal("expected config to be created")
	}
	if cfg.v == nil {
		t.Fatal("expected viper to be initialized")
	}
}

func TestConstants(t *testing.T) {
	if Name != "mcp-server-filesystem" {
		t.Errorf("unexpected name: %s", Name)
	}
	if Platform != "Secure Filesystem" {
		t.Errorf("unexpected platform: %s", Platform)
	}
	if MaxReadFileSize != 52428800 {
		t.Errorf("unexpected max read file size: %d", MaxReadFileSize)
	}
}
