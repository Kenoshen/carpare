package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"carpare/internal/db"
	"carpare/internal/models"
)

// modelYearOption is a <select> entry for choosing a ModelYear, labeled
// with its year and nameplate, e.g. "2022 Honda CR-V".
type modelYearOption struct {
	ID    string
	Label string
}

// listingView pairs a Listing with its ModelYear's display label.
type listingView struct {
	models.Listing
	Label string
}

type listingsPageData struct {
	Listings   []listingView
	ModelYears []modelYearOption
}

type listingEditPageData struct {
	models.Listing
	Label      string
	ProsText   string
	ConsText   string
	ModelYears []modelYearOption
}

func (h *Handlers) ListListings(w http.ResponseWriter, r *http.Request) {
	options, err := h.allModelYearOptions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	listings, err := h.allListingViews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.render(w, "listings.html", listingsPageData{Listings: listings, ModelYears: options})
}

func (h *Handlers) CreateListing(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	now := time.Now()
	listing, err := listingFromForm(r, db.NewID(), now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	listing.CreatedAt = now
	if err := h.store.Save(collListings, listing.ID, listing); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view := listingView{Listing: listing, Label: h.modelYearLabel(listing.ModelYearID)}
	h.renderFragment(w, "listings.html", "listing_row", view)
}

func (h *Handlers) EditListing(w http.ResponseWriter, r *http.Request) {
	var listing models.Listing
	if err := h.store.Get(collListings, r.PathValue("id"), &listing); err != nil {
		http.NotFound(w, r)
		return
	}
	options, err := h.allModelYearOptions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.render(w, "listing_edit.html", listingEditPageData{
		Listing:    listing,
		Label:      h.modelYearLabel(listing.ModelYearID),
		ProsText:   strings.Join(listing.Pros, "\n"),
		ConsText:   strings.Join(listing.Cons, "\n"),
		ModelYears: options,
	})
}

func (h *Handlers) UpdateListing(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var existing models.Listing
	if err := h.store.Get(collListings, id, &existing); err != nil {
		http.NotFound(w, r)
		return
	}
	listing, err := listingFromForm(r, id, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	listing.CreatedAt = existing.CreatedAt
	if err := h.store.Save(collListings, listing.ID, listing); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/listings", http.StatusSeeOther)
}

func (h *Handlers) DeleteListing(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(collListings, r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func listingFromForm(r *http.Request, id string, now time.Time) (models.Listing, error) {
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		return models.Listing{}, err
	}
	mileage, err := strconv.Atoi(r.FormValue("mileage"))
	if err != nil {
		return models.Listing{}, err
	}
	return models.Listing{
		ID:            id,
		ModelYearID:   r.FormValue("model_year_id"),
		Trim:          r.FormValue("trim"),
		VIN:           r.FormValue("vin"),
		Price:         price,
		Mileage:       mileage,
		Condition:     models.Condition(r.FormValue("condition")),
		ExteriorColor: r.FormValue("exterior_color"),
		InteriorColor: r.FormValue("interior_color"),
		SellerName:    r.FormValue("seller_name"),
		SellerType:    models.SellerType(r.FormValue("seller_type")),
		Location:      r.FormValue("location"),
		ListingURL:    r.FormValue("listing_url"),
		Status:        models.Status(r.FormValue("status")),
		Pros:          splitLines(r.FormValue("pros")),
		Cons:          splitLines(r.FormValue("cons")),
		Notes:         r.FormValue("notes"),
		UpdatedAt:     now,
	}, nil
}

func splitLines(s string) []string {
	var out []string
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// modelYearLabel returns a display label like "2022 Honda CR-V" for a
// ModelYear id, degrading gracefully if the referenced records are gone.
func (h *Handlers) modelYearLabel(modelYearID string) string {
	var my models.ModelYear
	if err := h.store.Get(collModelYears, modelYearID, &my); err != nil {
		return "(unknown)"
	}
	var cm models.CarModel
	if err := h.store.Get(collCarModels, my.CarModelID, &cm); err != nil {
		return strconv.Itoa(my.Year)
	}
	return strconv.Itoa(my.Year) + " " + cm.Make + " " + cm.Model
}

func (h *Handlers) allModelYearOptions() ([]modelYearOption, error) {
	carModels, err := h.allCarModels()
	if err != nil {
		return nil, err
	}
	byID := make(map[string]models.CarModel, len(carModels))
	for _, cm := range carModels {
		byID[cm.ID] = cm
	}
	var out []modelYearOption
	err = db.All(h.store, collModelYears, func(id string, doc models.ModelYear) error {
		cm := byID[doc.CarModelID]
		out = append(out, modelYearOption{
			ID:    doc.ID,
			Label: strconv.Itoa(doc.Year) + " " + cm.Make + " " + cm.Model,
		})
		return nil
	})
	return out, err
}

func (h *Handlers) allListingViews() ([]listingView, error) {
	var out []listingView
	err := db.All(h.store, collListings, func(id string, doc models.Listing) error {
		out = append(out, listingView{Listing: doc, Label: h.modelYearLabel(doc.ModelYearID)})
		return nil
	})
	return out, err
}
