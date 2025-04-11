package gateways

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrNotFound = errors.New("not found")
var ErrBadRequest = errors.New("bad request")

type identityAPIService struct {
	httpClient     shared.HTTPClientWrapper
	logger         zerolog.Logger
	identityAPIURL string
}

//go:generate mockgen -source identity_api.go -destination mocks/identity_api_mock.go -package mocks
type IdentityAPI interface {
	GetManufacturer(slug string) (*Manufacturer, error)
	GetDefinition(definitionID string) (*DeviceDefinition, error)
	GetVehicle(tokenID uint64) (*Vehicle, error)
}

// NewIdentityAPIService creates a new instance of IdentityAPI, initializing it with the provided logger, settings, and HTTP client.
// httpClient is used for testing really
func NewIdentityAPIService(logger *zerolog.Logger, settings *config.Settings, httpClient shared.HTTPClientWrapper) IdentityAPI {
	if httpClient == nil {
		h := map[string]string{}
		h["Content-Type"] = "application/json"
		httpClient, _ = shared.NewHTTPClientWrapper("", "", 10*time.Second, h, false) // ok to ignore err since only used for tor check
	}

	return &identityAPIService{
		httpClient:     httpClient,
		logger:         *logger,
		identityAPIURL: settings.IdentityAPIURL.String(),
	}
}

func (i *identityAPIService) GetVehicle(tokenID uint64) (*Vehicle, error) {
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
			Vehicle Vehicle `json:"vehicle"`
		} `json:"data"`
	}
	err := i.fetchWithQuery(query, &wrapper)
	if err != nil {
		return nil, err
	}
	if wrapper.Data.Vehicle.Id == "" {
		return nil, errors.Wrapf(ErrNotFound, "identity-api did not find vehicle with tokenId: %d", tokenID)
	}
	return &wrapper.Data.Vehicle, nil
}

func (i *identityAPIService) GetDefinition(definitionID string) (*DeviceDefinition, error) {
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
			DeviceDefinition DeviceDefinition `json:"deviceDefinition"`
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
func (i *identityAPIService) GetManufacturer(name string) (*Manufacturer, error) {
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

func (i *identityAPIService) fetchWithQuery(query string, result interface{}) error {
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

type Manufacturer struct {
	TokenID int    `json:"tokenId"`
	Name    string `json:"name"`
	TableID int    `json:"tableId"`
	Owner   string `json:"owner"`
}

type GraphQLRequest struct {
	Query string `json:"query"`
}

type DeviceDefinition struct {
	Model        string       `json:"model"`
	Year         int          `json:"year"`
	Manufacturer Manufacturer `json:"manufacturer"`
	ImageURI     string       `json:"imageURI"`
	Attributes   []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"attributes"`
}

type Vehicle struct {
	Id         string `json:"id"`
	Definition struct {
		Id    string `json:"id"`
		Make  string `json:"make"`
		Model string `json:"model"`
		Year  int    `json:"year"`
	} `json:"definition"`
	Owner string `json:"owner"`
}
