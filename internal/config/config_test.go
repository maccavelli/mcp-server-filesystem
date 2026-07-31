package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// Override HOME to use temp dir for UserConfigDir
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

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
