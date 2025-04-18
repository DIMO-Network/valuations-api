package gateways

import (
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
	httpClient http.ClientWrapper
	logger     zerolog.Logger
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
	httpClient, _ := http.NewClientWrapper(settings.IdentityAPIURL.String(), "", 10*time.Second, nil, true) // ok to ignore err since only used for tor check

	return &identityAPIService{
		httpClient: httpClient,
		logger:     *logger,
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
	err := i.httpClient.GraphQLQuery("", query, &wrapper)
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
	err := i.httpClient.GraphQLQuery("", query, &wrapper)
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
	err := i.httpClient.GraphQLQuery("", query, &wrapper)
	if err != nil {
		return nil, err
	}
	if wrapper.Data.Manufacturer.Name == "" {
		return nil, errors.Wrapf(ErrNotFound, "identity-api did not find manufacturer with name: %s", name)
	}
	return &wrapper.Data.Manufacturer, nil
}
