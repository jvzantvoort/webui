package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".webui.yaml")

	if err := WriteDefault(path); err != nil {
		t.Fatalf("WriteDefault() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("written file is empty")
	}
}

func TestWriteDefaultFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".webui.yaml")

	if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	err := WriteDefault(path)
	if err == nil {
		t.Error("expected error when file already exists, got nil")
	}

	// original content must be untouched
	data, _ := os.ReadFile(path)
	if string(data) != "existing" {
		t.Errorf("file content changed, got %q", data)
	}
}

func TestWriteDefaultIsValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".webui.yaml")

	if err := WriteDefault(path); err != nil {
		t.Fatalf("WriteDefault() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() on generated config error = %v", err)
	}
	if cfg.Server.Port == 0 {
		t.Error("generated config has no server port")
	}
	if len(cfg.Content) == 0 {
		t.Error("generated config has no content entries")
	}
	if len(cfg.Data) == 0 {
		t.Error("generated config has no data entries")
	}
}
