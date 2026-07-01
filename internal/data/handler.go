package data

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jvzantvoort/webui/internal/config"
	"github.com/jvzantvoort/webui/internal/schema"
	"github.com/jvzantvoort/webui/internal/tmpl"
)

// Handler serves data items (CSV files with CRUD forms).
type Handler struct {
	item config.DataItem
	r    *tmpl.Renderer
}

// NewHandler creates an http.Handler for the given DataItem.
func NewHandler(item config.DataItem, r *tmpl.Renderer) http.Handler {
	return &Handler{item: item, r: r}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	switch r.Method {
	case http.MethodGet:
		switch action {
		case "view":
			h.view(w, r)
		case "edit":
			h.editForm(w, r)
		default:
			h.list(w, r)
		}
	case http.MethodPost:
		switch action {
		case "edit":
			h.saveEdit(w, r)
		default:
			h.create(w, r)
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// list renders the data table.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.readCSV()
	if err != nil {
		http.Error(w, "failed to read data", http.StatusInternalServerError)
		return
	}

	var headers []string
	var indexed []tmpl.IndexedRow
	if len(rows) > 0 {
		headers = rows[0]
		for i, row := range rows[1:] {
			if !isEmptyRow(row) {
				indexed = append(indexed, tmpl.IndexedRow{Idx: i + 1, Cells: row})
			}
		}
	}

	h.r.Render(w, "data.html", tmpl.PageData{
		Title:   tmpl.MenuLabel(h.item.Menu, h.item.Name),
		Headers: headers,
		Rows:    indexed,
	})
}

// editForm renders a pre-filled form for the row identified by the "row" query param.
func (h *Handler) editForm(w http.ResponseWriter, r *http.Request) {
	rowIdx, err := strconv.Atoi(r.URL.Query().Get("row"))
	if err != nil || rowIdx < 1 {
		http.Error(w, "invalid row parameter", http.StatusBadRequest)
		return
	}

	rows, err := h.readCSV()
	if err != nil {
		http.Error(w, "failed to read data", http.StatusInternalServerError)
		return
	}
	if rowIdx >= len(rows) {
		http.NotFound(w, r)
		return
	}

	headers := rows[0]
	values := rowToMap(headers, rows[rowIdx])
	fields := h.buildFormFields(headers, values)

	h.r.Render(w, "form.html", tmpl.PageData{
		Title:   "Edit — " + tmpl.MenuLabel(h.item.Menu, h.item.Name),
		Fields:  fields,
		RowIdx:  rowIdx,
		PostURL: fmt.Sprintf("?action=edit&row=%d", rowIdx),
	})
}

// saveEdit validates the submitted form and writes the updated row back to the CSV.
func (h *Handler) saveEdit(w http.ResponseWriter, r *http.Request) {
	rowIdx, err := strconv.Atoi(r.URL.Query().Get("row"))
	if err != nil || rowIdx < 1 {
		http.Error(w, "invalid row parameter", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	rows, err := h.readCSV()
	if err != nil {
		http.Error(w, "failed to read data", http.StatusInternalServerError)
		return
	}
	if rowIdx >= len(rows) {
		http.NotFound(w, r)
		return
	}

	headers := rows[0]
	readOnly := h.readonlySet()

	newRow := make([]string, len(headers))
	for i, col := range headers {
		if readOnly[col] {
			if i < len(rows[rowIdx]) {
				newRow[i] = rows[rowIdx][i] // preserve original value
			}
		} else {
			newRow[i] = r.FormValue(col)
		}
	}
	rows[rowIdx] = newRow

	if err := h.writeCSV(rows); err != nil {
		http.Error(w, "failed to save data", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

// view renders a single record using the optional record_template, falling back
// to a generated definition list. Prev/Next buttons allow navigating between rows.
func (h *Handler) view(w http.ResponseWriter, r *http.Request) {
	rowIdx, err := strconv.Atoi(r.URL.Query().Get("row"))
	if err != nil || rowIdx < 1 {
		http.Error(w, "invalid row parameter", http.StatusBadRequest)
		return
	}

	rows, err := h.readCSV()
	if err != nil {
		http.Error(w, "failed to read data", http.StatusInternalServerError)
		return
	}
	if rowIdx >= len(rows) {
		http.NotFound(w, r)
		return
	}

	headers := rows[0]
	record := rowToMap(headers, rows[rowIdx])
	fields := h.buildFormFields(headers, record)

	// Find the nearest non-empty rows for prev/next navigation.
	prevIdx := 0
	for i := rowIdx - 1; i >= 1; i-- {
		if !isEmptyRow(rows[i]) {
			prevIdx = i
			break
		}
	}
	nextIdx := 0
	for i := rowIdx + 1; i < len(rows); i++ {
		if !isEmptyRow(rows[i]) {
			nextIdx = i
			break
		}
	}

	var body template.HTML
	if h.item.RecordTemplate != "" {
		body, err = h.renderRecordTemplate(record)
		if err != nil {
			http.Error(w, "failed to render record template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.r.Render(w, "view.html", tmpl.PageData{
		Title:   tmpl.MenuLabel(h.item.Menu, h.item.Name),
		Fields:  fields,
		Body:    body,
		Record:  record,
		RowIdx:  rowIdx,
		PrevIdx: prevIdx,
		NextIdx: nextIdx,
	})
}

// renderRecordTemplate parses the configured record_template file as an
// html/template and executes it with the record values as dot data.
// Column names become top-level keys: {{.first_name}}, {{.unit_price}}, etc.
// Values are HTML-escaped automatically by html/template.
func (h *Handler) renderRecordTemplate(record map[string]string) (template.HTML, error) {
	src, err := os.ReadFile(h.item.RecordTemplate)
	if err != nil {
		return "", fmt.Errorf("read record template: %w", err)
	}

	t, err := template.New("record").Parse(string(src))
	if err != nil {
		return "", fmt.Errorf("parse record template: %w", err)
	}

	// Convert map[string]string to map[string]interface{} so that html/template
	// supports both {{.key}} dot access and {{index . "key"}} for unusual names.
	data := make(map[string]interface{}, len(record))
	for k, v := range record {
		data[k] = v
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute record template: %w", err)
	}
	return template.HTML(buf.String()), nil
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// buildFormFields returns form fields driven by the schema if one is configured,
// falling back to plain text inputs for every CSV column.
func (h *Handler) buildFormFields(headers []string, values map[string]string) []tmpl.FormField {
	if h.item.Schema != "" {
		if sc, err := schema.Load(h.item.Schema); err == nil {
			return formFieldsFromSchema(sc, values)
		}
	}
	return formFieldsFromHeaders(headers, values)
}

// readonlySet returns the set of field names that must not be modified by the user.
func (h *Handler) readonlySet() map[string]bool {
	ro := map[string]bool{}
	if h.item.Schema == "" {
		return ro
	}
	sc, err := schema.Load(h.item.Schema)
	if err != nil {
		return ro
	}
	for _, f := range sc.Fields {
		if f.Readonly {
			ro[f.Name] = true
		}
	}
	return ro
}

// formFieldsFromSchema converts schema fields to tmpl.FormField, filling current values.
func formFieldsFromSchema(sc *schema.Schema, values map[string]string) []tmpl.FormField {
	fields := make([]tmpl.FormField, 0, len(sc.Fields))
	for _, f := range sc.Fields {
		label := f.Label
		if label == "" {
			label = f.Name
		}
		cur := values[f.Name]
		ff := tmpl.FormField{
			Name:      f.Name,
			Label:     label,
			InputType: schemaTypeToInput(f.Type),
			Value:     cur,
			Required:  f.Required,
			Readonly:  f.Readonly,
		}
		for _, opt := range f.Options {
			ff.Options = append(ff.Options, tmpl.FormOption{
				Value:    opt,
				Selected: opt == cur,
			})
		}
		fields = append(fields, ff)
	}
	return fields
}

// formFieldsFromHeaders builds plain text fields for every CSV column.
func formFieldsFromHeaders(headers []string, values map[string]string) []tmpl.FormField {
	fields := make([]tmpl.FormField, 0, len(headers))
	for _, h := range headers {
		fields = append(fields, tmpl.FormField{
			Name:      h,
			Label:     h,
			InputType: "text",
			Value:     values[h],
		})
	}
	return fields
}

// schemaTypeToInput maps a schema type name to an HTML input type.
func schemaTypeToInput(t string) string {
	switch t {
	case "email":
		return "email"
	case "number":
		return "number"
	case "date":
		return "date"
	default:
		return "text"
	}
}

// rowToMap zips a header slice and a row slice into a name→value map.
func rowToMap(headers, row []string) map[string]string {
	m := make(map[string]string, len(headers))
	for i, h := range headers {
		if i < len(row) {
			m[h] = row[i]
		}
	}
	return m
}

func (h *Handler) readCSV() ([][]string, error) {
	f, err := os.Open(h.item.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	// Strip UTF-8 BOM if present (bytes 0xEF 0xBB 0xBF).
	// Excel and other Windows tools prepend a BOM to CSV exports, which ends
	// up embedded in the first header name. Map lookups by the plain column
	// name then return "" and the edit form appears empty.
	if bom, err := br.Peek(3); err == nil && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		_, _ = br.Discard(3)
	}

	cr := csv.NewReader(br)
	cr.FieldsPerRecord = -1  // allow rows with fewer/more fields than the header
	cr.LazyQuotes = true     // accept unescaped quotes in field values
	cr.TrimLeadingSpace = true

	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	// Trim all cell values: removes trailing \r (Windows line endings that
	// slip through), extra surrounding whitespace, and non-breaking spaces.
	for i, row := range rows {
		for j, cell := range row {
			rows[i][j] = strings.TrimSpace(cell)
		}
	}
	return rows, nil
}

// isEmptyRow reports whether every cell in the row is an empty string.
// Used to skip blank lines that real-world CSV files often contain.
func isEmptyRow(row []string) bool {
	for _, v := range row {
		if v != "" {
			return false
		}
	}
	return true
}

// writeCSV writes rows atomically via a temp file in the same directory.
func (h *Handler) writeCSV(rows [][]string) error {
	dir := filepath.Dir(h.item.Path)
	tmp, err := os.CreateTemp(dir, ".webui-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op if rename succeeded

	w := csv.NewWriter(tmp)
	if err := w.WriteAll(rows); err != nil {
		tmp.Close()
		return err
	}
	w.Flush()
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, h.item.Path)
}
