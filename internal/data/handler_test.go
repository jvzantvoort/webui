package data

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jvzantvoort/webui/internal/config"
	"github.com/jvzantvoort/webui/internal/tmpl"
)

func newTestRenderer(t *testing.T) *tmpl.Renderer {
	t.Helper()
	r, err := tmpl.New(&config.Config{})
	if err != nil {
		t.Fatalf("tmpl.New() error = %v", err)
	}
	return r
}

const testCSV = "name,value\nalpha,1\nbeta,2\n"

func writeCSVFile(t *testing.T, content string) (dir, path string) {
	t.Helper()
	dir = t.TempDir()
	path = filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return
}

func TestListCSV(t *testing.T) {
	_, csvPath := writeCSVFile(t, testCSV)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/data/test/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "alpha") {
		t.Errorf("body missing 'alpha': %s", w.Body.String())
	}
}

func TestListCSVMissingFile(t *testing.T) {
	h := NewHandler(config.DataItem{Name: "test", Path: "/nonexistent/data.csv"}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/data/test/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestEditFormRendered(t *testing.T) {
	_, csvPath := writeCSVFile(t, testCSV)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/data/test/?action=edit&row=1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "alpha") {
		t.Errorf("form missing field value 'alpha': %s", body)
	}
	if !strings.Contains(body, `action=edit`) || !strings.Contains(body, `row=1`) {
		t.Errorf("form action URL missing: %s", body)
	}
}

func TestEditFormInvalidRow(t *testing.T) {
	_, csvPath := writeCSVFile(t, testCSV)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	for _, rowParam := range []string{"abc", "0", "999", ""} {
		req := httptest.NewRequest(http.MethodGet, "/data/test/?action=edit&row="+rowParam, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			t.Errorf("row=%q: expected non-200, got 200", rowParam)
		}
	}
}

func TestSaveEdit(t *testing.T) {
	_, csvPath := writeCSVFile(t, testCSV)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	form := url.Values{"name": {"gamma"}, "value": {"3"}}
	req := httptest.NewRequest(http.MethodPost, "/data/test/?action=edit&row=1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d (redirect)", w.Code, http.StatusSeeOther)
	}

	// Verify the CSV was updated.
	updated, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "gamma") {
		t.Errorf("CSV not updated, got: %s", updated)
	}
	if strings.Contains(string(updated), "alpha") {
		t.Errorf("old value 'alpha' still present after edit: %s", updated)
	}
}

