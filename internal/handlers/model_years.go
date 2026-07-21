package handlers

import (
	"cmp"
	"net/http"
	"slices"
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
	h.renderSortedModelYearRows(w)
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

func (h *Handlers) renderSortedModelYearRows(w http.ResponseWriter) {
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
	h.renderFragment(w, "model_years.html", "model_year_rows", modelYears)
}

func modelYearFromForm(r *http.Request, id string) (models.ModelYear, error) {
	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		return models.ModelYear{}, err
	}
	avgMPG, _ := strconv.ParseFloat(r.FormValue("average_mpg"), 64)
	rating, _ := strconv.Atoi(r.FormValue("rating"))
	minPrice, _ := strconv.ParseFloat(r.FormValue("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(r.FormValue("max_price"), 64)
	engineCylinders, _ := strconv.Atoi(r.FormValue("engine_cylinders"))
	engineLiters, _ := strconv.ParseFloat(r.FormValue("engine_liters"), 64)
	seating, _ := strconv.Atoi(r.FormValue("seating_capacity"))

	return models.ModelYear{
		ID:              id,
		CarModelID:      r.FormValue("car_model_id"),
		Year:            year,
		AverageMPG:      avgMPG,
		Rating:          rating,
		MinPrice:        minPrice,
		MaxPrice:        maxPrice,
		EngineCylinders: engineCylinders,
		EngineLiters:    engineLiters,
		DriveType:       models.DriveType(r.FormValue("drive_type")),
		FuelType:        models.FuelType(r.FormValue("fuel_type")),
		SeatingCapacity: seating,
		ImageURL:        r.FormValue("image_url"),
		Notes:           r.FormValue("notes"),
	}, nil
}

// allModelYearViews returns every ModelYear paired with its nameplate,
// sorted by make, then model, then year.
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
	if err != nil {
		return nil, err
	}
	slices.SortFunc(out, func(a, b modelYearView) int {
		if c := cmp.Compare(a.Make, b.Make); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Model, b.Model); c != 0 {
			return c
		}
		return cmp.Compare(a.Year, b.Year)
	})
	return out, nil
}
