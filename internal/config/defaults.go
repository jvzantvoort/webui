package config

import (
	"fmt"
	"os"
)

const DefaultConfig = `server:
  port: 3110
  index: README.md

browser:
  application: firefox
  autostart: true

content:
  - name: "help"
    path: help/
    content: markdown
    serve_images: true
    menu: "Help"

data:
  - name: "example"
    path: data/example.csv
    form: .webui/content/example.tmpl
    schema: .webui/content/example.yml
    menu: "Data / Example"
`

// WriteDefault writes the default config template to path.
// Returns an error if the file already exists.
func WriteDefault(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}
	return os.WriteFile(path, []byte(DefaultConfig), 0644)
}
