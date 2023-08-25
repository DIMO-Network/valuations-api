package models

type DeviceOffer struct {
	// Contains a list of offer sets, one for each source
	OfferSets []OfferSet `json:"offerSets"`
}
type OfferSet struct {
	// The source of the offers (eg. "drivly")
	Source string `json:"source"`
	// The time the offers were pulled
	Updated string `json:"updated,omitempty"`
	// The mileage used for the offers
	Mileage int `json:"mileage,omitempty"`
	// This will be the zip code used (if any) for the offers request regardless if the source uses it
	ZipCode string `json:"zipCode,omitempty"`
	// Contains a list of offers from the source
	Offers []Offer `json:"offers"`
}
type Offer struct {
	// The vendor of the offer (eg. "carmax", "carvana", etc.)
	Vendor string `json:"vendor"`
	// The offer price from the vendor
	Price int `json:"price,omitempty"`
	// The offer URL from the vendor
	URL string `json:"url,omitempty"`
	// An error from the vendor (eg. when the VIN is invalid)
	Error string `json:"error,omitempty"`
	// The grade of the offer from the vendor (eg. "RETAIL")
	Grade string `json:"grade,omitempty"`
	// The reason the offer was declined from the vendor
	DeclineReason string `json:"declineReason,omitempty"`
}
