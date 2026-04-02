package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `
[database]
driver = "sqlite"
sqlite_path = "/tmp/test.db"

[email]
enabled = true
imap_server = "imap.gmail.com"
imap_port = 993
username = "test@example.com"
app_password = "secret"
label = "Finance/Fidelity"
poll_interval_minutes = 15
account_name = "Test Account"

[simplefin]
enabled = false
access_url = ""

[csv]
watch_dir = "/tmp/import"
`
	tmpFile := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Database.Driver = %q, want %q", cfg.Database.Driver, "sqlite")
	}
	if cfg.Database.SQLitePath != "/tmp/test.db" {
		t.Errorf("Database.SQLitePath = %q, want %q", cfg.Database.SQLitePath, "/tmp/test.db")
	}
	if !cfg.Email.Enabled {
		t.Error("Email.Enabled = false, want true")
	}
	if cfg.Email.IMAPPort != 993 {
		t.Errorf("Email.IMAPPort = %d, want %d", cfg.Email.IMAPPort, 993)
	}
	if cfg.Email.Username != "test@example.com" {
		t.Errorf("Email.Username = %q, want %q", cfg.Email.Username, "test@example.com")
	}
	if cfg.CSV.WatchDir != "/tmp/import" {
		t.Errorf("CSV.WatchDir = %q, want %q", cfg.CSV.WatchDir, "/tmp/import")
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/foo/bar", filepath.Join(home, "foo", "bar")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, tt := range tests {
		got := expandHome(tt.input)
		if got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDefault(t *testing.T) {
	cfg, err := Default()
	if err != nil {
		t.Fatalf("Default() error: %v", err)
	}

	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Database.Driver = %q, want %q", cfg.Database.Driver, "sqlite")
	}
	if cfg.Email.IMAPServer != "imap.gmail.com" {
		t.Errorf("Email.IMAPServer = %q, want %q", cfg.Email.IMAPServer, "imap.gmail.com")
	}
	if cfg.Email.PollIntervalMinutes != 15 {
		t.Errorf("Email.PollIntervalMinutes = %d, want %d", cfg.Email.PollIntervalMinutes, 15)
	}
}

func TestLoadExpandsTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	content := `
[database]
driver = "sqlite"
sqlite_path = "~/.miser/miser.db"

[csv]
watch_dir = "~/.miser/import"
`
	tmpFile := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	wantDB := filepath.Join(home, ".miser", "miser.db")
	if cfg.Database.SQLitePath != wantDB {
		t.Errorf("Database.SQLitePath = %q, want %q", cfg.Database.SQLitePath, wantDB)
	}

	wantCSV := filepath.Join(home, ".miser", "import")
	if cfg.CSV.WatchDir != wantCSV {
		t.Errorf("CSV.WatchDir = %q, want %q", cfg.CSV.WatchDir, wantCSV)
	}
}
