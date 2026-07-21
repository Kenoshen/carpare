// Package templates embeds the html/template sources for the Carpare UI.
package templates

import (
	"embed"
	"fmt"
	"html/template"
)

//go:embed *.html
var FS embed.FS

// Load parses layout.html together with each other page template in FS,
// returning one *template.Template per page keyed by filename (e.g.
// "dashboard.html"). Each page gets its own combined template set so
// that pages can each define a "content" block without colliding.
func Load() (map[string]*template.Template, error) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("templates: reading embedded dir: %w", err)
	}

	pages := make(map[string]*template.Template)
	for _, e := range entries {
		name := e.Name()
		if name == "layout.html" {
			continue
		}
		tmpl, err := template.ParseFS(FS, "layout.html", name)
		if err != nil {
			return nil, fmt.Errorf("templates: parsing %q: %w", name, err)
		}
		pages[name] = tmpl
	}
	return pages, nil
}
