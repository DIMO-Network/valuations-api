package models

type DataPullStatusEnum string

const (
	// PulledInfoAndValuationStatus means we pulled vin, edmunds, build and valuations
	PulledInfoAndValuationStatus DataPullStatusEnum = "PulledAll"
	// PulledValuationDrivlyStatus means we only pulled offers and pricing
	PulledValuationDrivlyStatus   DataPullStatusEnum = "PulledValuations"
	PulledValuationVincarioStatus DataPullStatusEnum = "PulledValuationVincario"
	SkippedDataPullStatus         DataPullStatusEnum = "Skipped"
	ErrorDataPullStatus           DataPullStatusEnum = "Error"
)
