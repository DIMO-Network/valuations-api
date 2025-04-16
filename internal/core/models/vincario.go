package models

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
	// nolint
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
