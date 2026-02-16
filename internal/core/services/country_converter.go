package services

const NorthAmercanCountries = "USA,CAN,MEX,PRI"

// ConvertSupportedCountry converts two letter country code to three letter country code, but only for the countries we support currently (USA)
func ConvertSupportedCountry(countryTwoLetter string) string {
	switch countryTwoLetter {
	case "US":
		return "USA"
	}
	return ""
}
