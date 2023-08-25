package models

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