func TestSaveEditInvalidRow(t *testing.T) {
	_, csvPath := writeCSVFile(t, testCSV)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodPost, "/data/test/?action=edit&row=99", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCreateNotImplemented(t *testing.T) {
	h := NewHandler(config.DataItem{Name: "test", Path: "/dev/null"}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodPost, "/data/test/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := NewHandler(config.DataItem{Name: "test", Path: "/dev/null"}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodDelete, "/data/test/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// ── view handler tests ────────────────────────────────────────────────────

func TestViewRendered(t *testing.T) {
	_, csvPath := writeCSVFile(t, testCSV)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/data/test/?action=view&row=1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "alpha") {
		t.Errorf("view missing field value 'alpha': %s", body)
	}
	// Prev should be disabled (first data row), Next should link to row 2.
	if !strings.Contains(body, "row=2") {
		t.Errorf("view missing next link to row 2: %s", body)
	}
}

func TestViewPrevNext(t *testing.T) {
	csv := "a,b\nfirst,1\nsecond,2\nthird,3\n"
	_, csvPath := writeCSVFile(t, csv)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	// Middle row: expect both prev and next links.
	req := httptest.NewRequest(http.MethodGet, "/data/test/?action=view&row=2", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "row=1") {
		t.Errorf("middle row missing prev link (row=1): %s", body)
	}
	if !strings.Contains(body, "row=3") {
		t.Errorf("middle row missing next link (row=3): %s", body)
	}
}

func TestViewSkipsEmptyRows(t *testing.T) {
	// Go's csv.Reader silently drops blank lines, so use an all-empty-cells row
	// ("," for a 2-column CSV) to produce a real empty row that survives parsing.
	content := "a,b\nfirst,1\n,\nthird,3\n"
	_, csvPath := writeCSVFile(t, content)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	// Row 1 → next non-empty is row 3 (row 2 is all-empty cells).
	req := httptest.NewRequest(http.MethodGet, "/data/test/?action=view&row=1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "row=3") {
		t.Errorf("next link should skip empty row 2 and point to row=3: %s", w.Body.String())
	}
}

func TestViewInvalidRow(t *testing.T) {
	_, csvPath := writeCSVFile(t, testCSV)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	for _, param := range []string{"0", "999", "abc", ""} {
		req := httptest.NewRequest(http.MethodGet, "/data/test/?action=view&row="+param, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			t.Errorf("row=%q: expected non-200, got 200", param)
		}
	}
}

func TestViewCustomTemplate(t *testing.T) {
	dir, csvPath := writeCSVFile(t, testCSV)

	// Write a simple record template file.
	tmplPath := filepath.Join(dir, "record.html")
	if err := os.WriteFile(tmplPath, []byte(`<p class="custom">{{.name}}: {{.value}}</p>`), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(config.DataItem{
		Name:           "test",
		Path:           csvPath,
		RecordTemplate: tmplPath,
	}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/data/test/?action=view&row=1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `class="custom"`) {
		t.Errorf("custom template output missing in response: %s", body)
	}
	if !strings.Contains(body, "alpha") {
		t.Errorf("custom template missing field value 'alpha': %s", body)
	}
}

func TestViewCustomTemplateMissingFile(t *testing.T) {
	_, csvPath := writeCSVFile(t, testCSV)
	h := NewHandler(config.DataItem{
		Name:           "test",
		Path:           csvPath,
		RecordTemplate: "/nonexistent/record.html",
	}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/data/test/?action=view&row=1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ── readCSV failure-mode regression tests ─────────────────────────────────

// TestReadCSVBOM verifies that a UTF-8 BOM prepended by Excel does not corrupt
// the first header name and therefore does not cause edit forms to show empty fields.
func TestReadCSVBOM(t *testing.T) {
	bom := "\xef\xbb\xbf"
	content := bom + "name,value\nalpha,1\n"
	_, csvPath := writeCSVFile(t, content)

	h := &Handler{item: config.DataItem{Name: "test", Path: csvPath}}
	rows, err := h.readCSV()
	if err != nil {
		t.Fatalf("readCSV() error = %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("readCSV() returned no rows")
	}
	if rows[0][0] != "name" {
		t.Errorf("first header = %q, want %q (BOM was not stripped)", rows[0][0], "name")
	}
}

// TestReadCSVVariableFields verifies that rows with fewer columns than the header
// are parsed without error (FieldsPerRecord=-1).
func TestReadCSVVariableFields(t *testing.T) {
	content := "a,b,c\n1,2\n3,4,5\n"
	_, csvPath := writeCSVFile(t, content)

	h := &Handler{item: config.DataItem{Name: "test", Path: csvPath}}
	rows, err := h.readCSV()
	if err != nil {
		t.Fatalf("readCSV() error on variable-length rows = %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("got %d rows, want 3", len(rows))
	}
}

// TestReadCSVLazyQuotes verifies that unescaped quotes inside field values do
// not cause a parse error (LazyQuotes=true).
func TestReadCSVLazyQuotes(t *testing.T) {
	content := "name,note\nalpha,it's a \"test\" value\n"
	_, csvPath := writeCSVFile(t, content)

	h := &Handler{item: config.DataItem{Name: "test", Path: csvPath}}
	rows, err := h.readCSV()
	if err != nil {
		t.Fatalf("readCSV() error on lazy-quoted CSV = %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("got %d rows, want at least 2", len(rows))
	}
	if !strings.Contains(rows[1][1], "test") {
		t.Errorf("cell value = %q, expected to contain 'test'", rows[1][1])
	}
}

// TestReadCSVWindowsCRLF verifies that Windows-style line endings (CRLF) do
// not leave a trailing \r in cell values.
func TestReadCSVWindowsCRLF(t *testing.T) {
	content := "name,value\r\nalpha,1\r\nbeta,2\r\n"
	_, csvPath := writeCSVFile(t, content)

	h := &Handler{item: config.DataItem{Name: "test", Path: csvPath}}
	rows, err := h.readCSV()
	if err != nil {
		t.Fatalf("readCSV() error on CRLF CSV = %v", err)
	}
	for i, row := range rows {
		for j, cell := range row {
			if strings.HasSuffix(cell, "\r") {
				t.Errorf("rows[%d][%d] = %q has trailing \\r", i, j, cell)
			}
		}
	}
}

// TestReadCSVEmptyRowsSkipped verifies that blank lines in a CSV file are
// filtered out by list() and not rendered as data rows.
func TestReadCSVEmptyRowsSkipped(t *testing.T) {
	content := "name,value\nalpha,1\n\nbeta,2\n\n"
	_, csvPath := writeCSVFile(t, content)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/data/test/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "alpha") || !strings.Contains(body, "beta") {
		t.Errorf("body missing data rows — alpha:%v beta:%v\n%s",
			strings.Contains(body, "alpha"), strings.Contains(body, "beta"), body)
	}
}

// TestIsEmptyRow covers the isEmptyRow helper directly.
func TestIsEmptyRow(t *testing.T) {
	if !isEmptyRow([]string{"", "", ""}) {
		t.Error("all-empty row not detected as empty")
	}
	if !isEmptyRow([]string{}) {
		t.Error("zero-length row not detected as empty")
	}
	if isEmptyRow([]string{"", "x"}) {
		t.Error("row with one non-empty cell incorrectly detected as empty")
	}
}

// TestBOMEditFormFields is an end-to-end check: a BOM-prefixed CSV must still
// populate the edit form fields correctly (regression for the original bug report).
func TestBOMEditFormFields(t *testing.T) {
	bom := "\xef\xbb\xbf"
	content := bom + "name,value\nalpha,1\nbeta,2\n"
	_, csvPath := writeCSVFile(t, content)
	h := NewHandler(config.DataItem{Name: "test", Path: csvPath}, newTestRenderer(t))

	req := httptest.NewRequest(http.MethodGet, "/data/test/?action=edit&row=1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "alpha") {
		t.Errorf("edit form missing field value 'alpha' (BOM likely corrupted header lookup):\n%s", body)
	}
}
