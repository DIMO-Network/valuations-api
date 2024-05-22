package services

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/valuations-api/internal/config"

	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source vincario_api_service.go -destination mocks/vincario_api_service_mock.go -package mock_services
type VincarioAPIService interface {
	GetMarketValuation(vin string) (*VincarioMarketValueResponse, error)
}

type vincarioAPIService struct {
	settings      *config.Settings
	httpClientVIN shared.HTTPClientWrapper
	log           *zerolog.Logger
}

func NewVincarioAPIService(settings *config.Settings, log *zerolog.Logger) VincarioAPIService {
	if settings.VincarioAPIURL == "" || settings.VincarioAPISecret == "" {
		panic("Vincario configuration not set")
	}
	hcwv, _ := shared.NewHTTPClientWrapper(settings.VincarioAPIURL, "", 10*time.Second, nil, false)

	return &vincarioAPIService{
		settings:      settings,
		httpClientVIN: hcwv,
		log:           log,
	}
}

func (va *vincarioAPIService) GetMarketValuation(vin string) (*VincarioMarketValueResponse, error) {
	id := "vehicle-market-value"

	urlPath := vincarioPathBuilder(vin, id, va.settings.VincarioAPIKey, va.settings.VincarioAPISecret)
	// url with api access
	resp, err := va.httpClientVIN.ExecuteRequest(urlPath, "GET", nil)
	if err != nil {
		return nil, err
	}

	// decode JSON from response body
	var data VincarioMarketValueResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if data.MarketPrice.Europe.PriceAvg == 0 && data.MarketPrice.NorthAmerica.PriceAvg == 0 {
		return nil, fmt.Errorf("invalid valuation with 0 value returned - %+v", data)
	}

	return &data, nil
}

func vincarioPathBuilder(vin, id, key, secret string) string {
	s := vin + "|" + id + "|" + key + "|" + secret

	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)

	controlSum := hex.EncodeToString(bs[0:5])

	return "/" + key + "/" + controlSum + "/" + id + "/" + vin + ".json?new"
}

type VincarioMarketValueResponse struct {
	Vin           string `json:"vin"`
	Price         int    `json:"price"`
	PriceCurrency string `json:"price_currency"`
	Balance       struct {
		APIDecode             int `json:"API Decode"`
		APIStolenCheck        int `json:"API Stolen Check"`
		APIVehicleMarketValue int `json:"API Vehicle Market Value"`
		APIOEMVINLookup       int `json:"API OEM VIN Lookup"`
	} `json:"balance"`
	Vehicle struct {
		VehicleId int    `json:"vehicle_id"`
		Make      string `json:"make"`
		MakeId    int    `json:"make_id"`
		Model     string `json:"model"`
		ModelId   int    `json:"model_id"`
		ModelYear int    `json:"model_year"`
	} `json:"vehicle"`
	Period struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"period"`
	MarketPrice struct {
		Europe struct {
			PriceCount    int    `json:"price_count"`
			PriceCurrency string `json:"price_currency"`
			PriceBelow    int    `json:"price_below"`
			PriceMedian   int    `json:"price_median"`
			PriceAvg      int    `json:"price_avg"`
			PriceAbove    int    `json:"price_above"`
			PriceStdev    int    `json:"price_stdev"`
		} `json:"europe"`
		NorthAmerica struct {
			PriceCount    int    `json:"price_count"`
			PriceCurrency string `json:"price_currency"`
			PriceBelow    int    `json:"price_below"`
			PriceMedian   int    `json:"price_median"`
			PriceAvg      int    `json:"price_avg"`
			PriceAbove    int    `json:"price_above"`
			PriceStdev    int    `json:"price_stdev"`
		} `json:"north_america"`
	} `json:"market_price"`
	MarketOdometer struct {
		Europe struct {
			OdometerCount  int    `json:"odometer_count"`
			OdometerUnit   string `json:"odometer_unit"`
			OdometerMedian int    `json:"odometer_median"`
			OdometerAvg    int    `json:"odometer_avg"`
			OdometerStdev  int    `json:"odometer_stdev"`
		} `json:"europe"`
		NorthAmerica struct {
			OdometerCount  int    `json:"odometer_count"`
			OdometerUnit   string `json:"odometer_unit"`
			OdometerMedian int    `json:"odometer_median"`
			OdometerAvg    int    `json:"odometer_avg"`
			OdometerStdev  int    `json:"odometer_stdev"`
		} `json:"north_america"`
	} `json:"market_odometer"`
	Records []struct {
		Market        string `json:"market"`
		Continent     string `json:"continent"`
		Price         int    `json:"price"`
		PriceCurrency string `json:"price_currency"`
		Odometer      int    `json:"odometer,omitempty"`
		OdometerUnit  string `json:"odometer_unit,omitempty"`
	} `json:"records"`
}
