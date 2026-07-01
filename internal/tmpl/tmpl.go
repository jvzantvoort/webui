package tmpl

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/jvzantvoort/webui/internal/config"
	"github.com/jvzantvoort/webui/static"
)

// NavItem is one entry in the top navigation bar.
// When Children is non-empty the item is rendered as a hover dropdown.
type NavItem struct {
	Label    string
	URL      string
	Children []NavItem
}

// Nav holds the navigation items split by kind, matching the bar layout:
// data items on the left, content items to their right.
type Nav struct {
	Data    []NavItem
	Content []NavItem
}

// IndexedRow pairs a CSV data row with its position in the file.
// Idx is 1-based (row 0 is the header row).
type IndexedRow struct {
	Idx   int
	Cells []string
}

// FormOption is a single choice inside a <select> field.
type FormOption struct {
	Value    string
	Selected bool
}

// FormField describes one field in a create/edit form.
type FormField struct {
	Name      string
	Label     string
	InputType string // HTML input type ("text","email","number","date") or "select"
	Value     string
	Required  bool
	Readonly  bool
	Options   []FormOption // non-empty → rendered as <select>
}

// PageData is the template data passed to every page.
type PageData struct {
	Nav     Nav
	Title   string
	Body    template.HTML     // content pages: rendered markdown; view pages: custom record template output
	Headers []string          // data list pages: CSV header row
	Rows    []IndexedRow      // data list pages: CSV data rows with index
	Fields  []FormField       // form/view pages: fields with labels and values
	RowIdx  int               // form/view pages: 1-based index of the current row
	PostURL string            // form pages: form action URL
	PrevIdx int               // view pages: row index of the previous record (0 = none)
	NextIdx int               // view pages: row index of the next record (0 = none)
	Record  map[string]string // view pages: column→value map for the current record
}

// Renderer parses and executes page templates, injecting shared nav data.
// It uses a clone-per-page strategy so that each page template can properly
// override the {{block}} definitions in base.html without interfering with
// other pages.
type Renderer struct {
	pages map[string]*template.Template
	nav   Nav
}

var pageNames = []string{"index.html", "content.html", "data.html", "form.html", "view.html"}

// New creates a Renderer from the embedded template FS and the app config.
func New(cfg *config.Config) (*Renderer, error) {
	base, err := template.ParseFS(static.FS, "templates/base.html")
	if err != nil {
		return nil, fmt.Errorf("parse base template: %w", err)
	}

	pages := make(map[string]*template.Template, len(pageNames))
	for _, name := range pageNames {
		t, err := template.Must(base.Clone()).ParseFS(static.FS, "templates/"+name)
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", name, err)
		}
		pages[name] = t
	}

	r := &Renderer{pages: pages}
	for _, item := range cfg.Data {
		r.nav.Data = append(r.nav.Data, NavItem{
			Label: MenuLabel(item.Menu, item.Name),
			URL:   "/data/" + item.Name + "/",
		})
	}
	for _, item := range cfg.Content {
		r.nav.Content = append(r.nav.Content, NavItem{
			Label: MenuLabel(item.Menu, item.Name),
			URL:   "/" + item.Name + "/",
		})
	}
	return r, nil
}

// StaticNav returns a shallow copy of the renderer's static navigation.
// Callers that need to add per-request children should start from this copy.
func (r *Renderer) StaticNav() Nav {
	nav := Nav{
		Data:    make([]NavItem, len(r.nav.Data)),
		Content: make([]NavItem, len(r.nav.Content)),
	}
	copy(nav.Data, r.nav.Data)
	copy(nav.Content, r.nav.Content)
	return nav
}

// Render executes the named page template using the renderer's static nav.
func (r *Renderer) Render(w http.ResponseWriter, name string, d PageData) {
	r.renderWithNav(w, name, r.nav, d)
}

// RenderWithNav executes the named page template with a custom nav.
// Use this when the nav needs per-request augmentation (e.g. directory children).
func (r *Renderer) RenderWithNav(w http.ResponseWriter, name string, nav Nav, d PageData) {
	r.renderWithNav(w, name, nav, d)
}

func (r *Renderer) renderWithNav(w http.ResponseWriter, name string, nav Nav, d PageData) {
	t, ok := r.pages[name]
	if !ok {
		http.Error(w, "unknown template: "+name, http.StatusInternalServerError)
		return
	}
	d.Nav = nav
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "base.html", d); err != nil {
		fmt.Printf("template error: %v\n", err)
	}
}

// MenuLabel extracts a display label from a menu string.
// If the string contains "/" the last segment is used (e.g. "company/groups" → "groups").
// Surrounding quotes and whitespace are stripped.
// Falls back to the item name when the result is empty.
func MenuLabel(menu, name string) string {
	if menu == "" {
		return name
	}
	parts := strings.Split(menu, "/")
	label := strings.Trim(strings.TrimSpace(parts[len(parts)-1]), `"`)
	if label == "" {
		return name
	}
	return label
}
