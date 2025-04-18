package gateways

import (
	"strconv"
	"time"

	"github.com/DIMO-Network/shared/pkg/http"

	"github.com/DIMO-Network/valuations-api/internal/config"
	coremodels "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type telemetryAPIService struct {
	logger     zerolog.Logger
	httpClient http.ClientWrapper
}

//go:generate mockgen -source telemetry_api.go -destination mocks/telemetry_api_mock.go -package mock_gateways
type TelemetryAPI interface {
	GetLatestSignals(tokenID uint64, authHeader string) (*coremodels.SignalsLatest, error)
	GetVinVC(tokenID uint64, authHeader string) (*coremodels.VinVCLatest, error)
}

func NewTelemetryAPI(logger *zerolog.Logger, settings *config.Settings) TelemetryAPI {
	h := map[string]string{}
	httpClient, _ := http.NewClientWrapper(settings.TelemetryAPIURL.String(), "", 10*time.Second, h, true) // ok to ignore err since only used for tor check

	return &telemetryAPIService{
		logger:     *logger,
		httpClient: httpClient,
	}
}

// GetVinVC gets the VIN. authHeader must be full string with Bearer xxx
func (i *telemetryAPIService) GetVinVC(tokenID uint64, authHeader string) (*coremodels.VinVCLatest, error) {
	query := `{
vinVCLatest(tokenId:` + strconv.FormatUint(tokenID, 10) + `) {
    vin
    recordedBy
    recordedAt
    countryCode
    validFrom
    validTo
  }
}`

	var wrapper struct {
		Data struct {
			VinVCLatest coremodels.VinVCLatest `json:"vinVCLatest"`
		} `json:"data"`
	}
	err := i.httpClient.GraphQLQuery(authHeader, query, &wrapper)
	if err != nil {
		return nil, err
	}

	if wrapper.Data.VinVCLatest.Vin == "" {
		return nil, errors.Wrapf(ErrNotFound, "no vinVCLatest for tokenId: %d", tokenID)
	}
	return &wrapper.Data.VinVCLatest, nil
}

// GetLatestSignals odometer and location. authHeader must be full string with Bearer xxx
func (i *telemetryAPIService) GetLatestSignals(tokenID uint64, authHeader string) (*coremodels.SignalsLatest, error) {
	query := `{
signalsLatest(tokenId:` + strconv.FormatUint(tokenID, 10) + `) {
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
	}
}`

	var wrapper struct {
		Data struct {
			SignalsLatest coremodels.SignalsLatest `json:"signalsLatest"`
		} `json:"data"`
	}
	err := i.httpClient.GraphQLQuery(authHeader, query, &wrapper)
	if err != nil {
		return nil, err
	}

	return &wrapper.Data.SignalsLatest, nil
}
