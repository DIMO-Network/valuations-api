package controllers

import (
	"math/big"
	"strings"

	"github.com/DIMO-Network/shared/pkg/logfields"
	"github.com/DIMO-Network/valuations-api/internal/core/gateways"
	"github.com/pkg/errors"

	core "github.com/DIMO-Network/valuations-api/internal/core/models"

	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type VehiclesController struct {
	log                  *zerolog.Logger
	userDeviceService    services.UserDeviceAPIService
	drivlyValuationSvc   services.DrivlyValuationService
	vincarioValuationSvc services.VincarioValuationService
	identityAPI          gateways.IdentityAPI
	telemetryAPI         gateways.TelemetryAPI
	locationSvc          services.LocationService
}

func NewVehiclesController(log *zerolog.Logger,
	userDeviceSvc services.UserDeviceAPIService, drivlyValuationSvc services.DrivlyValuationService,
	vincarioValuationSvc services.VincarioValuationService, identityAPI gateways.IdentityAPI,
	telemetryAPI gateways.TelemetryAPI, locationSvc services.LocationService) *VehiclesController {
	return &VehiclesController{
		log:                  log,
		userDeviceService:    userDeviceSvc,
		drivlyValuationSvc:   drivlyValuationSvc,
		vincarioValuationSvc: vincarioValuationSvc,
		identityAPI:          identityAPI,
		telemetryAPI:         telemetryAPI,
		locationSvc:          locationSvc,
	}
}

// GetValuations godoc
// @Description gets valuations for a particular user device. Includes only price valuations, not offers. gets list of most recent
// @Tags        valuations
// @Produce     json
// @Param 		tokenId path string true "tokenId for vehicle to get offers"
// @Success     200 {object} core.DeviceValuation
// @Security    BearerAuth
// @Router      /v2/vehicles/{tokenId}/valuations [get]
func (vc *VehiclesController) GetValuations(c *fiber.Ctx) error {
	tidStr := c.Params("tokenId")
	tokenID, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}
	_, err := vc.identityAPI.GetVehicle(tokenID.Uint64())
	if err != nil {
		return err
	}

	privJWT := c.Get(fiber.HeaderAuthorization)

	//takeStr := c.Query("take")
	//take, err := strconv.Atoi(takeStr)
	//if err != nil || take <= 0 {
	//	take = 10
	//}
	// need to pass in userDeviceId until totally complete migration
	valuation, err := vc.userDeviceService.GetValuations(c.Context(), tokenID.Uint64(), privJWT)
	if err != nil {
		return err
	}

	return c.JSON(valuation)
}

