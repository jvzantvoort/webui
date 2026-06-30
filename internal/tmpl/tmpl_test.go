package tmpl

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jvzantvoort/webui/internal/config"
)

func TestMenuLabel(t *testing.T) {
	tests := []struct {
		menu, name, want string
	}{
		{"", "fallback", "fallback"},
		{"Help", "help", "Help"},
		{`"Help"`, "help", "Help"},
		{"company/groups", "grp", "groups"},
		{"company / groups", "grp", "groups"},
		{`"company" / "groups"`, "grp", "groups"},
		{"/", "x", "x"}, // both segments empty → fallback
	}
	for _, tt := range tests {
		got := MenuLabel(tt.menu, tt.name)
		if got != tt.want {
			t.Errorf("MenuLabel(%q, %q) = %q, want %q", tt.menu, tt.name, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Data: []config.DataItem{
			{Name: "groups", Menu: "company/groups"},
		},
		Content: []config.ContentItem{
			{Name: "help", Menu: "Help"},
		},
	}
	r, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if len(r.nav.Data) != 1 || r.nav.Data[0].Label != "groups" {
		t.Errorf("nav.Data = %v, want [{groups /data/groups/}]", r.nav.Data)
	}
	if len(r.nav.Content) != 1 || r.nav.Content[0].Label != "Help" {
		t.Errorf("nav.Content = %v, want [{Help /help/}]", r.nav.Content)
	}
	if r.nav.Data[0].URL != "/data/groups/" {
		t.Errorf("data URL = %q, want /data/groups/", r.nav.Data[0].URL)
	}
}

func TestRenderIndex(t *testing.T) {
	r, err := New(&config.Config{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	w := httptest.NewRecorder()
	r.Render(w, "index.html", PageData{Title: "webui"})

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<html") {
		t.Errorf("response missing <html>: %s", body)
	}
}

func TestRenderUnknownTemplate(t *testing.T) {
	r, err := New(&config.Config{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	w := httptest.NewRecorder()
	r.Render(w, "nonexistent.html", PageData{})

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
