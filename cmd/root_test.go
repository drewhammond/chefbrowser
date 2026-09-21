package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// Guards the INI codec wiring: viper >= 1.20 dropped built-in INI support, and
// inline '#' must survive in values (#399).
func TestInitConfigINI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "user.ini")
	if err := os.WriteFile(path, []byte(`
[logging]
level = debug

[custom_links.nodes.0]
title = Dash
href = https://example.test/d/1?host={{fqdn}}#panel-3
new_tab = true
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfgFile = path
	initConfig()

	if cfg.Logging.Level != "debug" {
		t.Errorf("logging.level = %q, want debug", cfg.Logging.Level)
	}
	if cfg.App.ListenAddr != "0.0.0.0:8080" {
		t.Errorf("default listen_addr not merged, got %q", cfg.App.ListenAddr)
	}
	link := cfg.CustomLinks.Nodes[0]
	if link.Href != "https://example.test/d/1?host={{fqdn}}#panel-3" || !link.NewTab {
		t.Errorf("custom link decoded wrong: %+v", link)
	}
}
