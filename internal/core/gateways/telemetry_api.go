package gateways

import (
	"encoding/json"
	"io"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type telemetryAPIService struct {
	httpClient     shared.HTTPClientWrapper
	logger         zerolog.Logger
	identityAPIURL string
}

//go:generate mockgen -source telemetry_api.go -destination mocks/telemetry_api_mock.go -package mocks
type TelemetryAPI interface {
	GetLatestSignals(name string) (*Manufacturer, error)
}

func NewTelemetryAPI(logger *zerolog.Logger, settings *config.Settings, httpClient shared.HTTPClientWrapper) TelemetryAPI {
	if httpClient == nil {
		h := map[string]string{}
		h["Content-Type"] = "application/json"
		httpClient, _ = shared.NewHTTPClientWrapper("", "", 10*time.Second, h, false) // ok to ignore err since only used for tor check
	}

	return &telemetryAPIService{
		httpClient:     httpClient,
		logger:         *logger,
		identityAPIURL: settings.IdentityAPIURL.String(),
	}
}

func (i *telemetryAPIService) GetLatestSignals(name string) (*Manufacturer, error) {
	query := `{
  manufacturer(by: {name: "` + name + `"}) {
    	tokenId
    	name
    	tableId
    	owner
  	  }
	}`
	var wrapper struct {
		Data struct {
			Manufacturer Manufacturer `json:"manufacturer"`
		} `json:"data"`
	}
	err := i.fetchWithQuery(query, &wrapper)
	if err != nil {
		return nil, err
	}
	if wrapper.Data.Manufacturer.Name == "" {
		return nil, errors.Wrapf(ErrNotFound, "identity-api did not find manufacturer with name: %s", name)
	}
	return &wrapper.Data.Manufacturer, nil
}

func (i *telemetryAPIService) fetchWithQuery(query string, result interface{}) error {
	// GraphQL request
	requestPayload := GraphQLRequest{Query: query}
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return err
	}

	// POST request
	res, err := i.httpClient.ExecuteRequest(i.identityAPIURL, "POST", payloadBytes)
	if err != nil {
		i.logger.Err(err).Str("func", "fetchWithQuery").Msgf("request payload: %s", string(payloadBytes))
		if _, ok := err.(shared.HTTPResponseError); !ok {
			return errors.Wrapf(err, "error calling identity api from url %s", i.identityAPIURL)
		}
	}
	defer res.Body.Close() // nolint

	if res.StatusCode == 404 {
		return ErrNotFound
	}
	if res.StatusCode == 400 {
		return ErrBadRequest
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Wrapf(err, "error reading response body from url %s", i.identityAPIURL)
	}

	if err := json.Unmarshal(bodyBytes, result); err != nil {
		return err
	}

	return nil
}
