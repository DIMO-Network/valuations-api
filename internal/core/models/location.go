package models

type LocationResponse struct {
	// PostalCode is the postal code of the location or ZIP Code in the USA
	PostalCode string `json:"postalCode"`
	// CountryCode is the ISO 3166-1 alpha-2 country code
	CountryCode string `json:"countryCode"`
}
