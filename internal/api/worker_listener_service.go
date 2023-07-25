package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/valuations-api/internal/core/services"

	"github.com/DIMO-Network/shared"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	vehicleValuationRequestEventType = "zone.dimo.vehicle.valuation.request"
)

type WorkerListenerService struct {
	logger                   zerolog.Logger
	userDeviceService        services.UserDeviceAPIService
	vincarioValuationService services.VincarioValuationService
	drivlyValuationService   services.DrivlyValuationService
}

type RunValuationRequest struct {
	VIN          string `json:"vin"`
	UserDeviceID string `json:"userDeviceId"`
}

func NewWorkerListenerService(logger zerolog.Logger, userDeviceService services.UserDeviceAPIService,
	vincarioValuationService services.VincarioValuationService,
	drivlyValuationService services.DrivlyValuationService) *WorkerListenerService {
	return &WorkerListenerService{
		logger:                   logger,
		userDeviceService:        userDeviceService,
		vincarioValuationService: vincarioValuationService,
		drivlyValuationService:   drivlyValuationService,
	}
}

func (wls *WorkerListenerService) ProcessWorker(messages <-chan *message.Message) {
	for msg := range messages {
		err := wls.processMessage(msg)
		if err != nil {
			wls.logger.Err(err).Msg("error processing task status message")
		}
	}
}

func (wls *WorkerListenerService) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	event := new(shared.CloudEvent[RunValuationRequest])
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		wls.logger.Warn().Str("payload", string(msg.Payload)).Msg("failed to unmarshall RunValuationRequest payload")
		return errors.Wrap(err, "error parsing RunValuationRequest payload")
	}

	return wls.processEvent(event)
}

func (wls *WorkerListenerService) processEvent(event *shared.CloudEvent[RunValuationRequest]) error {
	var (
		ctx = context.Background()
	)

	switch event.Type {
	case vehicleValuationRequestEventType:
		// improvement here could be to make this a command, but not sure what we gain
		userDevice, err := wls.userDeviceService.GetUserDevice(ctx, event.Data.UserDeviceID)
		if err != nil {
			return err
		}
		// improvement, just check for north american region but need to resolve country to region (devices-api)
		if userDevice.CountryCode == "USA" || userDevice.CountryCode == "CAN" || userDevice.CountryCode == "MEX" {
			status, err := wls.drivlyValuationService.PullValuation(ctx, userDevice.Id, userDevice.DeviceDefinitionId, *userDevice.Vin)
			if err != nil {
				wls.logger.Err(err).Str("vin", *userDevice.Vin).Msg("error pulling drivly data")
			} else {
				wls.logger.Info().Msgf("Drivly   %s vin: %s, country: %s", status, userDevice.Vin, userDevice.CountryCode)
			}
		} else {
			status, err := wls.vincarioValuationService.PullValuation(ctx, userDevice.Id, userDevice.DeviceDefinitionId, *userDevice.Vin)
			if err != nil {
				wls.logger.Err(err).Str("vin", *userDevice.Vin).Msg("error pulling vincario data")
			} else {
				wls.logger.Info().Msgf("Vincario %s vin: %s, country: %s", status, *userDevice.Vin, userDevice.CountryCode)
			}
		}

		return nil
	default:
		return fmt.Errorf("unexpected event type %s", event.Type)
	}
}
