package services

import (
	"encoding/json"
	"github.com/DIMO-Network/valuations-api/internal/core/gateways"
	"strings"
	"time"

	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/shared/pkg/payloads"
	"github.com/DIMO-Network/valuations-api/internal/config"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

const NorthAmercanCountries = "USA,CAN,MEX,PRI"

type VehicleMintValuationIngest interface {
	ProcessVehicleMintMsg(ctx goka.Context, msg any)
}

type vehicleMintValuationIngest struct {
	DBS                      func() *db.ReaderWriter
	logger                   zerolog.Logger
	vincarioValuationService VincarioValuationService
	drivlyValuationService   DrivlyValuationService
	telemetryAPI             gateways.TelemetryAPI
	identityAPI              gateways.IdentityAPI
	locationSvc              LocationService
}

func NewVehicleMintValuationIngest(dbs func() *db.ReaderWriter, logger zerolog.Logger, settings *config.Settings,
	telemetryAPI gateways.TelemetryAPI,
	identityAPI gateways.IdentityAPI,
) VehicleMintValuationIngest {
	return &vehicleMintValuationIngest{
		DBS:                      dbs,
		logger:                   logger,
		identityAPI:              identityAPI,
		telemetryAPI:             telemetryAPI,
		vincarioValuationService: NewVincarioValuationService(dbs, &logger, settings, identityAPI),
		drivlyValuationService:   NewDrivlyValuationService(dbs, &logger, settings),
		locationSvc:              NewLocationService(dbs, settings, &logger),
	}
}

// ProcessVehicleMintMsg gets mint event types and requests a valuation and offer for the VIN in the message
func (i *vehicleMintValuationIngest) ProcessVehicleMintMsg(ctx goka.Context, msg any) {
	// if have issues with context etc use context.Background() instead of the goka one
	event := msg.(*payloads.CloudEvent[json.RawMessage])
	// event.ID is the userDeviceId, set from devices-api
	if event.Type != "com.dimo.zone.device.mint" && event.Source != "devices-api" {
		i.logger.Debug().Msgf("not processing event since of type: %s", event.Type) // change this to debug level after testing
		return
	}
	localLog := i.logger.With().Str("func", "ProcessVehicleMintMsg").Logger()

	jsonBytes, err := event.Data.MarshalJSON()
	if err != nil {
		i.logger.Err(err).Msg("failed to marshal event data")
		return
	}
	// we can access the data based on devices-api services.UserDeviceMintEvent
	vin := strings.TrimSpace(gjson.GetBytes(jsonBytes, "device.vin").String())
	tokenID := gjson.GetBytes(jsonBytes, "nft.tokenId").Uint()
	localLog = localLog.With().Str("vin", vin).Uint64("token_id", tokenID).Logger()
	if len(vin) != 17 && tokenID == 0 {
		localLog.Warn().Str("payload", string(event.Data)).Msg("invalid vin or tokenId")
		return
	}

	vehicle, err := i.identityAPI.GetVehicle(tokenID)
	if err != nil {
		localLog.Error().Msg("unable to find vehicle")
		return
	}

	localLog = localLog.With().Str("definition_id", vehicle.Definition.Id).Logger()

	// problem here is most likely we won't have any telemetry data yet since vehicle was just minted
	// ideally this would go into a delayed queue, or be triggered from events when we get telemetry
	// and then send a notification to user later on when we get their valuation
	signals, err := i.telemetryAPI.GetLatestSignals(tokenID)
	if err != nil {
		localLog.Error().Uint64("token_id", tokenID).Msg("unable to find telemetry signals, skipping")
		return
	}
	location, err := i.locationSvc.GetGeoDecodedLocation(ctx.Context(), signals, tokenID)
	if err != nil {
		localLog.Error().Uint64("token_id", tokenID).Msg("unable to find location, skipping")
		return
	}
	// we currently have two vendors for valuations
	if strings.Contains(NorthAmercanCountries, location.CountryCode) {
		status, err := i.drivlyValuationService.PullValuation(ctx.Context(), tokenID, vehicle.Definition.Id, vin)
		if err != nil {
			localLog.Err(err).Msg("valuation request - error pulling drivly data")
		} else {
			localLog.Info().Msgf("valuation request from Drivly completed OK with status %s", status)
		}
		// in NA, we can also pull the offer
		status, err = i.drivlyValuationService.PullOffer(ctx.Context(), tokenID, vin)
		if err != nil && status != core.SkippedDataPullStatus {
			localLog.Err(err).Msg("failed to process offer request due to internal error")
		} else {
			localLog.Info().Msgf("instant offer from Drivly completed OK with status %s", status)
		}
	} else {
		status, err := i.vincarioValuationService.PullValuation(ctx.Context(), tokenID, vehicle.Definition.Id, vin)
		if err != nil {
			localLog.Err(err).Msg("valuation request - error pulling vincario data")
		} else {
			localLog.Info().Msgf("valuation request from Vincario completed OK with status %s", status)
		}
	}
	// todo metrics
}

// VehicleMintEvent is emitted by devices-api registry/storage.go Handle(...)
type VehicleMintEvent struct {
	ID          string          `json:"id"`
	Source      string          `json:"source"`
	Specversion string          `json:"specversion"`
	Subject     string          `json:"subject"`
	Time        time.Time       `json:"time"`
	Type        string          `json:"type"`
	Data        json.RawMessage `json:"data"`
}
