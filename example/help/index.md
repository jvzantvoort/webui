# webui

**webui** is a lightweight tool that turns local files stored in a git repository
into a browser-based UI.  It serves markdown folders as readable documentation
and CSV files as editable data tables — no database required.

## Quick start

```bash
# Generate a starter config in the current directory
webui init

# Start the server (reads .webui.yaml in the current directory)
webui serve

# Override the config file location
webui serve -config /path/to/.webui.yaml
```

## Navigation

- **Data items** appear on the left side of the top bar.
- **Content items** appear to the right of the separator.
- Sections with multiple files show a dropdown menu.
- The **Stop** button shuts the server down and attempts to close the tab.

## Sections

### Content

A content section points at a directory of files.  Markdown files (`.md`) are
rendered as HTML.  All other files are served as-is.  If the folder contains
two or more files a dropdown menu is generated automatically.

### Data

A data section points at a single CSV file.  The first row is treated as the
header.  Each subsequent row becomes a table row with **Edit** and **Delete**
actions.  A schema file (optional) controls field types, labels, required
flags, and dropdown options.

See [Configuration](configuration.md) for the full reference.
