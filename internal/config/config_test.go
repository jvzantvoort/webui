package config

import (
	"os"
	"testing"
)

const testYAML = `
server:
  port: 3110
  index: README.md
browser:
  application: firefox
  autostart: true
content:
  - name: help
    path: help/
    content: markdown
    serve_images: true
    menu: Help
data:
  - name: company groups
    path: company/groups.csv
    form: .webui/content/groupdata.tmpl
    schema: .webui/content/groupdata.yml
    menu: company
`

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "webui-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestLoad(t *testing.T) {
	cfg, err := Load(writeTempConfig(t, testYAML))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 3110 {
		t.Errorf("Server.Port = %d, want 3110", cfg.Server.Port)
	}
	if cfg.Server.Index != "README.md" {
		t.Errorf("Server.Index = %q, want README.md", cfg.Server.Index)
	}
	if cfg.Browser.Application != "firefox" {
		t.Errorf("Browser.Application = %q, want firefox", cfg.Browser.Application)
	}
	if !cfg.Browser.Autostart {
		t.Error("Browser.Autostart = false, want true")
	}
	if len(cfg.Content) != 1 {
		t.Fatalf("len(Content) = %d, want 1", len(cfg.Content))
	}
	if cfg.Content[0].Name != "help" {
		t.Errorf("Content[0].Name = %q, want help", cfg.Content[0].Name)
	}
	if !cfg.Content[0].ServeImages {
		t.Error("Content[0].ServeImages = false, want true")
	}
	if len(cfg.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(cfg.Data))
	}
}

func TestLoadDefaultPort(t *testing.T) {
	cfg, err := Load(writeTempConfig(t, "server:\n  index: README.md\n"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Port != 3110 {
		t.Errorf("Server.Port = %d, want 3110 (default)", cfg.Server.Port)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/to/config.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	_, err := Load(writeTempConfig(t, ":\tinvalid::yaml"))
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}
