package handlers

import (
	"cmp"
	"net/http"
	"slices"
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
	Year  int
	Make  string
	Model string
	Label string
}

// listingView pairs a Listing with its ModelYear's display label and the
// make/model/year used to sort it alongside other listings.
type listingView struct {
	models.Listing
	Label string
	Make  string
	Model string
	Year  int
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
	h.renderSortedListingRows(w)
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
		ImageURL:      r.FormValue("image_url"),
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

// modelYearInfo is the nameplate/year info needed to label and sort a
// ModelYear reference.
type modelYearInfo struct {
	Year  int
	Make  string
	Model string
}

func (info modelYearInfo) label() string {
	if info == (modelYearInfo{}) {
		return "(unknown)"
	}
	return strconv.Itoa(info.Year) + " " + info.Make + " " + info.Model
}

// lookupModelYearInfo loads the nameplate/year for a ModelYear id,
// degrading gracefully (zero value) if the referenced records are gone.
func (h *Handlers) lookupModelYearInfo(modelYearID string) modelYearInfo {
	var my models.ModelYear
	if err := h.store.Get(collModelYears, modelYearID, &my); err != nil {
		return modelYearInfo{}
	}
	var cm models.CarModel
	_ = h.store.Get(collCarModels, my.CarModelID, &cm)
	return modelYearInfo{Year: my.Year, Make: cm.Make, Model: cm.Model}
}

// modelYearLabel returns a display label like "2022 Honda CR-V" for a
// ModelYear id, degrading gracefully if the referenced records are gone.
func (h *Handlers) modelYearLabel(modelYearID string) string {
	return h.lookupModelYearInfo(modelYearID).label()
}

func (h *Handlers) renderSortedListingRows(w http.ResponseWriter) {
	listings, err := h.allListingViews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderFragment(w, "listings.html", "listing_rows", listings)
}

// allModelYearOptions returns every ModelYear as a <select> option,
// sorted by make, then model, then year.
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
			Year:  doc.Year,
			Make:  cm.Make,
			Model: cm.Model,
			Label: strconv.Itoa(doc.Year) + " " + cm.Make + " " + cm.Model,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(out, func(a, b modelYearOption) int {
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

// allListingViews returns every Listing paired with its nameplate label,
// sorted by make, then model, then year, then price.
func (h *Handlers) allListingViews() ([]listingView, error) {
	var out []listingView
	err := db.All(h.store, collListings, func(id string, doc models.Listing) error {
		info := h.lookupModelYearInfo(doc.ModelYearID)
		out = append(out, listingView{
			Listing: doc,
			Label:   info.label(),
			Make:    info.Make,
			Model:   info.Model,
			Year:    info.Year,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(out, func(a, b listingView) int {
		if c := cmp.Compare(a.Make, b.Make); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Model, b.Model); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Year, b.Year); c != 0 {
			return c
		}
		return cmp.Compare(a.Price, b.Price)
	})
	return out, nil
}
