package schema

import (
	"os"
	"testing"
)

const testYAML = `
fields:
  - name: id
    type: string
    label: "ID"
    required: true
    readonly: true
  - name: group
    type: string
    label: "Group"
    required: true
    options:
      - engineering
      - marketing
  - name: active
    type: bool
    label: "Active"
    default: "true"
`

func writeTempSchema(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "schema-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestLoad(t *testing.T) {
	s, err := Load(writeTempSchema(t, testYAML))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(s.Fields) != 3 {
		t.Fatalf("len(Fields) = %d, want 3", len(s.Fields))
	}

	id := s.Fields[0]
	if id.Name != "id" || !id.Required || !id.Readonly {
		t.Errorf("id field = %+v", id)
	}

	grp := s.Fields[1]
	if len(grp.Options) != 2 {
		t.Errorf("group options = %v, want 2", grp.Options)
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load("/nonexistent/schema.yml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestFieldByName(t *testing.T) {
	s, _ := Load(writeTempSchema(t, testYAML))

	f := s.FieldByName("group")
	if f == nil || f.Label != "Group" {
		t.Errorf("FieldByName(group) = %v", f)
	}
	if s.FieldByName("nonexistent") != nil {
		t.Error("FieldByName(nonexistent) should return nil")
	}
}
