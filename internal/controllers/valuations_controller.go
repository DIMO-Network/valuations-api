package controllers

import (
	"encoding/json"

	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type ValuationsController struct {
	log               *zerolog.Logger
	userDeviceService services.UserDeviceAPIService
	natsService       *services.NATSService
}

func NewValuationsController(log *zerolog.Logger,
	userDeviceSvc services.UserDeviceAPIService,
	natsService *services.NATSService) *ValuationsController {
	return &ValuationsController{
		log:               log,
		userDeviceService: userDeviceSvc,
		natsService:       natsService,
	}
}

// GetValuations godoc
// @Description gets valuations for a particular user device. Includes only price valuations, not offers. only gets the latest valuation.
// @Tags        valuations
// @Produce     json
// @Param 		userDeviceID path string true "userDeviceID for vehicle to get offers"
// @Success     200 {object} core.DeviceValuation
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/valuations [get]
func (vc *ValuationsController) GetValuations(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	ud, err := vc.userDeviceService.GetUserDevice(c.Context(), udi)
	if err != nil {
		return err
	}

	valuation, err := vc.userDeviceService.GetUserDeviceValuations(c.Context(), udi, ud.CountryCode)
	if err != nil {
		return err
	}

	return c.JSON(valuation)
}

// GetOffers godoc
// @Description gets any existing offers for a particular user device. You must call instant-offer endpoint first to pull.
// @Tags        offers
// @Produce     json
// @Param 		userDeviceID path string true "userDeviceID for vehicle to get offers"
// @Success     200 {object} core.DeviceOffer
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/offers [get]
func (vc *ValuationsController) GetOffers(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	offer, err := vc.userDeviceService.GetUserDeviceOffers(c.Context(), udi)
	if err != nil {
		return err
	}

	return c.JSON(offer)
}

// GetInstantOffer godoc
// @Description makes a request for an instant offer for a particular user device. Simply returns success if able to create job.
// @Description You will need to query the offers endpoint to see if a new offer showed up. Job can take about a minute to complete.
// @Tags        offers
// @Produce     json
// @Param 		userDeviceID path string true "userDeviceID for vehicle to get offers"
// @Success     200
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/instant-offer [get]
func (vc *ValuationsController) GetInstantOffer(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")

	localLog := vc.log.With().Str("user_device_id", udi).Logger()

	ud, err := vc.userDeviceService.GetUserDevice(c.Context(), udi)
	if err != nil {
		vc.log.Err(err).Msg("failed to get user device")
		return err
	}

	canRequestInsantOffer, err := vc.userDeviceService.CanRequestInstantOffer(c.Context(), ud.Id)
	if err != nil {
		vc.log.Err(err).Msg("failed to check if user can request instant offer")
		return err
	}

	if !canRequestInsantOffer {
		return fiber.NewError(fiber.StatusBadRequest, "already requested in last 7 days")
	}

	didGetErrorLastTime, err := vc.userDeviceService.LastRequestDidGiveError(c.Context(), ud.Id)

	if err != nil {
		vc.log.Err(err).Msg("failed to check if user can request instant offer")
		return err
	}

	if didGetErrorLastTime {
		return fiber.NewError(fiber.StatusBadRequest, "no offers found for you vehicle in last request")
	}

	request := core.OfferRequest{UserDeviceID: ud.Id}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		localLog.Err(err).Msg("failed to marshal offer request")
		return err
	}

	ack, err := vc.natsService.JetStream.Publish(vc.natsService.OfferSubject, requestBytes)
	if err != nil {
		localLog.Err(err).Msg("failed to publish offer request")
	} else {
		localLog.Info().Msgf("published offer request with id: %v", ack.Sequence)
	}

	return c.JSON(fiber.Map{
		"message": "instant offer request sent",
	})
}
