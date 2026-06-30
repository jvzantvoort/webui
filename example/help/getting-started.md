# Getting started

## 1. Initialise a repository

webui expects to run from the root of a directory (ideally a git repo).
Run `webui init` to create a `.webui.yaml` config with sensible defaults:

```bash
mkdir my-project && cd my-project
git init
webui init
```

## 2. Edit the config

Open `.webui.yaml` and adjust the paths to match your files:

```yaml
server:
  port: 3110
  index: README.md

content:
  - name: "docs"
    path: docs/
    content: markdown
    menu: "Documentation"

data:
  - name: "contacts"
    path: data/contacts.csv
    schema: schema/contacts.yml
    menu: "Data / Contacts"
```

## 3. Create your data files

CSV files need a header row.  Keep them committed to git so you have a
full edit history.  The header names must match the `name:` fields in
your schema exactly.

```
id,first_name,last_name,email,department,active
1,Alice,Harding,alice.harding@example.com,Finance,yes
2,Bob,Steele,bob.steele@example.com,Marketing,yes
```

### Quoting rules

Follow standard CSV quoting when a cell value contains a comma, a
newline, or a double-quote character:

| Situation | Example cell | CSV encoding |
|-----------|-------------|--------------|
| Value contains a comma | `Smith, Jr.` | `"Smith, Jr."` |
| Value contains a double-quote | `27" Monitor` | `"27"" Monitor"` |
| Plain value | `Alice` | `Alice` |

Most spreadsheet applications (Excel, LibreOffice Calc, Numbers) produce
correctly quoted CSV on export.  webui reads CSV with lazy-quote handling
enabled, so minor quoting inconsistencies in hand-edited files will not
cause a parse error.

## 4. Optionally add a schema

A schema YAML file lets you control how the edit form looks — field
labels, input types, required/readonly flags, and dropdown options.
Columns not listed in the schema are shown in the table but hidden from
the edit form.

```yaml
fields:
  - name: id
    type: string
    label: "ID"
    readonly: true        # displayed but never overwritten on save

  - name: first_name
    type: string
    label: "First name"
    required: true

  - name: email
    type: email           # browser validates the format
    label: "Email address"
    required: true

  - name: department
    type: string
    label: "Department"
    options:              # rendered as a <select> dropdown
      - Engineering
      - Finance
      - HR
      - Marketing
      - Sales
```

See [Configuration](configuration.md) for the full field-type reference.

## 5. Start the server

```bash
# From the directory that contains .webui.yaml:
webui serve

# Or point at a config file explicitly:
webui serve -config /path/to/.webui.yaml
```

The browser opens automatically if `browser.autostart: true` is set.
Navigate to `http://localhost:3110` to view the UI.

## 6. Check the version

```bash
webui version
```
