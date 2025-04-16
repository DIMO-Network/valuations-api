package models

type LocationResponse struct {
	PostalCode  string `json:"postalCode"`
	CountryCode string `json:"countryCode"`
}
