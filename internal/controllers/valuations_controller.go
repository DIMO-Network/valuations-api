package controllers

import (
	"sync"
	"time"

	"github.com/DIMO-Network/valuations-api/internal/controllers/helpers"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type ValuationsController struct {
	log               *zerolog.Logger
	userDeviceService services.UserDeviceAPIService
	drivlyService     services.DrivlyAPIService
}

func NewValuationsController(log *zerolog.Logger, userDeviceSvc services.UserDeviceAPIService, drivlyService services.DrivlyAPIService) *ValuationsController {
	return &ValuationsController{
		log:               log,
		userDeviceService: userDeviceSvc,
		drivlyService:     drivlyService,
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
	// long polling call to drivly api to get instant offer
	// return instant offer
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	ud, err := vc.userDeviceService.GetUserDevice(c.Context(), udi)

	if err != nil {
		return err
	}

	if ud.UserId != userID {
		return fiber.NewError(fiber.StatusForbidden, "user does not have access to this vehicle")
	}

	wg := &sync.WaitGroup{}
	m := &sync.RWMutex{}
	ch := make(chan map[string]interface{})

	wg.Add(1)
	go func(ch chan map[string]interface{}, vin string, wg *sync.WaitGroup, m *sync.RWMutex) {
		params := services.ValuationRequestData{}
		offer, err := vc.drivlyService.GetOffersByVIN(*ud.Vin, &params)

		if err != nil {
			m.Lock()
			ch <- map[string]interface{}{
				"error": err,
			}
			m.Unlock()
		}

		if offer != nil {
			m.Lock()
			ch <- map[string]interface{}{
				"offer": offer,
			}
			m.Unlock()
		}

		wg.Done()
	}(ch, *ud.Vin, wg, m)

	wg.Wait()

	select {
	case offer := <-ch:
		return c.JSON(offer)
	case <-time.After(10 * time.Second):
		return fiber.NewError(fiber.StatusRequestTimeout, "request timed out")
	}
}
