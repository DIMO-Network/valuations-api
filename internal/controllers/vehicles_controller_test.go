package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	mock_gateways "github.com/DIMO-Network/valuations-api/internal/core/gateways/mocks"

	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"

	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	mock_services "github.com/DIMO-Network/valuations-api/internal/core/services/mocks"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/dbtest"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	userID = "user123123"
)

type VehiclesControllerTestSuite struct {
	suite.Suite
	controller           *VehiclesController
	ctx                  context.Context
	mockCtrl             *gomock.Controller
	app                  *fiber.App
	userDeviceSvc        *mock_services.MockUserDeviceAPIService
	drivlyValuationSvc   *mock_services.MockDrivlyValuationService
	vincarioValuationSvc *mock_services.MockVincarioValuationService
	identity             *mock_gateways.MockIdentityAPI
	telemetry            *mock_gateways.MockTelemetryAPI
	locationSvc          *mock_services.MockLocationService
}

// SetupSuite starts container db
func (s *VehiclesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	logger := dbtest.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl
	s.userDeviceSvc = mock_services.NewMockUserDeviceAPIService(mockCtrl)
	s.drivlyValuationSvc = mock_services.NewMockDrivlyValuationService(mockCtrl)
	s.vincarioValuationSvc = mock_services.NewMockVincarioValuationService(mockCtrl)
	s.identity = mock_gateways.NewMockIdentityAPI(mockCtrl)
	s.telemetry = mock_gateways.NewMockTelemetryAPI(mockCtrl)
	s.locationSvc = mock_services.NewMockLocationService(mockCtrl)

	controller := NewVehiclesController(logger, s.userDeviceSvc, s.drivlyValuationSvc, s.vincarioValuationSvc, s.identity, s.telemetry, s.locationSvc)
	app := dbtest.SetupAppFiber(*logger)
	app.Get("/vehicles/:tokenID/offers", dbtest.AuthInjectorTestHandler(userID), controller.GetOffers)
	app.Get("/vehicles/:tokenID/valuations", dbtest.AuthInjectorTestHandler(userID), controller.GetValuations)
	app.Post("/vehicles/:tokenID/valuations", dbtest.AuthInjectorTestHandler(userID), controller.RequestValuationOnly)
	s.controller = controller

	s.app = app
}

func (s *VehiclesControllerTestSuite) SetupTest() {

}

// TearDownTest after each test truncate tables
func (s *VehiclesControllerTestSuite) TearDownTest() {
}

// TearDownSuite cleanup at end by terminating container
func (s *VehiclesControllerTestSuite) TearDownSuite() {
	s.mockCtrl.Finish() // might need to do mockctrl on every test, and refactor setup into one method
}

// Test Runner
func TestVehiclesControllerTestSuite(t *testing.T) {
	suite.Run(t, new(VehiclesControllerTestSuite))
}

func (s *VehiclesControllerTestSuite) TestPostRequestValuationOnly_Drivly1() {
	// arrange db, insert some user_devices
	tokenID := uint64(12345)
	vin := "vinny"

	s.telemetry.EXPECT().GetVinVC(gomock.Any(), tokenID, gomock.Any()).Return(&core.VinVCLatest{
		Vin:         vin,
		CountryCode: "USA",
	}, nil)
	s.drivlyValuationSvc.EXPECT().PullValuation(gomock.Any(), tokenID, vin, "").
		Return(core.PulledValuationDrivlyStatus, nil)

	request := dbtest.BuildRequest("POST", fmt.Sprintf("/vehicles/%d/valuations", tokenID), "")
	response, _ := s.app.Test(request)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}

func (s *VehiclesControllerTestSuite) TestGetValuations_Drivly2() {
	tokenID := uint64(12345)

	s.identity.EXPECT().GetVehicle(tokenID).Return(&core.Vehicle{
		ID: "xxx",
		Definition: struct {
			ID    string `json:"id"`
			Make  string `json:"make"`
			Model string `json:"model"`
			Year  int    `json:"year"`
		}{ID: "ford_escape_2022"},
		Owner: "0x123",
	}, nil)

	s.userDeviceSvc.EXPECT().GetValuations(gomock.Any(), tokenID, gomock.Any()).Return(&core.DeviceValuation{
		ValuationSets: []core.ValuationSet{
			{
				Vendor:           "drivly",
				Updated:          "",
				Mileage:          30137,
				ZipCode:          "",
				TradeInSource:    "",
				TradeIn:          44800,
				TradeInClean:     0,
				TradeInAverage:   0,
				TradeInRough:     0,
				RetailSource:     "",
				Retail:           55200,
				RetailClean:      0,
				RetailAverage:    0,
				RetailRough:      0,
				OdometerUnit:     "km",
				Odometer:         30137,
				UserDisplayPrice: 51440,
				Currency:         "USD",
			},
		},
	}, nil)

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/vehicles/%d/valuations", tokenID), "")
	response, _ := s.app.Test(request)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}

func (s *VehiclesControllerTestSuite) TestGetOffers() {

	tokenID := uint64(12345)

	s.identity.EXPECT().GetVehicle(tokenID).Return(&core.Vehicle{
		ID: "xxx",
		Definition: struct {
			ID    string `json:"id"`
			Make  string `json:"make"`
			Model string `json:"model"`
			Year  int    `json:"year"`
		}{ID: "ford_escape_2022"},
		Owner: "0x123",
	}, nil)

	s.userDeviceSvc.EXPECT().GetOffers(gomock.Any(), tokenID).Return(&core.DeviceOffer{
		OfferSets: []core.OfferSet{
			{
				Source:  "drivly",
				Updated: "",
				Mileage: 0,
				ZipCode: "",
				Offers:  []core.Offer{{Vendor: "vroom", Price: 10123, Error: "Error in v1/acquisition/appraisal POST", DeclineReason: ""}, {Vendor: "carvana", Price: 10123, URL: "", Error: "", Grade: "", DeclineReason: "Make[Ford],Model[Mustang Mach-E],Year[2022] is not eligible for offer."}, {Vendor: "carmax", Price: 10123, DeclineReason: "", Error: "Error in v1/acquisition/appraisal POST"}},
			},
		},
	}, nil)

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/vehicles/%d/offers", tokenID), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	require.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}
