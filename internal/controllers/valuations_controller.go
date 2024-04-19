package controllers

import (
	"strings"

	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type ValuationsController struct {
	log               *zerolog.Logger
	userDeviceService services.UserDeviceAPIService
	drivlySvc         services.DrivlyValuationService
	vincarioSvc       services.VincarioValuationService
}

func NewValuationsController(log *zerolog.Logger,
	userDeviceSvc services.UserDeviceAPIService,
	drivlySvc services.DrivlyValuationService, vincarioSvc services.VincarioValuationService) *ValuationsController {
	return &ValuationsController{
		log:               log,
		userDeviceService: userDeviceSvc,
		drivlySvc:         drivlySvc,
		vincarioSvc:       vincarioSvc,
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
		localLog.Err(err).Msg("failed to get user device")
		return err
	}
	if ud.TokenId == nil {
		return fiber.NewError(fiber.StatusBadRequest, "your vehicle is not minted or not setup correctly - missing tokenId")
	}

	canRequestInsantOffer, err := vc.userDeviceService.CanRequestInstantOffer(c.Context(), ud.Id)
	if err != nil {
		localLog.Err(err).Msg("failed to check if user can request instant offer")
		return err
	}

	if !canRequestInsantOffer {
		return fiber.NewError(fiber.StatusBadRequest, "already requested in last 7 days")
	}

	didGetErrorLastTime, err := vc.userDeviceService.LastRequestDidGiveError(c.Context(), ud.Id)

	if err != nil {
		localLog.Err(err).Msg("failed to check if user can request instant offer")
		return err
	}

	if didGetErrorLastTime {
		return fiber.NewError(fiber.StatusBadRequest, "no offers found for you vehicle in last request")
	}
	var valuationErr error
	var status core.DataPullStatusEnum
	// this used to be async with nats, but trying just making it syncronous since doesn't really take that long, more now that vroom disabled.
	if strings.Contains(services.NorthAmercanCountries, ud.CountryCode) {
		status, valuationErr = vc.drivlySvc.PullOffer(c.Context(), ud.Id, *ud.TokenId, *ud.Vin)
	} else {
		status, valuationErr = vc.vincarioSvc.PullValuation(c.Context(), ud.Id, *ud.TokenId, ud.DeviceDefinitionId, *ud.Vin)
	}
	if valuationErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, valuationErr.Error())
	}

	return c.JSON(fiber.Map{
		"message": "instant offer request completed: " + status,
	})
}
