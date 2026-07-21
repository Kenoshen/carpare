// Package handlers implements the standard http.Handlers backing the
// Carpare web UI.
package handlers

import (
	"html/template"
	"net/http"

	"carpare/internal/db"
)

type Handlers struct {
	pages map[string]*template.Template
	store *db.Store
}

func New(pages map[string]*template.Template, store *db.Store) *Handlers {
	return &Handlers{pages: pages, store: store}
}

func (h *Handlers) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.Dashboard)
	return mux
}

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.render(w, "dashboard.html", nil)
}

func (h *Handlers) render(w http.ResponseWriter, page string, data any) {
	tmpl, ok := h.pages[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
