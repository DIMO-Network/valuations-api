package controllers

import (
	"encoding/json"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"math/big"
	"strconv"
)

type VehiclesController struct {
	log               *zerolog.Logger
	userDeviceService services.UserDeviceAPIService
	natsService       *services.NATSService
}

func NewVehiclesController(log *zerolog.Logger,
	userDeviceSvc services.UserDeviceAPIService,
	natsService *services.NATSService) *VehiclesController {
	return &VehiclesController{
		log:               log,
		userDeviceService: userDeviceSvc,
		natsService:       natsService,
	}
}

// GetValuations godoc
// @Description gets valuations for a particular user device. Includes only price valuations, not offers. only gets the latest valuation.
// @Tags        valuations
// @Produce     json
// @Param 		tokenId path string true "tokenId for vehicle to get offers"
// @Success     200 {object} models.DeviceValuation
// @Security    BearerAuth
// @Router      /vehicles/{tokenId}/valuations [get]
func (vc *VehiclesController) GetValuations(c *fiber.Ctx) error {
	tidStr := c.Params("tokenId")
	tokenID, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}

	ud, err := vc.userDeviceService.GetUserDeviceByTokenID(c.Context(), tokenID)
	if err != nil {
		return err
	}

	takeStr := c.Query("take")
	take, err := strconv.Atoi(takeStr)
	if err != nil || take <= 0 {
		take = 10
	}

	valuation, err := vc.userDeviceService.GetUserDeviceValuationsByTokenID(c.Context(), tokenID, ud.CountryCode, take)
	if err != nil {
		return err
	}

	return c.JSON(valuation)
}

// GetOffers godoc
// @Description gets any existing offers for a particular user device. You must call instant-offer endpoint first to pull.
// @Tags        offers
// @Produce     json
// @Param 		tokenId path string true "tokenId for vehicle to get offers"
// @Success     200 {object} models.DeviceOffer
// @Security    BearerAuth
// @Router      /vehicles/{tokenId}/offers [get]
func (vc *VehiclesController) GetOffers(c *fiber.Ctx) error {
	tidStr := c.Params("tokenId")
	tokenID, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}

	takeStr := c.Query("take")
	take, err := strconv.Atoi(takeStr)
	if err != nil || take <= 0 {
		take = 10
	}

	offer, err := vc.userDeviceService.GetUserDeviceOffersByTokenID(c.Context(), tokenID, take)
	if err != nil {
		return err
	}

	return c.JSON(offer)
}

// RequestInstantOffer godoc
// @Description makes a request for an instant offer for a particular user device. Simply returns success if able to create job.
// @Description You will need to query the offers endpoint to see if a new offer showed up. Job can take about a minute to complete.
// @Tags        offers
// @Produce     json
// @Param 		tokenId path string true "tokenId for vehicle to get offers"
// @Success     200
// @Security    BearerAuth
// @Router      /vehicles/{tokenId}/instant-offer [get]
func (vc *VehiclesController) RequestInstantOffer(c *fiber.Ctx) error {
	tidStr := c.Params("tokenId")
	tokenID, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}

	localLog := vc.log.With().Str("token_id", tidStr).Logger()

	ud, err := vc.userDeviceService.GetUserDeviceByTokenID(c.Context(), tokenID)
	if err != nil {
		vc.log.Err(err).Msg("failed to get user device")
		return err
	}

	canRequestInsantOffer, err := vc.userDeviceService.CanRequestInstantOfferByTokenID(c.Context(), tokenID)
	if err != nil {
		vc.log.Err(err).Msg("failed to check if user can request instant offer")
		return err
	}

	if !canRequestInsantOffer {
		return fiber.NewError(fiber.StatusBadRequest, "already requested in last 30 days")
	}

	didGetErrorLastTime, err := vc.userDeviceService.LastRequestDidGiveErrorByTokenID(c.Context(), tokenID)

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
