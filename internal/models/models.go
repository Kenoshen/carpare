// Package models defines the domain types tracked by Carpare.
package models

import "time"

// CarModel is a make/model nameplate, e.g. "Honda" "CR-V", independent
// of any particular year.
type CarModel struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
}

// DriveType is a drivetrain layout.
type DriveType string

const (
	DriveFWD DriveType = "fwd"
	DriveRWD DriveType = "rwd"
	DriveAWD DriveType = "awd"
	Drive4WD DriveType = "4wd"
)

// FuelType is the fuel/power source a ModelYear runs on.
type FuelType string

const (
	FuelGas      FuelType = "gas"
	FuelHybrid   FuelType = "hybrid"
	FuelPHEV     FuelType = "phev"
	FuelElectric FuelType = "electric"
	FuelDiesel   FuelType = "diesel"
)

// ModelYear is one model year of a CarModel, e.g. "2022 Honda CR-V". It
// carries the specs and ratings that hold across that year's lineup,
// regardless of which specific listing you're looking at.
type ModelYear struct {
	ID         string `json:"id"`
	CarModelID string `json:"car_model_id"`
	Year       int    `json:"year"`

	MPGCity           float64   `json:"mpg_city,omitempty"`
	MPGHighway        float64   `json:"mpg_highway,omitempty"`
	SafetyRating      float64   `json:"safety_rating,omitempty"`      // e.g. NHTSA overall stars, 0-5
	ReliabilityRating float64   `json:"reliability_rating,omitempty"` // e.g. JD Power/RepairPal score
	Horsepower        int       `json:"horsepower,omitempty"`
	DriveType         DriveType `json:"drive_type,omitempty"`
	FuelType          FuelType  `json:"fuel_type,omitempty"`
	SeatingCapacity   int       `json:"seating_capacity,omitempty"`
	CargoCubicFeet    float64   `json:"cargo_cubic_feet,omitempty"`
	Notes             string    `json:"notes,omitempty"`
}

// Condition describes a listing's sale condition.
type Condition string

const (
	ConditionNew  Condition = "new"
	ConditionUsed Condition = "used"
	ConditionCPO  Condition = "certified_pre_owned"
)

// SellerType distinguishes who is selling a listing.
type SellerType string

const (
	SellerDealer  SellerType = "dealer"
	SellerPrivate SellerType = "private"
)

// Status tracks where a listing stands in your buying process.
type Status string

const (
	StatusConsidering Status = "considering"
	StatusContacted   Status = "contacted"
	StatusTestDriven  Status = "test_driven"
	StatusRejected    Status = "rejected"
	StatusPurchased   Status = "purchased"
)

// Listing is a specific real-world vehicle you're considering: one
// physical car, for sale somewhere, at a specific price and mileage.
type Listing struct {
	ID          string `json:"id"`
	ModelYearID string `json:"model_year_id"`

	Trim          string    `json:"trim,omitempty"`
	VIN           string    `json:"vin,omitempty"`
	Price         float64   `json:"price"`
	Mileage       int       `json:"mileage"`
	Condition     Condition `json:"condition,omitempty"`
	ExteriorColor string    `json:"exterior_color,omitempty"`
	InteriorColor string    `json:"interior_color,omitempty"`

	SellerName string     `json:"seller_name,omitempty"`
	SellerType SellerType `json:"seller_type,omitempty"`
	Location   string     `json:"location,omitempty"`
	ListingURL string     `json:"listing_url,omitempty"`

	Status Status   `json:"status"`
	Pros   []string `json:"pros,omitempty"`
	Cons   []string `json:"cons,omitempty"`
	Notes  string   `json:"notes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
