package commands

import (
	"context"
	"encoding/json"
	"strings"

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
	udSvc services.UserDeviceService,
	natsSvc *services.NATSService) RunValuationCommandHandler {
	return &runValuationCommandHandler{
		DBS:                      dbs,
		logger:                   logger,
		userDeviceService:        userDeviceService,
		vincarioValuationService: services.NewVincarioValuationService(dbs, &logger, settings, userDeviceService),
		drivlyValuationService:   services.NewDrivlyValuationService(dbs, &logger, settings, ddSvc, uddSvc, udSvc),
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
			mtd, err := msg.Metadata()

			if err != nil {
				h.nak(msg, nil)
				localLog.Info().Err(err).Msg("unable to parse metadata for message")
				continue
			}

			select {
			case <-ctx.Done():
				return nil
			default:

				var valuationDecode RunValuationCommandRequest

				if err := json.Unmarshal(msg.Data, &valuationDecode); err != nil {
					h.nak(msg, &valuationDecode)
					localLog.Info().Err(err).Msg("unable to parse vin from message")
					continue
				}

				userDevice, err := h.userDeviceService.GetUserDevice(ctx, valuationDecode.UserDeviceID)

				if err != nil && userDevice.Vin == &valuationDecode.VIN {
					h.nak(msg, &valuationDecode)
					localLog.Info().Err(err).Msg("unable to find user device")
					continue
				}

				h.inProgress(msg)

				if strings.Contains(NorthAmercanCountries, userDevice.CountryCode) {
					status, err := h.drivlyValuationService.PullValuation(ctx, userDevice.Id, userDevice.DeviceDefinitionId, *userDevice.Vin)
					if err != nil {
						localLog.Err(err).Str("vin", *userDevice.Vin).Msg("valuation request - error pulling drivly data")
					} else {
						localLog.Info().Msgf("valuation request - Drivly %s vin: %s, country: %s", status, *userDevice.Vin, userDevice.CountryCode)
					}
				} else {
					status, err := h.vincarioValuationService.PullValuation(ctx, userDevice.Id, userDevice.DeviceDefinitionId, *userDevice.Vin)
					if err != nil {
						localLog.Err(err).Str("vin", *userDevice.Vin).Msg("valuation request - error pulling vincario data")
					} else {
						localLog.Info().Msgf("valuation request - Vincario %s vin: %s, country: %s", status, *userDevice.Vin, userDevice.CountryCode)
					}
				}

				if err := msg.Ack(); err != nil {
					localLog.Err(err).Msg("message ack failed")
				}

				localLog.Info().Str("vin", valuationDecode.VIN).
					Str("user_device_id", valuationDecode.UserDeviceID).
					Uint64("numDelivered", mtd.NumDelivered).Msg("valuation request completed")
			}
		}
	}
}

func (h *runValuationCommandHandler) inProgress(msg *nats.Msg) {
	if err := msg.InProgress(); err != nil {
		h.logger.Err(err).Msg("message in progress failed")
	}
}

func (h *runValuationCommandHandler) nak(msg *nats.Msg, params *RunValuationCommandRequest) {
	err := msg.Nak()
	if params == nil {
		h.logger.Err(err).Msg("message nak failed")
	} else {
		h.logger.Err(err).Str("vin", params.VIN).Str("user_device_id", params.UserDeviceID).Msg("message nak failed")
	}
}
