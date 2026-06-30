package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scaffold creates a temp tree with the given content and data paths.
type scaffold struct {
	dir        string
	contentDir string
	dataFile   string
	formFile   string
	schemaFile string
}

func newScaffold(t *testing.T) *scaffold {
	t.Helper()
	dir := t.TempDir()
	s := &scaffold{
		dir:        dir,
		contentDir: filepath.Join(dir, "help"),
		dataFile:   filepath.Join(dir, "data.csv"),
		formFile:   filepath.Join(dir, "form.tmpl"),
		schemaFile: filepath.Join(dir, "schema.yml"),
	}
	if err := os.Mkdir(s.contentDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{s.dataFile, s.formFile, s.schemaFile} {
		if err := os.WriteFile(f, []byte("placeholder"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func (s *scaffold) config() *Config {
	return &Config{
		Content: []ContentItem{
			{Name: "help", Path: s.contentDir},
		},
		Data: []DataItem{
			{Name: "example", Path: s.dataFile, Form: s.formFile, Schema: s.schemaFile},
		},
	}
}

func TestValidateOK(t *testing.T) {
	s := newScaffold(t)
	if err := s.config().Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestValidateEmptyConfig(t *testing.T) {
	if err := (&Config{}).Validate(); err != nil {
		t.Errorf("Validate() on empty config unexpected error: %v", err)
	}
}

func TestValidateMissingContentDir(t *testing.T) {
	s := newScaffold(t)
	cfg := s.config()
	cfg.Content[0].Path = filepath.Join(s.dir, "nonexistent")

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing content dir, got nil")
	}
	if !strings.Contains(err.Error(), "help") {
		t.Errorf("error should mention item name 'help': %v", err)
	}
}

func TestValidateContentPathIsFile(t *testing.T) {
	s := newScaffold(t)
	cfg := s.config()
	cfg.Content[0].Path = s.dataFile // a file, not a dir

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when content path is a file, got nil")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("error should mention 'directory': %v", err)
	}
}

func TestValidateMissingDataFile(t *testing.T) {
	s := newScaffold(t)
	cfg := s.config()
	cfg.Data[0].Path = filepath.Join(s.dir, "missing.csv")

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing data file, got nil")
	}
	if !strings.Contains(err.Error(), "example") {
		t.Errorf("error should mention item name 'example': %v", err)
	}
}

func TestValidateMissingFormFile(t *testing.T) {
	s := newScaffold(t)
	cfg := s.config()
	cfg.Data[0].Form = filepath.Join(s.dir, "missing.tmpl")

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing form file, got nil")
	}
	if !strings.Contains(err.Error(), "form") {
		t.Errorf("error should mention 'form': %v", err)
	}
}

func TestValidateMissingSchemaFile(t *testing.T) {
	s := newScaffold(t)
	cfg := s.config()
	cfg.Data[0].Schema = filepath.Join(s.dir, "missing.yml")

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing schema file, got nil")
	}
	if !strings.Contains(err.Error(), "schema") {
		t.Errorf("error should mention 'schema': %v", err)
	}
}

func TestValidateCollectsAllErrors(t *testing.T) {
	// Both content dir and data file are missing — both should appear in the error.
	cfg := &Config{
		Content: []ContentItem{
			{Name: "help", Path: "/nonexistent/help"},
		},
		Data: []DataItem{
			{Name: "example", Path: "/nonexistent/data.csv"},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "help") {
		t.Errorf("error should mention content item 'help': %v", err)
	}
	if !strings.Contains(err.Error(), "example") {
		t.Errorf("error should mention data item 'example': %v", err)
	}
}

func TestValidateDataFileIsDir(t *testing.T) {
	s := newScaffold(t)
	cfg := s.config()
	cfg.Data[0].Path = s.contentDir // a directory, not a file

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when data path is a directory, got nil")
	}
	if !strings.Contains(err.Error(), "file") {
		t.Errorf("error should mention 'file': %v", err)
	}
}

func TestValidateOptionalFieldsSkipped(t *testing.T) {
	s := newScaffold(t)
	cfg := s.config()
	// Clear optional form/schema — should still pass.
	cfg.Data[0].Form = ""
	cfg.Data[0].Schema = ""

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error when form/schema are empty: %v", err)
	}
}
