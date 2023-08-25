package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source run_valuation.go -destination mocks/run_valuation_mock.go

type RunValuationCommandHandler interface {
	Execute(ctx context.Context) error
}

const NorthAmercanCountries = "USA,CAN,MEX,PRI"

type runValuationCommandHandler struct {
	DBS                      func() *db.ReaderWriter
	logger                   zerolog.Logger
	userDeviceService        services.UserDeviceAPIService
	NATSSvc                  *services.NATSService
	vincarioValuationService services.VincarioValuationService
	drivlyValuationService   services.DrivlyValuationService
}

func NewRunValuationCommandHandler(dbs func() *db.ReaderWriter, logger zerolog.Logger, settings *config.Settings,
	userDeviceService services.UserDeviceAPIService,
	ddSvc services.DeviceDefinitionsAPIService,
	uddSvc services.UserDeviceDataAPIService,
	natsSvc *services.NATSService) RunValuationCommandHandler {
	return &runValuationCommandHandler{
		DBS:                      dbs,
		logger:                   logger,
		userDeviceService:        userDeviceService,
		vincarioValuationService: services.NewVincarioValuationService(dbs, &logger, settings, userDeviceService),
		drivlyValuationService:   services.NewDrivlyValuationService(dbs, &logger, settings, ddSvc, uddSvc, userDeviceService),
		NATSSvc:                  natsSvc,
	}
}

type RunValuationCommandRequest struct {
	VIN          string `json:"vin"`
	UserDeviceID string `json:"userDeviceId"`
}

type RunTestSignalCommandResponse struct {
}

func (h *runValuationCommandHandler) Execute(ctx context.Context) error {
	sub, err := h.NATSSvc.JetStream.PullSubscribe(h.NATSSvc.JetStreamSubject, h.NATSSvc.DurableConsumer, nats.AckWait(h.NATSSvc.AckTimeout))

	if err != nil {
		return err
	}
	localLog := h.logger.With().Str("func", "RunValuation.Execute").Logger()

	for {
		msgs, err := sub.Fetch(1, nats.MaxWait(h.NATSSvc.AckTimeout))
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}

			return err
		}

		for _, msg := range msgs {
			select {
			case <-ctx.Done():
				return nil
			default:
				err := h.processMessage(ctx, localLog, msg)
				if err != nil {
					h.nak(msg)
					localLog.Err(err).Str("payload", string(msg.Data)).Msg("failed to process valuation request")
					continue
				}
				if err := msg.Ack(); err != nil {
					localLog.Err(err).Msg("message ack failed")
				}
			}
		}
	}
}

func (h *runValuationCommandHandler) nak(msg *nats.Msg) {
	err := msg.Nak()
	if err != nil {
		h.logger.Err(err).Msg("message nak failed")
	}
}

// processMessage handles the logic to run a valuation request.
func (h *runValuationCommandHandler) processMessage(ctx context.Context, localLog zerolog.Logger, msg *nats.Msg) error {
	localLog.Info().Str("payload", string(msg.Data)).Msgf("processing valuation request message with subject %s", msg.Subject)

	var valuationDecode RunValuationCommandRequest
	mtd, err := msg.Metadata()
	numDelivered := uint64(0)
	if err != nil {
		localLog.Warn().Err(err).Msg("unable to parse metadata for message")
	} else {
		numDelivered = mtd.NumDelivered
	}
	if err := json.Unmarshal(msg.Data, &valuationDecode); err != nil {
		return errors.Wrap(err, "unable to parse vin from message")
	}
	localLog = localLog.With().Str("vin", valuationDecode.VIN).Uint64("num_delivered", numDelivered).
		Str("user_device_id", valuationDecode.UserDeviceID).Logger()

	userDevice, err := h.userDeviceService.GetUserDevice(ctx, valuationDecode.UserDeviceID)
	if err != nil {
		return errors.Wrap(err, "unable to find user device. udId: "+valuationDecode.UserDeviceID)
	}
	// note that the VIN in user device may likely be empty at this point, not sure why but lets just use the payload one
	if len(valuationDecode.VIN) == 0 {
		return fmt.Errorf("VIN is empty from message payload. udId: %s. payload object %+v", valuationDecode.UserDeviceID, valuationDecode)
	}
	localLog = localLog.With().Str("country", userDevice.CountryCode).Str("device_definition_id", userDevice.DeviceDefinitionId).Logger()

	_ = msg.InProgress() // ignore err if can't set to in progress

	if strings.Contains(NorthAmercanCountries, userDevice.CountryCode) {
		status, err := h.drivlyValuationService.PullValuation(ctx, userDevice.Id, userDevice.DeviceDefinitionId, valuationDecode.VIN)
		if err != nil {
			localLog.Err(err).Msg("valuation request - error pulling drivly data")
		} else {
			localLog.Info().Msgf("valuation request from Drivly completed with status %s", status)
		}
	} else {
		status, err := h.vincarioValuationService.PullValuation(ctx, userDevice.Id, userDevice.DeviceDefinitionId, valuationDecode.VIN)
		if err != nil {
			localLog.Err(err).Msg("valuation request - error pulling vincario data")
		} else {
			localLog.Info().Msgf("valuation request from Vincario completed with status %s", status)
		}
	}

	return nil
}
