package models

type DrivlyVINSummary struct {
	Pricing   map[string]interface{}
	Offers    map[string]interface{}
	AutoCheck map[string]interface{}
	Build     map[string]interface{}
	Cargurus  map[string]interface{}
	Carvana   map[string]interface{}
	Carmax    map[string]interface{}
	Carstory  map[string]interface{}
	Edmunds   map[string]interface{}
	TMV       map[string]interface{}
	KBB       map[string]interface{}
	VRoom     map[string]interface{}
}

type ValuationRequestData struct {
	Mileage *float64 `json:"mileage,omitempty"`
	ZipCode *string  `json:"zipCode,omitempty"`
}