// GetOffers godoc
// @Description gets any existing offers for a particular user device. You must call instant-offer endpoint first to pull newer. Returns list.
// @Tags        offers
// @Produce     json
// @Param 		tokenId path string true "tokenId for vehicle to get offers"
// @Success     200 {object} core.DeviceOffer
// @Security    BearerAuth
// @Router      /v2/vehicles/{tokenId}/offers [get]
func (vc *VehiclesController) GetOffers(c *fiber.Ctx) error {
	tidStr := c.Params("tokenId")
	tokenID, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}
	_, err := vc.identityAPI.GetVehicle(tokenID.Uint64())
	if err != nil {
		return err
	}

	//takeStr := c.Query("take")
	//take, err := strconv.Atoi(takeStr)
	//if err != nil || take <= 0 {
	//	take = 10
	//}
	// todo change below to get list. Make sure that if older than 7 days does not include offer link
	offer, err := vc.userDeviceService.GetOffers(c.Context(), tokenID.Uint64())
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
// @Router      /v2/vehicles/{tokenId}/instant-offer [post]
func (vc *VehiclesController) RequestInstantOffer(c *fiber.Ctx) error {
	tidStr := c.Params("tokenId")
	tokenID, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}

	privJWT := c.Get(fiber.HeaderAuthorization)

	localLog := vc.log.With().Str(logfields.VehicleTokenID, tidStr).Str(logfields.HTTPPath, c.Path()).Logger()

	canRequestInsantOffer, err := vc.userDeviceService.CanRequestInstantOffer(c.Context(), tokenID.Uint64())
	if err != nil {
		localLog.Err(err).Msg("failed to check if user can request instant offer")
		return err
	}

	if !canRequestInsantOffer {
		return fiber.NewError(fiber.StatusBadRequest, "already requested in last 30 days")
	}

	didGetErrorLastTime, err := vc.userDeviceService.LastRequestDidGiveError(c.Context(), tokenID.Uint64())
	if err != nil {
		localLog.Err(err).Msg("failed to check if user can request instant offer")
		return err
	}
	if didGetErrorLastTime {
		return fiber.NewError(fiber.StatusBadRequest, "no offers found for you vehicle in last request")
	}
	var valuationErr error
	var status core.DataPullStatusEnum

	signals, err := vc.telemetryAPI.GetLatestSignals(c.Context(), tokenID.Uint64(), privJWT)
	if err != nil {
		return errors.Wrap(err, "failed to get latest signals for tokenId: "+tidStr)
	}
	location, err := vc.locationSvc.GetGeoDecodedLocation(c.Context(), signals, tokenID.Uint64())
	if err != nil {
		return errors.Wrap(err, "failed to get geo decoded location for tokenId: "+tidStr)
	}
	vinVC, err := vc.telemetryAPI.GetVinVC(c.Context(), tokenID.Uint64(), privJWT)
	if err != nil {
		return errors.Wrap(err, "failed to get vinVC for tokenId: "+tidStr)
	}

	if strings.Contains(services.NorthAmercanCountries, location.CountryCode) {
		status, valuationErr = vc.drivlyValuationSvc.PullOffer(c.Context(), tokenID.Uint64(), vinVC.Vin, privJWT)
	} else {
		status, valuationErr = vc.vincarioValuationSvc.PullValuation(c.Context(), tokenID.Uint64(), vinVC.Vin)
	}
	if valuationErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, valuationErr.Error())
	}
	localLog.Info().Msgf("succesfully requested offer with status %s", status)

	return c.JSON(fiber.Map{
		"message": "instant offer request completed: " + status,
	})
}

// RequestValuationOnly godoc
// @Description request valuation only from drivly or vincario (if drivly fails)
// @Tags        valuations
// @Produce     json
// @Param 		tokenId path string true "tokenId for vehicle to get valuation"
// @Success     200
// @Security    BearerAuth
// @Router      /v2/vehicles/{tokenId}/valuation [post]
func (vc *VehiclesController) RequestValuationOnly(c *fiber.Ctx) error {
	tidStr := c.Params("tokenId")
	tokenID, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}

	privJWT := c.Get(fiber.HeaderAuthorization)

	localLog := vc.log.With().Str(logfields.VehicleTokenID, tidStr).Str(logfields.HTTPPath, c.Path()).Logger()

	var valuationErr error
	var status core.DataPullStatusEnum

	vinVC, err := vc.telemetryAPI.GetVinVC(c.Context(), tokenID.Uint64(), privJWT)
	if err != nil {
		return errors.Wrap(err, "failed to get vinVC for tokenId: "+tidStr)
	}

	status, valuationErr = vc.drivlyValuationSvc.PullValuation(c.Context(), tokenID.Uint64(), vinVC.Vin, privJWT)
	if valuationErr != nil {
		localLog.Err(valuationErr).Msg("failed to get valuation from drivly, retrying with vincario")
		status, valuationErr = vc.vincarioValuationSvc.PullValuation(c.Context(), tokenID.Uint64(), vinVC.Vin)
	}
	if valuationErr != nil {
		localLog.Err(valuationErr).Msg("failed to get valuation from vincario")
		return fiber.NewError(fiber.StatusInternalServerError, valuationErr.Error())
	}
	localLog.Info().Msgf("succesfully requested offer with status %s", status)

	return c.JSON(fiber.Map{
		"message": "valuation request completed: " + status,
	})
}
