package content

import (
	"net/http"
	"net/http/httptest"
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

func TestServeMarkdown(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Hello"), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(config.ContentItem{Name: "help", Path: dir, Content: "markdown"}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/help/index.md", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", ct)
	}
	if !strings.Contains(w.Body.String(), "<h1>") {
		t.Errorf("body missing <h1>: %s", w.Body.String())
	}
}

func TestServeDefaultIndex(t *testing.T) {
	// Default index is README.md when no Index field is set.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Readme"), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(config.ContentItem{Name: "help", Path: dir, Content: "markdown"}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/help/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Readme") {
		t.Errorf("body should contain README.md content: %s", w.Body.String())
	}
}

func TestServeConfiguredIndex(t *testing.T) {
	// When Index is explicitly set, that file is served at the section root.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Configured index"), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(config.ContentItem{Name: "help", Path: dir, Content: "markdown", Index: "index.md"}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/help/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Configured index") {
		t.Errorf("body should contain configured index content: %s", w.Body.String())
	}
}

func TestServeDefaultIndexMissing(t *testing.T) {
	// When the index file is absent the handler returns 404, not a crash.
	dir := t.TempDir()

	h := NewHandler(config.ContentItem{Name: "help", Path: dir, Content: "markdown"}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/help/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestServeMissingFile(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(config.ContentItem{Name: "help", Path: dir, Content: "markdown"}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/help/missing.md", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestServeStaticContent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(config.ContentItem{Name: "docs", Path: dir, Content: "static"}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/docs/file.txt", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "hello" {
		t.Errorf("body = %q, want 'hello'", w.Body.String())
	}
}

// TestDropdownAppearsWithMultipleFiles verifies that when a content folder
// contains multiple files, the rendered page includes dropdown nav links.
func TestDropdownAppearsWithMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"index.md", "installation.md", "usage.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("# "+name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Build a renderer that knows about the "help" content item so the
	// enriched nav can find and annotate the matching entry.
	r, err := tmpl.New(&config.Config{
		Content: []config.ContentItem{
			{Name: "help", Path: dir, Content: "markdown", Menu: "Help"},
		},
	})
	if err != nil {
		t.Fatalf("tmpl.New() error = %v", err)
	}

	h := NewHandler(config.ContentItem{Name: "help", Path: dir, Content: "markdown", Menu: "Help"}, r)
	req := httptest.NewRequest(http.MethodGet, "/help/index.md", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "dropdown-menu") {
		t.Errorf("expected Bootstrap dropdown in nav for multi-file folder, got: %s", body)
	}
	if !strings.Contains(body, "installation") {
		t.Errorf("dropdown missing link for installation.md: %s", body)
	}
	if !strings.Contains(body, "usage") {
		t.Errorf("dropdown missing link for usage.md: %s", body)
	}
}

// TestNoDropdownWithSingleFile verifies that a folder with only one file
// does not produce a dropdown (a plain link suffices).
func TestNoDropdownWithSingleFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Only"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := tmpl.New(&config.Config{
		Content: []config.ContentItem{
			{Name: "help", Path: dir, Content: "markdown", Menu: "Help"},
		},
	})
	if err != nil {
		t.Fatalf("tmpl.New() error = %v", err)
	}

	h := NewHandler(config.ContentItem{Name: "help", Path: dir, Content: "markdown", Menu: "Help"}, r)
	req := httptest.NewRequest(http.MethodGet, "/help/index.md", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if strings.Contains(w.Body.String(), "dropdown-menu") {
		t.Error("expected no dropdown for single-file folder")
	}
}

// ── ignore-pattern tests ──────────────────────────────────────────────────

func TestIgnoredFilesAbsentFromListing(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"guide.md", "notes.md", "photo.jpg", "diagram.gif"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	r, err := tmpl.New(&config.Config{
		Content: []config.ContentItem{
			{Name: "docs", Path: dir, Content: "markdown",
				Ignore: []string{"*.jpg", "*.gif"}, Menu: "Docs"},
		},
	})
	if err != nil {
		t.Fatalf("tmpl.New: %v", err)
	}

	h := NewHandler(config.ContentItem{
		Name: "docs", Path: dir, Content: "markdown",
		Ignore: []string{"*.jpg", "*.gif"}, Menu: "Docs",
	}, r)
	req := httptest.NewRequest(http.MethodGet, "/docs/guide.md", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	if strings.Contains(body, "photo.jpg") {
		t.Errorf("ignored file photo.jpg should not appear in nav listing")
	}
	if strings.Contains(body, "diagram.gif") {
		t.Errorf("ignored file diagram.gif should not appear in nav listing")
	}
	// The non-ignored markdown files must still be present.
	if !strings.Contains(body, "notes") {
		t.Errorf("notes.md should appear in nav listing")
	}
}

func TestIgnoredFilesStillServable(t *testing.T) {
	// Ignore only affects the listing; ignored files can still be fetched directly.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("JFIF"), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(config.ContentItem{
		Name: "docs", Path: dir, Content: "static",
		Ignore: []string{"*.jpg"},
	}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/docs/photo.jpg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ignored file should still be directly servable, got status %d", w.Code)
	}
}

func TestIsIgnored(t *testing.T) {
	h := &Handler{item: config.ContentItem{Ignore: []string{"*.jpg", "draft-*", "README.md"}}}

	cases := []struct {
		name    string
		ignored bool
	}{
		{"photo.jpg", true},
		{"photo.JPG", false}, // filepath.Match is case-sensitive on Linux
		{"draft-notes.md", true},
		{"README.md", true},
		{"guide.md", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := h.isIgnored(tc.name); got != tc.ignored {
			t.Errorf("isIgnored(%q) = %v, want %v", tc.name, got, tc.ignored)
		}
	}
}

// ── GFM extension tests ───────────────────────────────────────────────────

func TestServeMarkdownTable(t *testing.T) {
	dir := t.TempDir()
	md := "| Name  | Value |\n|-------|-------|\n| alpha | 1     |\n| beta  | 2     |\n"
	if err := os.WriteFile(filepath.Join(dir, "table.md"), []byte(md), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(config.ContentItem{Name: "docs", Path: dir, Content: "markdown"}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/docs/table.md", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<table>") {
		t.Errorf("markdown table not rendered as <table>: %s", body)
	}
	if !strings.Contains(body, "<th>") {
		t.Errorf("markdown table missing <th> header cells: %s", body)
	}
}

func TestServeMarkdownStrikethrough(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "s.md"), []byte("~~removed~~"), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(config.ContentItem{Name: "docs", Path: dir, Content: "markdown"}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/docs/s.md", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "<del>") {
		t.Errorf("strikethrough not rendered as <del>: %s", w.Body.String())
	}
}

func TestServeMarkdownTaskList(t *testing.T) {
	dir := t.TempDir()
	md := "- [x] done\n- [ ] todo\n"
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(md), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(config.ContentItem{Name: "docs", Path: dir, Content: "markdown"}, newTestRenderer(t))
	req := httptest.NewRequest(http.MethodGet, "/docs/tasks.md", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `type="checkbox"`) {
		t.Errorf("task list not rendered with checkboxes: %s", body)
	}
	if !strings.Contains(body, "checked") {
		t.Errorf("completed task not rendered as checked: %s", body)
	}
}

func TestFileLabel(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"index.md", "Index"},
		{"getting-started.md", "Getting Started"},
		{"api_reference.md", "Api Reference"},
		{"FAQ.md", "Faq"},
		{"no-extension", "No Extension"},
	}
	for _, tt := range tests {
		got := fileLabel(tt.in)
		if got != tt.want {
			t.Errorf("fileLabel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
