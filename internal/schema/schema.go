package schema

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Field describes one column in a CSV data file.
type Field struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`    // string, email, number, date, bool
	Label    string   `yaml:"label"`
	Required bool     `yaml:"required"`
	Readonly bool     `yaml:"readonly"`
	Options  []string `yaml:"options"` // non-empty → rendered as <select>
	Default  string   `yaml:"default"`
}

// Schema is the top-level structure of a .yml schema file.
type Schema struct {
	Fields []Field `yaml:"fields"`
}

// Load reads and parses a schema YAML file.
func Load(path string) (*Schema, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open schema: %w", err)
	}
	defer f.Close()

	var s Schema
	if err := yaml.NewDecoder(f).Decode(&s); err != nil {
		return nil, fmt.Errorf("decode schema: %w", err)
	}
	return &s, nil
}

// FieldByName returns the Field with the given name, or nil.
func (s *Schema) FieldByName(name string) *Field {
	for i := range s.Fields {
		if s.Fields[i].Name == name {
			return &s.Fields[i]
		}
	}
	return nil
}
