// Package handlers implements the standard http.Handlers backing the
// Carpare web UI.
package handlers

import (
	"html/template"
	"net/http"

	"carpare/internal/db"
)

const (
	collCarModels  = "car_models"
	collModelYears = "model_years"
	collListings   = "listings"
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

	mux.HandleFunc("GET /car-models", h.ListCarModels)
	mux.HandleFunc("POST /car-models", h.CreateCarModel)
	mux.HandleFunc("GET /car-models/{id}", h.ViewCarModel)
	mux.HandleFunc("GET /car-models/{id}/edit", h.EditCarModel)
	mux.HandleFunc("PUT /car-models/{id}", h.UpdateCarModel)
	mux.HandleFunc("DELETE /car-models/{id}", h.DeleteCarModel)

	mux.HandleFunc("GET /model-years", h.ListModelYears)
	mux.HandleFunc("POST /model-years", h.CreateModelYear)
	mux.HandleFunc("GET /model-years/{id}/edit", h.EditModelYear)
	mux.HandleFunc("POST /model-years/{id}", h.UpdateModelYear)
	mux.HandleFunc("DELETE /model-years/{id}", h.DeleteModelYear)

	mux.HandleFunc("GET /listings", h.ListListings)
	mux.HandleFunc("POST /listings", h.CreateListing)
	mux.HandleFunc("GET /listings/{id}/edit", h.EditListing)
	mux.HandleFunc("POST /listings/{id}", h.UpdateListing)
	mux.HandleFunc("DELETE /listings/{id}", h.DeleteListing)

	return mux
}

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.render(w, "dashboard.html", nil)
}

// render executes a page's "layout" template, producing a full HTML
// document.
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

// renderFragment executes a single named block from page (without the
// surrounding layout), for htmx partial swaps.
func (h *Handlers) renderFragment(w http.ResponseWriter, page, block string, data any) {
	tmpl, ok := h.pages[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, block, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
