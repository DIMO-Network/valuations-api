package services

import (
	"encoding/json"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

const NorthAmercanCountries = "USA,CAN,MEX,PRI"

type VehicleMintValuationIngest interface {
	ProcessVehicleMintMsg(ctx goka.Context, msg any)
}

type vehicleMintValuationIngest struct {
	DBS                      func() *db.ReaderWriter
	logger                   zerolog.Logger
	userDeviceService        UserDeviceAPIService
	vincarioValuationService VincarioValuationService
	drivlyValuationService   DrivlyValuationService
}

func NewVehicleMintValuationIngest(dbs func() *db.ReaderWriter, logger zerolog.Logger, settings *config.Settings,
	userDeviceService UserDeviceAPIService,
	ddSvc DeviceDefinitionsAPIService,
	uddSvc UserDeviceDataAPIService,
) VehicleMintValuationIngest {
	return &vehicleMintValuationIngest{
		DBS:                      dbs,
		logger:                   logger,
		userDeviceService:        userDeviceService,
		vincarioValuationService: NewVincarioValuationService(dbs, &logger, settings, userDeviceService),
		drivlyValuationService:   NewDrivlyValuationService(dbs, &logger, settings, ddSvc, uddSvc, userDeviceService),
	}
}

// ProcessVehicleMintMsg gets mint event types and requests a valuation and offer for the VIN in the message
func (i *vehicleMintValuationIngest) ProcessVehicleMintMsg(ctx goka.Context, msg any) {
	// if have issues with context etc use context.Background() instead of the goka one
	event := msg.(*VehicleMintEvent)
	// event.ID is the userDeviceId, set from devices-api
	if event.Type != "com.dimo.zone.device.mint" {
		i.logger.Info().Msgf("not processing event since of type: %s", event.Type) // change this to debug level after testing
	}
	localLog := i.logger.With().Str("userDeviceId", event.ID).Logger()
	// change below to debug once validate
	localLog.Info().Str("payload", string(event.Data)).Msg("processing vehicle mint event for valuation/offer trigger")

	jsonBytes, err := event.Data.MarshalJSON()
	if err != nil {
		localLog.Err(err).Msg("failed to marshal event data")
		return
	}
	vin := gjson.GetBytes(jsonBytes, "device.vin").String()
	tokenId := gjson.GetBytes(jsonBytes, "nft.tokenId").Uint()
	localLog = localLog.With().Str("vin", vin).Uint64("tokenId", tokenId).Logger()

	if len(vin) == 0 {
		localLog.Error().Msg("vin was empty")
		return
	}
	userDevice, err := i.userDeviceService.GetUserDevice(ctx.Context(), event.ID)
	if err != nil {
		localLog.Error().Msg("unable to find user device")
		return
	}
	localLog = localLog.With().Str("country", userDevice.CountryCode).Str("deviceDefinitionId", userDevice.DeviceDefinitionId).Logger()
	// todo: pass the tokenId so it can be set in the db
	// todo: move tests over
	// we currently have two vendors for valuations
	if strings.Contains(NorthAmercanCountries, userDevice.CountryCode) {
		status, err := i.drivlyValuationService.PullValuation(ctx.Context(), userDevice.Id, userDevice.DeviceDefinitionId, vin)
		if err != nil {
			localLog.Err(err).Msg("valuation request - error pulling drivly data")
		} else {
			localLog.Info().Msgf("valuation request from Drivly completed OK with status %s", status)
		}
		// in NA, we can also pull the offer
		status, err = i.drivlyValuationService.PullOffer(ctx.Context(), userDevice.Id)
		if err != nil && status != core.SkippedDataPullStatus {
			localLog.Err(err).Msg("failed to process offer request due to internal error")
		} else {
			localLog.Info().Msgf("valuation request from Drivly completed OK with status %s", status)
		}
	} else {
		status, err := i.vincarioValuationService.PullValuation(ctx.Context(), userDevice.Id, userDevice.DeviceDefinitionId, vin)
		if err != nil {
			localLog.Err(err).Msg("valuation request - error pulling vincario data")
		} else {
			localLog.Info().Msgf("valuation request from Vincario completed OK with status %s", status)
		}
	}
	return
	// todo metrics
}

// VehicleMintEvent is emitted by devices-api registry/storage.go
type VehicleMintEvent struct {
	ID          string          `json:"id"`
	Source      string          `json:"source"`
	Specversion string          `json:"specversion"`
	Subject     string          `json:"subject"`
	Time        time.Time       `json:"time"`
	Type        string          `json:"type"`
	Data        json.RawMessage `json:"data"`
}
