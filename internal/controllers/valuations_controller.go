package controllers

import (
	"encoding/json"

	"github.com/DIMO-Network/valuations-api/internal/controllers/helpers"
	"github.com/DIMO-Network/valuations-api/internal/core/models"
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
// @Tags        user-devices
// @Produce     json
// @Param       userDeviceID path string true "user device id"
// @Success     200 {object} controllers.DeviceValuation
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/valuations [get]
func (vc *ValuationsController) GetValuations(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)
	ud, err := vc.userDeviceService.GetUserDevice(c.Context(), udi)

	if err != nil {
		return err
	}

	if ud.UserId != userID {
		return fiber.NewError(fiber.StatusForbidden, "user does not have access to this vehicle")
	}

	dVal, err := vc.userDeviceService.GetUserDeviceValuations(c.Context(), udi, ud.CountryCode)

	if err != nil {
		return err
	}

	return c.JSON(dVal)
}

// GetOffers godoc
// @Description gets offers for a particular user device
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.DeviceOffer
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/offers [get]
func (vc *ValuationsController) GetOffers(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)
	ud, err := vc.userDeviceService.GetUserDevice(c.Context(), udi)

	if err != nil {
		return err
	}

	if ud.UserId != userID {
		return fiber.NewError(fiber.StatusForbidden, "user does not have access to this vehicle")
	}

	dOffer, err := vc.userDeviceService.GetUserDeviceOffers(c.Context(), udi)

	if err != nil {
		return err
	}

	return c.JSON(dOffer)
}

// GetInstantOffer godoc
// @Description gets instant offer for a particular user device
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.DeviceOffer
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/instant-offer [get]
func (vc *ValuationsController) GetInstantOffer(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	ud, err := vc.userDeviceService.GetUserDevice(c.Context(), udi)

	if err != nil {
		return err
	}

	if ud.UserId != userID {
		return fiber.NewError(fiber.StatusForbidden, "user does not have access to this vehicle")
	}

	request := models.OfferRequest{VIN: *ud.Vin}

	requestBytes, err := json.Marshal(request)

	if err != nil {
		return err
	}

	ack, err := vc.natsService.JetStream.Publish(vc.natsService.OfferSubject, requestBytes)

	if err != nil {
		vc.log.Err(err).Msg("failed to publish offer")
	} else {
		vc.log.Info().Msgf("published offer with id: %v", ack.Sequence)
	}

	return c.JSON(map[string]interface{}{
		"message": "offer request sent",
	})
}
