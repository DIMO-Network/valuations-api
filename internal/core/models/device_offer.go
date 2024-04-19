package models

import (
	"reflect"
	"strings"

	"github.com/tidwall/gjson"
)

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
	// Whether estimated, real, or from the market (eg. vincario). Market, Estimated, Real
	OdometerMeasurementType OdometerMeasurementEnum `json:"odometerMeasurementType"`
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

func DecodeOfferFromJSON(drivlyJSON []byte) OfferSet {
	drivlyOffers := OfferSet{}
	drivlyOffers.Source = "drivly"

	// Drivly Offers
	gjson.GetBytes(drivlyJSON, `@keys.#(%"*Price")#`).ForEach(func(_, value gjson.Result) bool {
		offer := Offer{}
		offer.Vendor = strings.TrimSuffix(value.String(), "Price") // eg. vroom, carvana, or carmax
		gjson.GetBytes(drivlyJSON, `@keys.#(%"`+offer.Vendor+`*")#`).ForEach(func(_, value gjson.Result) bool {
			prop := strings.TrimPrefix(value.String(), offer.Vendor)
			if prop == "Url" {
				prop = "URL"
			}
			if !reflect.ValueOf(&offer).Elem().FieldByName(prop).CanSet() {
				return true
			}
			val := gjson.GetBytes(drivlyJSON, value.String())
			switch val.Type {
			case gjson.Null: // ignore null values
				return true
			case gjson.Number: // for "Price"
				reflect.ValueOf(&offer).Elem().FieldByName(prop).Set(reflect.ValueOf(int(val.Int())))
			case gjson.JSON: // for "Error"
				if prop == "Error" {
					val = gjson.GetBytes(drivlyJSON, value.String()+".error.title")
					if val.Exists() {
						offer.Error = val.String()
					}
				}
			default: // for everything else
				reflect.ValueOf(&offer).Elem().FieldByName(prop).Set(reflect.ValueOf(val.String()))
			}
			return true
		})
		drivlyOffers.Offers = append(drivlyOffers.Offers, offer)
		return true
	})

	return drivlyOffers
}
