package controllers

import (
	"database/sql"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/controllers/helpers"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ValuationsController struct {
	dbs               func() *db.ReaderWriter
	log               *zerolog.Logger
	userDeviceService services.UserDeviceAPIService
}

func NewValuationsController(log *zerolog.Logger, dbs func() *db.ReaderWriter) *ValuationsController {
	return &ValuationsController{
		log: log,
		dbs: dbs,
	}
}

func (vc *ValuationsController) GetValuations(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")

	logger := helpers.GetLogger(c, vc.log).With().Str("route", c.Route().Path).Logger()

	dVal := DeviceValuation{
		ValuationSets: []ValuationSet{},
	}

	// Drivly data
	valuationData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.UserDeviceID.EQ(null.StringFrom(udi)),
		qm.Where("pricing_metadata is not null or vincario_metadata is not null"),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(c.Context(), vc.dbs().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if valuationData != nil {
		if valuationData.PricingMetadata.Valid {
			drivlyVal := ValuationSet{
				Vendor:        "drivly",
				TradeInSource: "drivly",
				RetailSource:  "drivly",
				Updated:       valuationData.UpdatedAt.Format(time.RFC3339),
			}
			drivlyJSON := valuationData.PricingMetadata.JSON
			requestJSON := valuationData.RequestMetadata.JSON
			drivlyMileage := gjson.GetBytes(drivlyJSON, "mileage")
			if drivlyMileage.Exists() {
				drivlyVal.Mileage = int(drivlyMileage.Int())
				drivlyVal.Odometer = int(drivlyMileage.Int())
				drivlyVal.OdometerUnit = "miles"
			} else {
				requestMileage := gjson.GetBytes(requestJSON, "mileage")
				if requestMileage.Exists() {
					drivlyVal.Mileage = int(requestMileage.Int())
				}
			}
			requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
			if requestZipCode.Exists() {
				drivlyVal.ZipCode = requestZipCode.String()
			}
			// Drivly Trade-In
			drivlyVal.TradeIn = extractDrivlyValuation(drivlyJSON, "trade")
			drivlyVal.TradeInAverage = drivlyVal.TradeIn
			// Drivly Retail
			drivlyVal.Retail = extractDrivlyValuation(drivlyJSON, "retail")
			drivlyVal.RetailAverage = drivlyVal.Retail
			drivlyVal.Currency = "USD"

			// often drivly saves valuations with 0 for value, if this is case do not consider it
			if drivlyVal.Retail > 0 || drivlyVal.TradeIn > 0 {
				// set the price to display to users
				drivlyVal.UserDisplayPrice = (drivlyVal.Retail + drivlyVal.TradeIn) / 2
				dVal.ValuationSets = append(dVal.ValuationSets, drivlyVal)
			} else {
				logger.Warn().Msg("did not find a drivly trade-in or retail value, or json in unexpected format")
			}
		} else if valuationData.VincarioMetadata.Valid {
			vincarioVal := ValuationSet{
				Vendor:        "vincario",
				TradeInSource: "vincario",
				RetailSource:  "vincario",
				Updated:       valuationData.UpdatedAt.Format(time.RFC3339),
			}
			valJSON := valuationData.VincarioMetadata.JSON
			requestJSON := valuationData.RequestMetadata.JSON
			odometerMarket := gjson.GetBytes(valJSON, "market_odometer.odometer_avg")
			if odometerMarket.Exists() {
				vincarioVal.Mileage = int(odometerMarket.Int())
				vincarioVal.Odometer = int(odometerMarket.Int())
				vincarioVal.OdometerUnit = gjson.GetBytes(valJSON, "market_odometer.odometer_unit").String()
			}
			// todo this needs to be implemented in the load_valuations script
			requestPostalCode := gjson.GetBytes(requestJSON, "postalCode")
			if requestPostalCode.Exists() {
				vincarioVal.ZipCode = requestPostalCode.String()
			}
			// vincario Trade-In - just using the price below mkt mean
			vincarioVal.TradeIn = int(gjson.GetBytes(valJSON, "market_price.price_below").Int())
			vincarioVal.TradeInAverage = vincarioVal.TradeIn
			// vincario Retail - just using the price above mkt mean
			vincarioVal.Retail = int(gjson.GetBytes(valJSON, "market_price.price_above").Int())
			vincarioVal.RetailAverage = vincarioVal.Retail

			vincarioVal.UserDisplayPrice = int(gjson.GetBytes(valJSON, "market_price.price_avg").Int())
			vincarioVal.Currency = gjson.GetBytes(valJSON, "market_price.price_currency").String()

			// often drivly saves valuations with 0 for value, if this is case do not consider it
			if vincarioVal.Retail > 0 || vincarioVal.TradeIn > 0 {
				dVal.ValuationSets = append(dVal.ValuationSets, vincarioVal)
			} else {
				logger.Warn().Msg("did not find a market value from vincario, or valJSON in unexpected format")
			}

		}
	}

	return c.JSON(dVal)
}

func (vc *ValuationsController) GetOffers(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")

	dOffer := DeviceOffer{
		OfferSets: []OfferSet{},
	}

	// Drivly data
	drivlyVinData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.UserDeviceID.EQ(null.StringFrom(udi)),
		models.ExternalVinDatumWhere.OfferMetadata.IsNotNull(), // offer_metadata is sourced from drivly
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(c.Context(), vc.dbs().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if drivlyVinData != nil {
		drivlyOffers := OfferSet{}
		drivlyOffers.Source = "drivly"
		drivlyJSON := drivlyVinData.OfferMetadata.JSON
		requestJSON := drivlyVinData.RequestMetadata.JSON
		drivlyOffers.Updated = drivlyVinData.UpdatedAt.Format(time.RFC3339)
		requestMileage := gjson.GetBytes(requestJSON, "mileage")
		if requestMileage.Exists() {
			drivlyOffers.Mileage = int(requestMileage.Int())
		}
		requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
		if requestZipCode.Exists() {
			drivlyOffers.ZipCode = requestZipCode.String()
		}
		// Drivly Offers
		gjson.GetBytes(drivlyJSON, `@keys.#(%"*Price")#`).ForEach(func(key, value gjson.Result) bool {
			offer := Offer{}
			offer.Vendor = strings.TrimSuffix(value.String(), "Price") // eg. vroom, carvana, or carmax
			gjson.GetBytes(drivlyJSON, `@keys.#(%"`+offer.Vendor+`*")#`).ForEach(func(key, value gjson.Result) bool {
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
							// reflect.ValueOf(&offer).Elem().FieldByName(prop).Set(reflect.ValueOf(val.String()))
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
		dOffer.OfferSets = append(dOffer.OfferSets, drivlyOffers)
	}

	return c.JSON(dOffer)
}

type DeviceValuation struct {
	// Contains a list of valuation sets, one for each vendor
	ValuationSets []ValuationSet `json:"valuationSets"`
}
type ValuationSet struct {
	// The source of the valuation (eg. "drivly" or "blackbook")
	Vendor string `json:"vendor"`
	// The time the valuation was pulled or in the case of blackbook, this may be the event time of the device odometer which was used for the valuation
	Updated string `json:"updated,omitempty"`
	// The mileage used for the valuation
	Mileage int `json:"mileage,omitempty"`
	// This will be the zip code used (if any) for the valuation request regardless if the vendor uses it
	ZipCode string `json:"zipCode,omitempty"`
	// Useful when Drivly returns multiple vendors and we've selected one (eg. "drivly:blackbook")
	TradeInSource string `json:"tradeInSource,omitempty"`
	// tradeIn is equal to tradeInAverage when available
	TradeIn int `json:"tradeIn,omitempty"`
	// tradeInClean, tradeInAverage, and tradeInRough my not always be available
	TradeInClean   int `json:"tradeInClean,omitempty"`
	TradeInAverage int `json:"tradeInAverage,omitempty"`
	TradeInRough   int `json:"tradeInRough,omitempty"`
	// Useful when Drivly returns multiple vendors and we've selected one (eg. "drivly:blackbook")
	RetailSource string `json:"retailSource,omitempty"`
	// retail is equal to retailAverage when available
	Retail int `json:"retail,omitempty"`
	// retailClean, retailAverage, and retailRough my not always be available
	RetailClean   int    `json:"retailClean,omitempty"`
	RetailAverage int    `json:"retailAverage,omitempty"`
	RetailRough   int    `json:"retailRough,omitempty"`
	OdometerUnit  string `json:"odometerUnit"`
	Odometer      int    `json:"odometer"`
	// UserDisplayPrice the top level value to show to users in mobile app
	UserDisplayPrice int `json:"userDisplayPrice"`
	// eg. USD or EUR
	Currency string `json:"currency"`
}

// extractDrivlyValuation pulls out the price from the drivly json, based on the passed in key, eg. trade or retail. calculates average if no root property found
func extractDrivlyValuation(drivlyJSON []byte, key string) int {
	if gjson.GetBytes(drivlyJSON, key).Exists() && !gjson.GetBytes(drivlyJSON, key).IsObject() {
		v := gjson.GetBytes(drivlyJSON, key).String()
		vf, _ := strconv.ParseFloat(v, 64)
		return int(vf)
	}
	// get all values
	pricings := map[string]int{}
	if gjson.GetBytes(drivlyJSON, key+".blackBook.totalAvg").Exists() {
		values := gjson.GetManyBytes(drivlyJSON, key+".blackBook.totalRough", key+".blackBook.totalAvg", key+".blackBook.totalClean")
		pricings["blackbook"] = int(values[1].Int())
	}
	if gjson.GetBytes(drivlyJSON, key+".kelley.good").Exists() {
		pricings["kbb"] = int(gjson.GetBytes(drivlyJSON, key+".kelley.good").Int())
	}
	if gjson.GetBytes(drivlyJSON, key+".edmunds.average").Exists() {
		values := gjson.GetManyBytes(drivlyJSON, key+".edmunds.rough", key+".edmunds.average", key+".edmunds.clean")
		pricings["edmunds"] = int(values[1].Int())
	}
	if len(pricings) > 1 {
		sum := 0
		for _, v := range pricings {
			sum += v
		}
		return sum / len(pricings)
	}

	return 0
}

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
