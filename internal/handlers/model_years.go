package handlers

import (
	"net/http"
	"strconv"

	"carpare/internal/db"
	"carpare/internal/models"
)

// modelYearView pairs a ModelYear with its nameplate for display.
type modelYearView struct {
	models.ModelYear
	Make  string
	Model string
}

type modelYearsPageData struct {
	ModelYears []modelYearView
	CarModels  []models.CarModel
}

type modelYearEditPageData struct {
	models.ModelYear
	Label     string
	CarModels []models.CarModel
}

func (h *Handlers) ListModelYears(w http.ResponseWriter, r *http.Request) {
	carModels, err := h.allCarModels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	modelYears, err := h.allModelYearViews(carModels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.render(w, "model_years.html", modelYearsPageData{ModelYears: modelYears, CarModels: carModels})
}

func (h *Handlers) CreateModelYear(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	my, err := modelYearFromForm(r, db.NewID())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.Save(collModelYears, my.ID, my); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view, err := h.modelYearView(my)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderFragment(w, "model_years.html", "model_year_row", view)
}

func (h *Handlers) EditModelYear(w http.ResponseWriter, r *http.Request) {
	var my models.ModelYear
	if err := h.store.Get(collModelYears, r.PathValue("id"), &my); err != nil {
		http.NotFound(w, r)
		return
	}
	carModels, err := h.allCarModels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	label := strconv.Itoa(my.Year)
	for _, cm := range carModels {
		if cm.ID == my.CarModelID {
			label = strconv.Itoa(my.Year) + " " + cm.Make + " " + cm.Model
			break
		}
	}
	h.render(w, "model_year_edit.html", modelYearEditPageData{ModelYear: my, Label: label, CarModels: carModels})
}

func (h *Handlers) UpdateModelYear(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	my, err := modelYearFromForm(r, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.Save(collModelYears, my.ID, my); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/model-years", http.StatusSeeOther)
}

func (h *Handlers) DeleteModelYear(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(collModelYears, r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func modelYearFromForm(r *http.Request, id string) (models.ModelYear, error) {
	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		return models.ModelYear{}, err
	}
	mpgCity, _ := strconv.ParseFloat(r.FormValue("mpg_city"), 64)
	mpgHighway, _ := strconv.ParseFloat(r.FormValue("mpg_highway"), 64)
	safety, _ := strconv.ParseFloat(r.FormValue("safety_rating"), 64)
	reliability, _ := strconv.ParseFloat(r.FormValue("reliability_rating"), 64)
	horsepower, _ := strconv.Atoi(r.FormValue("horsepower"))
	seating, _ := strconv.Atoi(r.FormValue("seating_capacity"))
	cargo, _ := strconv.ParseFloat(r.FormValue("cargo_cubic_feet"), 64)

	return models.ModelYear{
		ID:                id,
		CarModelID:        r.FormValue("car_model_id"),
		Year:              year,
		MPGCity:           mpgCity,
		MPGHighway:        mpgHighway,
		SafetyRating:      safety,
		ReliabilityRating: reliability,
		Horsepower:        horsepower,
		DriveType:         models.DriveType(r.FormValue("drive_type")),
		FuelType:          models.FuelType(r.FormValue("fuel_type")),
		SeatingCapacity:   seating,
		CargoCubicFeet:    cargo,
		ImageURL:          r.FormValue("image_url"),
		Notes:             r.FormValue("notes"),
	}, nil
}

// modelYearView loads the ModelYear's nameplate for display. If the
// nameplate has been deleted, it renders with a blank Make/Model rather
// than failing.
func (h *Handlers) modelYearView(my models.ModelYear) (modelYearView, error) {
	var cm models.CarModel
	_ = h.store.Get(collCarModels, my.CarModelID, &cm)
	return modelYearView{ModelYear: my, Make: cm.Make, Model: cm.Model}, nil
}

func (h *Handlers) allModelYearViews(carModels []models.CarModel) ([]modelYearView, error) {
	byID := make(map[string]models.CarModel, len(carModels))
	for _, cm := range carModels {
		byID[cm.ID] = cm
	}
	var out []modelYearView
	err := db.All(h.store, collModelYears, func(id string, doc models.ModelYear) error {
		cm := byID[doc.CarModelID]
		out = append(out, modelYearView{ModelYear: doc, Make: cm.Make, Model: cm.Model})
		return nil
	})
	return out, err
}
