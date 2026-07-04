package content

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"github.com/jvzantvoort/webui/internal/config"
	"github.com/jvzantvoort/webui/internal/tmpl"
)

// Handler serves content items (markdown folders, static files).
type Handler struct {
	item config.ContentItem
	md   goldmark.Markdown
	r    *tmpl.Renderer
}

// NewHandler creates an http.Handler for the given ContentItem.
func NewHandler(item config.ContentItem, r *tmpl.Renderer) http.Handler {
	return &Handler{
		item: item,
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.Table,
				extension.Strikethrough,
				extension.Linkify,
				extension.TaskList,
			),
		),
		r: r,
	}
}

// indexFile returns the filename served when the section root is requested.
func (h *Handler) indexFile() string {
	if h.item.Index != "" {
		return h.item.Index
	}
	return "README.md"
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	prefix := "/" + h.item.Name + "/"
	rel := strings.TrimPrefix(req.URL.Path, prefix)
	if rel == "" {
		rel = h.indexFile()
	}

	fpath := filepath.Join(h.item.Path, filepath.Clean(rel))

	if h.item.ServeImages {
		switch strings.ToLower(filepath.Ext(fpath)) {
		case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
			http.ServeFile(w, req, fpath)
			return
		}
	}

	raw, err := os.ReadFile(fpath)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	if h.item.Content == "markdown" {
		var buf bytes.Buffer
		if err := h.md.Convert(raw, &buf); err != nil {
			http.Error(w, "render error", http.StatusInternalServerError)
			return
		}
		h.r.RenderWithNav(w, "content.html", h.enrichedNav(), tmpl.PageData{
			Title: tmpl.MenuLabel(h.item.Menu, h.item.Name),
			Body:  template.HTML(buf.String()),
		})
		return
	}

	_, _ = w.Write(raw)
}

// enrichedNav returns a copy of the static nav with the file list for this
// content section injected as children on the matching nav item.
// A dropdown is only added when there are two or more files to choose from.
func (h *Handler) enrichedNav() tmpl.Nav {
	nav := h.r.StaticNav()
	children := h.listFiles()
	if len(children) < 2 {
		return nav
	}
	target := "/" + h.item.Name + "/"
	for i, item := range nav.Content {
		if item.URL == target {
			nav.Content[i].Children = children
			break
		}
	}
	return nav
}

// isIgnored reports whether name matches any of the configured ignore patterns.
// Patterns follow filepath.Match syntax (e.g. "*.jpg", "draft-*").
// A malformed pattern is silently skipped.
func (h *Handler) isIgnored(name string) bool {
	for _, pattern := range h.item.Ignore {
		if matched, err := filepath.Match(pattern, name); err == nil && matched {
			return true
		}
	}
	return false
}

// listFiles scans the content directory and returns a NavItem for each
// servable file that is not hidden and does not match an ignore pattern,
// sorted alphabetically.
func (h *Handler) listFiles() []tmpl.NavItem {
	entries, err := os.ReadDir(h.item.Path)
	if err != nil {
		return nil
	}

	base := "/" + h.item.Name + "/"
	var items []tmpl.NavItem
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") || h.isIgnored(e.Name()) {
			continue
		}
		items = append(items, tmpl.NavItem{
			Label: fileLabel(e.Name()),
			URL:   base + e.Name(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Label < items[j].Label
	})
	return items
}

// fileLabel converts a filename into a readable label:
// strips the extension, replaces hyphens/underscores with spaces, and
// capitalises the first letter of each word.
func fileLabel(filename string) string {
	name := filename
	if i := strings.LastIndexByte(name, '.'); i > 0 {
		name = name[:i]
	}
	name = strings.NewReplacer("-", " ", "_", " ").Replace(name)
	words := strings.Fields(name)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}
