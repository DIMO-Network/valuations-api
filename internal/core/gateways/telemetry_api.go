package gateways

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/DIMO-Network/valuations-api/internal/config"
	coremodels "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type telemetryAPIService struct {
	logger          zerolog.Logger
	telemetryAPIURL string
}

//go:generate mockgen -source telemetry_api.go -destination mocks/telemetry_api_mock.go -package mock_gateways
type TelemetryAPI interface {
	GetLatestSignals(ctx context.Context, tokenID uint64, authHeader string) (*coremodels.SignalsLatest, error)
	GetVinVC(ctx context.Context, tokenID uint64, authHeader string) (*coremodels.VinVCLatest, error)
}

func NewTelemetryAPI(logger *zerolog.Logger, settings *config.Settings) TelemetryAPI {
	return &telemetryAPIService{
		logger:          *logger,
		telemetryAPIURL: settings.TelemetryAPIURL.String(),
	}
}

// GetVinVC gets the VIN. authHeader must be full string with Bearer xxx
func (i *telemetryAPIService) GetVinVC(ctx context.Context, tokenID uint64, authHeader string) (*coremodels.VinVCLatest, error) {
	query := `{
vinVCLatest(tokenId:` + strconv.Itoa(int(tokenID)) + `) {
    vin
    recordedBy
    recordedAt
    countryCode
    validFrom
    validTo
  }
}`
	req, err := http.NewRequest("POST", i.telemetryAPIURL, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get vinVC, status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var wrapper struct {
		Data struct {
			VinVCLatest coremodels.VinVCLatest `json:"vinVCLatest"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, errors.Wrapf(err, "failed to decode response")
	}

	if wrapper.Data.VinVCLatest.Vin == "" {
		return nil, errors.Wrapf(ErrNotFound, "no vinVCLatest for tokenId: %d", tokenID)
	}
	return &wrapper.Data.VinVCLatest, nil
}

// GetLatestSignals odometer and location. authHeader must be full string with Bearer xxx
func (i *telemetryAPIService) GetLatestSignals(ctx context.Context, tokenID uint64, authHeader string) (*coremodels.SignalsLatest, error) {
	query := `{
signalsLatest(tokenId:` + strconv.Itoa(int(tokenID)) + `) {
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
	req, err := http.NewRequest("POST", i.telemetryAPIURL, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get vinVC, status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var wrapper struct {
		Data struct {
			SignalsLatest coremodels.SignalsLatest `json:"signalsLatest"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, errors.Wrapf(err, "failed to decode response")
	}

	return &wrapper.Data.SignalsLatest, nil
}
