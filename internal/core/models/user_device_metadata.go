package models

import (
	"encoding/json"
	"fmt"
)

type PowertrainType string

const (
	ICE  PowertrainType = "ICE"
	HEV  PowertrainType = "HEV"
	PHEV PowertrainType = "PHEV"
	BEV  PowertrainType = "BEV"
	FCEV PowertrainType = "FCEV"
)

func (p PowertrainType) String() string {
	return string(p)
}

func (p *PowertrainType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// Potentially an invalid value.
	switch bv := PowertrainType(s); bv {
	case ICE, HEV, PHEV, BEV, FCEV:
		*p = bv
		return nil
	default:
		return fmt.Errorf("unrecognized value: %s", s)
	}
}

type UserDeviceMetadata struct {
	PowertrainType          *PowertrainType `json:"powertrainType,omitempty"`
	ElasticDefinitionSynced bool            `json:"elasticDefinitionSynced,omitempty"`
	ElasticRegionSynced     bool            `json:"elasticRegionSynced,omitempty"`
	PostalCode              *string         `json:"postal_code"`
	GeoDecodedCountry       *string         `json:"geoDecodedCountry"`
	GeoDecodedStateProv     *string         `json:"geoDecodedStateProv"`
	// CANProtocol is the protocol that was detected by edge-network from the autopi.
	CANProtocol *string `json:"canProtocol,omitempty"`
}
