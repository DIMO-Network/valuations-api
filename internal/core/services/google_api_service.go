package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DIMO-Network/valuations-api/internal/config"
)

//go:generate mockgen -source google_api_service.go -destination mocks/google_api_service_mock.go
type GoogleGeoAPIService interface {
	GeoDecodeLatLong(lat, lng float64) (*MapsGeocodeResp, error)
}

func NewGoogleGeoAPIService(settings *config.Settings) GoogleGeoAPIService {
	return &googleGeoAPIService{googleAPIKey: settings.GoogleMapsAPIKey}
}

type googleGeoAPIService struct {
	googleAPIKey string
}

func (dda *googleGeoAPIService) GeoDecodeLatLong(lat, lng float64) (*MapsGeocodeResp, error) {
	resp, err := http.Get(fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?latlng=%f,%f&key=%s", lat, lng, dda.googleAPIKey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	var data Result
	_ = json.Unmarshal(buf.Bytes(), &data) //nolint
	if len(data.Results) > 0 {
		r := MapsGeocodeResp{}
		for _, ac := range data.Results[0].AddressComponents {
			for _, t := range ac.Types {
				switch t {
				case "postal_code":
					r.PostalCode = ac.LongName
				case "country":
					r.Country = ac.ShortName
				case "administrative_area_level_1":
					r.AdminAreaLevel1 = ac.ShortName
				case "administrative_area_level_2":
					r.AdminAreaLevel2 = ac.LongName
				case "locality":
					r.Locality = ac.LongName
				case "route":
					r.Route = ac.LongName
				case "street_number":
					r.StreetNumber = ac.LongName
				}
			}
		}
		return &r, nil
	}
	return nil, fmt.Errorf("no results found")
}

type AddressComponents struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

type Result struct {
	Results []struct {
		AddressComponents []AddressComponents `json:"address_components"`
	} `json:"results"`
}

type MapsGeocodeResp struct {
	PostalCode   string
	StreetNumber string
	// street name
	Route string
	// eg Bergenfield
	Locality string
	// eg Bergen County
	AdminAreaLevel2 string
	// eg NJ (state)
	AdminAreaLevel1  string
	Country          string
	FormattedAddress string
}
