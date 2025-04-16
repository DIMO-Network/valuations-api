package gateways

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	coremodels "github.com/DIMO-Network/valuations-api/internal/core/models"

	"github.com/DIMO-Network/shared/pkg/http"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrNotFound = errors.New("not found")
var ErrBadRequest = errors.New("bad request")

type identityAPIService struct {
	httpClient     http.ClientWrapper
	logger         zerolog.Logger
	identityAPIURL string
}

//go:generate mockgen -source identity_api.go -destination mocks/identity_api_mock.go -package mock_gateways
type IdentityAPI interface {
	GetManufacturer(slug string) (*coremodels.Manufacturer, error)
	GetDefinition(definitionID string) (*coremodels.DeviceDefinition, error)
	GetVehicle(tokenID uint64) (*coremodels.Vehicle, error)
}

// NewIdentityAPIService creates a new instance of IdentityAPI, initializing it with the provided logger, settings, and HTTP client.
// httpClient is used for testing really
func NewIdentityAPIService(logger *zerolog.Logger, settings *config.Settings) IdentityAPI {
	h := map[string]string{}
	h["Content-Type"] = "application/json"
	httpClient, _ := http.NewClientWrapper("", "", 10*time.Second, h, false) // ok to ignore err since only used for tor check

	return &identityAPIService{
		httpClient:     httpClient,
		logger:         *logger,
		identityAPIURL: settings.IdentityAPIURL.String(),
	}
}

func (i *identityAPIService) GetVehicle(tokenID uint64) (*coremodels.Vehicle, error) {
	query := `{
  vehicle(tokenId: ` + strconv.FormatUint(tokenID, 10) + `) {
    id
    definition{
      id
      make
      model
      year
    }
    owner 
	}
  }`
	var wrapper struct {
		Data struct {
			Vehicle coremodels.Vehicle `json:"vehicle"`
		} `json:"data"`
	}
	err := i.fetchWithQuery(query, &wrapper)
	if err != nil {
		return nil, err
	}
	if wrapper.Data.Vehicle.ID == "" {
		return nil, errors.Wrapf(ErrNotFound, "identity-api did not find vehicle with tokenId: %d", tokenID)
	}
	return &wrapper.Data.Vehicle, nil
}

func (i *identityAPIService) GetDefinition(definitionID string) (*coremodels.DeviceDefinition, error) {
	query := `{
  deviceDefinition(by: {id: "` + definitionID + `"}) {
    model,
    year,
    manufacturer {
      tokenId
      name
    },
    imageURI,
    
    attributes {
      name,
      value
    }
  }
	}`
	var wrapper struct {
		Data struct {
			DeviceDefinition coremodels.DeviceDefinition `json:"deviceDefinition"`
		} `json:"data"`
	}
	err := i.fetchWithQuery(query, &wrapper)
	if err != nil {
		return nil, err
	}
	if wrapper.Data.DeviceDefinition.Model == "" {
		return nil, errors.Wrapf(ErrNotFound, "identity-api did not find device definition with id: %s", definitionID)
	}
	return &wrapper.Data.DeviceDefinition, nil
}

// GetManufacturer from identity-api by the name - must match exactly. Returns the token id and other on chain info
func (i *identityAPIService) GetManufacturer(name string) (*coremodels.Manufacturer, error) {
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
			Manufacturer coremodels.Manufacturer `json:"manufacturer"`
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

func (i *identityAPIService) fetchWithQuery(query string, result interface{}) error {
	// GraphQL request
	requestPayload := coremodels.GraphQLRequest{Query: query}
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return err
	}

	// POST request
	res, err := i.httpClient.ExecuteRequest(i.identityAPIURL, "POST", payloadBytes)
	if err != nil {
		i.logger.Err(err).Str("func", "fetchWithQuery").Msgf("request payload: %s", string(payloadBytes))
		if _, ok := err.(http.ResponseError); !ok {
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
