package gateways

import (
	"context"
	"github.com/setnicka/graphql"
	"strconv"
	"time"

	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type telemetryAPIService struct {
	logger  zerolog.Logger
	gclient *graphql.Client
}

//go:generate mockgen -source telemetry_api.go -destination mocks/telemetry_api_mock.go -package mock_gateways
type TelemetryAPI interface {
	GetLatestSignals(ctx context.Context, tokenID uint64, authHeader string) (*SignalsLatest, error)
	GetVinVC(ctx context.Context, tokenID uint64, authHeader string) (*VinVCLatest, error)
}

func NewTelemetryAPI(logger *zerolog.Logger, settings *config.Settings) TelemetryAPI {

	return &telemetryAPIService{
		logger:  *logger,
		gclient: graphql.NewClient(settings.TelemetryAPIURL.String()),
	}
}

// GetVinVC gets the VIN. authHeader must be full string with Bearer xxx
func (i *telemetryAPIService) GetVinVC(ctx context.Context, tokenID uint64, authHeader string) (*VinVCLatest, error) {
	tIDStr := strconv.FormatUint(tokenID, 10)
	req := graphql.NewRequest(`vinVCLatest(tokenId:$tokenId) {
    vin
    recordedBy
    recordedAt
    countryCode
    validFrom
    validTo
  }`)
	req.Var("tokenId", tIDStr)
	req.Header.Set("Authorization", authHeader)

	var wrapper struct {
		Data struct {
			VinVCLatest VinVCLatest `json:"vinVCLatest"`
		} `json:"data"`
	}

	if err := i.gclient.Run(ctx, req, &wrapper); err != nil {
		return nil, err
	}

	if wrapper.Data.VinVCLatest.Vin == "" {
		return nil, errors.Wrapf(ErrNotFound, "no vinVCLatest for tokenId: %d", tokenID)
	}
	return &wrapper.Data.VinVCLatest, nil
}

// GetLatestSignals odometer and location. authHeader must be full string with Bearer xxx
func (i *telemetryAPIService) GetLatestSignals(ctx context.Context, tokenID uint64, authHeader string) (*SignalsLatest, error) {
	tIDStr := strconv.FormatUint(tokenID, 10)
	req := graphql.NewRequest(`signalsLatest(tokenId:$tokenId) {
		powertrainTransmissionTravelledDistance {
			timestamp
			value
		}
		currentLocationLatitude {
			timestamp
			value
		}
		currentLocationLongitude {
			timestamp
			value
		}
}`)
	req.Var("tokenId", tIDStr)
	req.Header.Set("Authorization", authHeader)

	var wrapper struct {
		Data struct {
			SignalsLatest SignalsLatest `json:"signalsLatest"`
		} `json:"data"`
	}
	if err := i.gclient.Run(ctx, req, &wrapper); err != nil {
		return nil, err
	}
	if wrapper.Data.SignalsLatest.PowertrainTransmissionTravelledDistance.Value == 0 {
		return nil, errors.Wrapf(ErrNotFound, "no odometer for tokenId: %s", tokenID)
	}
	return &wrapper.Data.SignalsLatest, nil
}

type SignalsLatest struct {
	PowertrainTransmissionTravelledDistance struct {
		Timestamp time.Time `json:"timestamp"`
		Value     float64   `json:"value"`
	} `json:"powertrainTransmissionTravelledDistance"`
	CurrentLocationLatitude struct {
		Timestamp time.Time `json:"timestamp"`
		Value     float64   `json:"value"`
	} `json:"currentLocationLatitude"`
	CurrentLocationLongitude struct {
		Timestamp time.Time `json:"timestamp"`
		Value     float64   `json:"value"`
	} `json:"currentLocationLongitude"`
}

type VinVCLatest struct {
	Vin         string    `json:"vin"`
	RecordedBy  string    `json:"recordedBy"`
	RecordedAt  time.Time `json:"recordedAt"`
	CountryCode string    `json:"countryCode"`
	ValidFrom   time.Time `json:"validFrom"`
	ValidTo     time.Time `json:"validTo"`
}
