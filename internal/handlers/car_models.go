package handlers

import (
	"net/http"

	"carpare/internal/db"
	"carpare/internal/models"
)

type carModelsPageData struct {
	CarModels []models.CarModel
}

func (h *Handlers) ListCarModels(w http.ResponseWriter, r *http.Request) {
	carModels, err := h.allCarModels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.render(w, "car_models.html", carModelsPageData{CarModels: carModels})
}

func (h *Handlers) CreateCarModel(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cm := models.CarModel{
		ID:    db.NewID(),
		Make:  r.FormValue("make"),
		Model: r.FormValue("model"),
	}
	if err := h.store.Save(collCarModels, cm.ID, cm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderFragment(w, "car_models.html", "car_model_row", cm)
}

func (h *Handlers) ViewCarModel(w http.ResponseWriter, r *http.Request) {
	var cm models.CarModel
	if err := h.store.Get(collCarModels, r.PathValue("id"), &cm); err != nil {
		http.NotFound(w, r)
		return
	}
	h.renderFragment(w, "car_models.html", "car_model_row", cm)
}

func (h *Handlers) EditCarModel(w http.ResponseWriter, r *http.Request) {
	var cm models.CarModel
	if err := h.store.Get(collCarModels, r.PathValue("id"), &cm); err != nil {
		http.NotFound(w, r)
		return
	}
	h.renderFragment(w, "car_models.html", "car_model_form", cm)
}

func (h *Handlers) UpdateCarModel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cm := models.CarModel{
		ID:    id,
		Make:  r.FormValue("make"),
		Model: r.FormValue("model"),
	}
	if err := h.store.Save(collCarModels, cm.ID, cm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderFragment(w, "car_models.html", "car_model_row", cm)
}

func (h *Handlers) DeleteCarModel(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(collCarModels, r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) allCarModels() ([]models.CarModel, error) {
	var out []models.CarModel
	err := db.All(h.store, collCarModels, func(id string, doc models.CarModel) error {
		out = append(out, doc)
		return nil
	})
	return out, err
}
