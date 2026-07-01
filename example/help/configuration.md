# Configuration reference

The config file is `.webui.yaml` by default, read from the working directory.

## server

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `port` | int | `3110` | HTTP listen port |
| `index` | string | — | File to serve at `/` (e.g. `README.md`) |

## browser

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `application` | string | — | Browser binary to launch (e.g. `firefox`) |
| `autostart` | bool | `false` | Open the browser on server start |

## content items

Each entry under `content:` defines a folder served as readable content.

```yaml
content:
  - name: "help"        # URL segment: /help/
    path: help/         # path relative to the working directory
    content: markdown   # "markdown" renders .md files; anything else serves raw
    serve_images: true  # also serve image files (png, jpg, gif, svg, webp)
    menu: "Help"        # label shown in the navigation bar
```

## CSV file format

CSV files must have a header row.  Values that contain a comma, a
newline, or a double-quote must be enclosed in double-quotes.  A
literal double-quote inside a quoted value is escaped by doubling it:

```
sku,name,quantity
MON-001,"27"" 4K Monitor",15
PAPER-A4,A4 Paper (500 sheets),200
```

webui reads CSV with `LazyQuotes` enabled, so minor quoting
inconsistencies (common in hand-edited files) will not abort the
parse.  A UTF-8 BOM — prepended by Microsoft Excel on export — is
stripped automatically.

## data items

Each entry under `data:` defines a CSV file with a CRUD table.

```yaml
data:
  - name: "contacts"                          # URL segment: /contacts/
    path: data/contacts.csv                   # CSV file path (relative to working directory)
    schema: schema/contacts.yml               # optional: controls the edit form
    record_template: templates/contact.html   # optional: custom record view
    menu: "Data / Contacts"                   # label shown in the navigation bar
```

The `menu` value may use ` / ` as a visual separator; it does not create
nested routes — it is purely a display label.

### record_template

When `record_template` is set, clicking **View** on a data row renders that
record using the template file instead of the default definition list.

The template is a standard Go `html/template` fragment (not a full page — the
navbar and layout are provided by webui).  Column values are available as
top-level dot keys matching the CSV header names exactly:

```html
<h3>{{.first_name}} {{.last_name}}</h3>
<p>{{.email}}</p>
```

For column names that are not valid Go identifiers (e.g. contain spaces), use
the `index` function instead:

```html
<p>{{index . "unit price"}}</p>
```

Values are HTML-escaped automatically.  The template may use any Bootstrap
classes since the Bootstrap CSS is always loaded by the base layout.

Navigation buttons (← Back, ‹ Prev, Next ›, Edit) are provided by the layout
and do not need to be included in the template file.

## Schema file format

Schema files are YAML.  Each entry in `fields:` maps to one CSV column.

```yaml
fields:
  - name: id           # must match the CSV header exactly
    type: string       # string | email | number | date
    label: "ID"        # display label in the edit form
    readonly: true     # field is shown but not editable
    required: false

  - name: status
    type: string
    label: "Status"
    options:           # non-empty list → rendered as <select>
      - Active
      - Inactive
      - Pending
```

### Field types

| Type | HTML input | Notes |
|------|-----------|-------|
| `string` | `text` | default |
| `email` | `email` | browser validates format |
| `number` | `number` | browser validates numeric |
| `date` | `date` | browser date picker |

Columns not listed in the schema are still displayed in the table but
are hidden from the edit form.
