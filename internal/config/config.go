package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Database  DatabaseConfig  `toml:"database"`
	Email     EmailConfig     `toml:"email"`
	SimpleFin SimpleFinConfig `toml:"simplefin"`
	CSV       CSVConfig       `toml:"csv"`
}

type DatabaseConfig struct {
	Driver      string `toml:"driver"`
	SQLitePath  string `toml:"sqlite_path"`
	PostgresURL string `toml:"postgres_url"`
}

type EmailConfig struct {
	Enabled             bool   `toml:"enabled"`
	IMAPServer          string `toml:"imap_server"`
	IMAPPort            int    `toml:"imap_port"`
	Username            string `toml:"username"`
	AppPassword         string `toml:"app_password"`
	Label               string `toml:"label"`
	PollIntervalMinutes int    `toml:"poll_interval_minutes"`
	AccountName         string `toml:"account_name"`
}

type SimpleFinConfig struct {
	Enabled   bool   `toml:"enabled"`
	AccessURL string `toml:"access_url"`
}

type CSVConfig struct {
	WatchDir string `toml:"watch_dir"`
}

// DefaultDir returns the default miser config/data directory (~/.miser).
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".miser"), nil
}

// DefaultPath returns the default config file path (~/.miser/config.toml).
func DefaultPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads and parses a TOML config file from the given path.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("load config %s: %w", path, err)
	}
	cfg.expandPaths()
	return cfg, nil
}

// Default returns a Config with sensible defaults (before any file is loaded).
func Default() (*Config, error) {
	dir, err := DefaultDir()
	if err != nil {
		return nil, err
	}
	return &Config{
		Database: DatabaseConfig{
			Driver:     "sqlite",
			SQLitePath: filepath.Join(dir, "miser.db"),
		},
		Email: EmailConfig{
			IMAPServer:          "imap.gmail.com",
			IMAPPort:            993,
			Label:               "Finance/Fidelity",
			PollIntervalMinutes: 15,
		},
		CSV: CSVConfig{
			WatchDir: filepath.Join(dir, "import"),
		},
	}, nil
}

// expandPaths resolves ~ prefixes in file paths.
func (c *Config) expandPaths() {
	c.Database.SQLitePath = expandHome(c.Database.SQLitePath)
	c.CSV.WatchDir = expandHome(c.CSV.WatchDir)
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
