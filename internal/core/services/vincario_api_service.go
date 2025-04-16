package services

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	core "github.com/DIMO-Network/valuations-api/internal/core/models"

	"github.com/DIMO-Network/valuations-api/internal/config"

	"github.com/DIMO-Network/shared/pkg/http"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source vincario_api_service.go -destination mocks/vincario_api_service_mock.go -package mock_services
type VincarioAPIService interface {
	GetMarketValuation(vin string) (*core.VincarioMarketValueResponse, error)
}

type vincarioAPIService struct {
	settings      *config.Settings
	httpClientVIN http.ClientWrapper
	log           *zerolog.Logger
}

func NewVincarioAPIService(settings *config.Settings, log *zerolog.Logger) VincarioAPIService {
	if settings.VincarioAPIURL == "" || settings.VincarioAPISecret == "" {
		panic("Vincario configuration not set")
	}
	hcwv, _ := http.NewClientWrapper(settings.VincarioAPIURL, "", 10*time.Second, nil, false)

	return &vincarioAPIService{
		settings:      settings,
		httpClientVIN: hcwv,
		log:           log,
	}
}

func (va *vincarioAPIService) GetMarketValuation(vin string) (*core.VincarioMarketValueResponse, error) {
	id := "vehicle-market-value"

	urlPath := vincarioPathBuilder(vin, id, va.settings.VincarioAPIKey, va.settings.VincarioAPISecret)
	// url with api access
	resp, err := va.httpClientVIN.ExecuteRequest(urlPath, "GET", nil)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// decode JSON from response body
	var data core.VincarioMarketValueResponse
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if data.MarketPrice.Europe.PriceAvg == 0 && data.MarketPrice.NorthAmerica.PriceAvg == 0 {
		return nil, fmt.Errorf("invalid valuation with 0 value returned - %s", string(bodyBytes))
	}

	return &data, nil
}

func vincarioPathBuilder(vin, id, key, secret string) string {
	s := vin + "|" + id + "|" + key + "|" + secret

	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)

	controlSum := hex.EncodeToString(bs[0:5])

	return "/" + key + "/" + controlSum + "/" + id + "/" + vin + ".json?new"
}
