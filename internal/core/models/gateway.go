package models

import "time"

type Manufacturer struct {
	TokenID int    `json:"tokenId"`
	Name    string `json:"name"`
	TableID int    `json:"tableId"`
	Owner   string `json:"owner"`
}

type GraphQLRequest struct {
	Query string `json:"query"`
}

type DeviceDefinition struct {
	Model        string       `json:"model"`
	Year         int          `json:"year"`
	Manufacturer Manufacturer `json:"manufacturer"`
	ImageURI     string       `json:"imageURI"`
	Attributes   []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"attributes"`
}

type Vehicle struct {
	Id         string `json:"id"`
	Definition struct {
		Id    string `json:"id"`
		Make  string `json:"make"`
		Model string `json:"model"`
		Year  int    `json:"year"`
	} `json:"definition"`
	Owner string `json:"owner"`
}

type SignalsLatest struct {
	PowertrainTransmissionTravelledDistance struct {
		Timestamp time.Time `json:"timestamp"`
		Value     float64   `json:"value"`
	} `json:"powertrainTransmissionTravelledDistance"`
	CurrentLocationLatitude struct {
		Timestamp time.Time `json:"timestamp"`
		Value     float64   `json:"value"`
	} `json:"currentLocationLatitude"`
	CurrentLocationLongitude struct {
		Timestamp time.Time `json:"timestamp"`
		Value     float64   `json:"value"`
	} `json:"currentLocationLongitude"`
}

type VinVCLatest struct {
	Vin         string    `json:"vin"`
	RecordedBy  string    `json:"recordedBy"`
	RecordedAt  time.Time `json:"recordedAt"`
	CountryCode string    `json:"countryCode"`
	ValidFrom   time.Time `json:"validFrom"`
	ValidTo     time.Time `json:"validTo"`
}
