// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v2/vehicles/{tokenId}/instant-offer": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "makes a request for an instant offer for a particular user device. Simply returns success if able to create job.\nYou will need to query the offers endpoint to see if a new offer showed up. Job can take about a minute to complete.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "offers"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "tokenId for vehicle to get offers",
                        "name": "tokenId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/v2/vehicles/{tokenId}/offers": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "gets any existing offers for a particular user device. You must call instant-offer endpoint first to pull newer. Returns list.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "offers"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "tokenId for vehicle to get offers",
                        "name": "tokenId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_DIMO-Network_valuations-api_internal_core_models.DeviceOffer"
                        }
                    }
                }
            }
        },
        "/v2/vehicles/{tokenId}/valuations": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "gets valuations for a particular user device. Includes only price valuations, not offers. gets list of most recent",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "valuations"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "tokenId for vehicle to get offers",
                        "name": "tokenId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_DIMO-Network_valuations-api_internal_core_models.DeviceValuation"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "github_com_DIMO-Network_valuations-api_internal_core_models.DeviceOffer": {
            "type": "object",
            "properties": {
                "offerSets": {
                    "description": "Contains a list of offer sets, one for each source",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_DIMO-Network_valuations-api_internal_core_models.OfferSet"
                    }
                }
            }
        },
        "github_com_DIMO-Network_valuations-api_internal_core_models.DeviceValuation": {
            "type": "object",
            "properties": {
                "valuationSets": {
                    "description": "Contains a list of valuation sets, one for each vendor",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_DIMO-Network_valuations-api_internal_core_models.ValuationSet"
                    }
                }
            }
        },
        "github_com_DIMO-Network_valuations-api_internal_core_models.OdometerMeasurementEnum": {
            "type": "string",
            "enum": [
                "Real",
                "Estimated",
                "Market"
            ],
            "x-enum-varnames": [
                "Real",
                "Estimated",
                "Market"
            ]
        },
        "github_com_DIMO-Network_valuations-api_internal_core_models.Offer": {
            "type": "object",
            "properties": {
                "declineReason": {
                    "description": "The reason the offer was declined from the vendor",
                    "type": "string"
                },
                "error": {
                    "description": "An error from the vendor (eg. when the VIN is invalid)",
                    "type": "string"
                },
                "grade": {
                    "description": "The grade of the offer from the vendor (eg. \"RETAIL\")",
                    "type": "string"
                },
                "price": {
                    "description": "The offer price from the vendor",
                    "type": "integer"
                },
                "url": {
                    "description": "The offer URL from the vendor",
                    "type": "string"
                },
                "vendor": {
                    "description": "The vendor of the offer (eg. \"carmax\", \"carvana\", etc.)",
                    "type": "string"
                }
            }
        },
        "github_com_DIMO-Network_valuations-api_internal_core_models.OfferSet": {
            "type": "object",
            "properties": {
                "mileage": {
                    "description": "The mileage used for the offers",
                    "type": "integer"
                },
                "odometerMeasurementType": {
                    "description": "Whether estimated, real, or from the market (eg. vincario). Market, Estimated, Real",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_DIMO-Network_valuations-api_internal_core_models.OdometerMeasurementEnum"
                        }
                    ]
                },
                "offers": {
                    "description": "Contains a list of offers from the source",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_DIMO-Network_valuations-api_internal_core_models.Offer"
                    }
                },
                "source": {
                    "description": "The source of the offers (eg. \"drivly\")",
                    "type": "string"
                },
                "updated": {
                    "description": "The time the offers were pulled",
                    "type": "string"
                },
                "zipCode": {
                    "description": "This will be the zip code used (if any) for the offers request regardless if the source uses it",
                    "type": "string"
                }
            }
        },
        "github_com_DIMO-Network_valuations-api_internal_core_models.ValuationSet": {
            "type": "object",
            "properties": {
                "currency": {
                    "description": "eg. USD or EUR",
                    "type": "string"
                },
                "mileage": {
                    "description": "The mileage used for the valuation",
                    "type": "integer"
                },
                "odometer": {
                    "description": "Odometer used for calculating values",
                    "type": "integer"
                },
                "odometerMeasurementType": {
                    "description": "whether estimated, real, or from the market (eg. vincario). Market, Estimated, Real",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_DIMO-Network_valuations-api_internal_core_models.OdometerMeasurementEnum"
                        }
                    ]
                },
                "odometerUnit": {
                    "type": "string"
                },
                "retail": {
                    "description": "retail is equal to retailAverage when available",
                    "type": "integer"
                },
                "retailAverage": {
                    "type": "integer"
                },
                "retailClean": {
                    "description": "retailClean, retailAverage, and retailRough my not always be available",
                    "type": "integer"
                },
                "retailRough": {
                    "type": "integer"
                },
                "retailSource": {
                    "description": "Useful when Drivly returns multiple vendors and we've selected one (eg. \"drivly:blackbook\")",
                    "type": "string"
                },
                "tradeIn": {
                    "description": "tradeIn is equal to tradeInAverage when available",
                    "type": "integer"
                },
                "tradeInAverage": {
                    "type": "integer"
                },
                "tradeInClean": {
                    "description": "tradeInClean, tradeInAverage, and tradeInRough my not always be available",
                    "type": "integer"
                },
                "tradeInRough": {
                    "type": "integer"
                },
                "tradeInSource": {
                    "description": "Useful when Drivly returns multiple vendors and we've selected one (eg. \"drivly:blackbook\")",
                    "type": "string"
                },
                "updated": {
                    "description": "The time the valuation was pulled or in the case of blackbook, this may be the event time of the device odometer which was used for the valuation",
                    "type": "string"
                },
                "userDisplayPrice": {
                    "description": "UserDisplayPrice the top level value to show to users in mobile app",
                    "type": "integer"
                },
                "vendor": {
                    "description": "The source of the valuation (eg. \"drivly\" or \"blackbook\")",
                    "type": "string"
                },
                "zipCode": {
                    "description": "This will be the zip code used (if any) for the valuation request regardless if the vendor uses it",
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "DIMO Vehicle Valuations API",
	Description:      "API to get latest valuation for a given connected vehicle belonging to user. Tokens must be privilege tokens.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
