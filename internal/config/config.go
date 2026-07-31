package config

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	// Project Identity
	Name     = "mcp-server-filesystem"
	Platform = "Secure Filesystem"

	// Safety limits.
	MaxReadFileSize  = 50 * 1024 * 1024 // 50MB max file read
	MaxTreeDepth     = 20               // Max directory tree recursion depth
	MaxSearchResults = 1000             // Max results from search_files
)

const DefaultConfigTemplate = `# mcp-server-filesystem configuration
`

type Config struct {
	v *viper.Viper
}

func New() *Config {
	v := viper.New()
	v.AutomaticEnv()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configDir, err := os.UserConfigDir(); err == nil {
		targetDir := filepath.Join(configDir, "mcp-server-filesystem")
		v.AddConfigPath(targetDir)

		if err := v.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) || os.IsNotExist(err) {
				if mkErr := os.MkdirAll(targetDir, 0o750); mkErr == nil { //nolint:gosec // user config directory uses standard permissions
					targetFile := filepath.Join(targetDir, "config.yaml")
					_ = os.WriteFile(targetFile, []byte(DefaultConfigTemplate), 0o600) //nolint:gosec // default config template is non-sensitive
				}
				_ = v.ReadInConfig()
			} else {
				slog.Error("failed to read config", "error", err)
			}
		}
	}

	cfg := &Config{
		v: v,
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		slog.Info("[Viper] Config file modified", "file", e.Name)
	})

	return cfg
}
