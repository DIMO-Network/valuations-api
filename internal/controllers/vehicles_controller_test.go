package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"

	"github.com/DIMO-Network/devices-api/pkg/grpc"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	mock_services "github.com/DIMO-Network/valuations-api/internal/core/services/mocks"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/dbtest"
	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
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

	var err error

	if err != nil {
		s.T().Fatal(err)
	}

	controller := NewVehiclesController(logger, s.userDeviceSvc, s.drivlyValuationSvc, s.vincarioValuationSvc)
	app := dbtest.SetupAppFiber(*logger)
	app.Get("/vehicles/:tokenID/offers", dbtest.AuthInjectorTestHandler(userID), controller.GetOffers)
	app.Get("/vehicles/:tokenID/valuations", dbtest.AuthInjectorTestHandler(userID), controller.GetValuations)
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

func (s *VehiclesControllerTestSuite) TestGetDeviceValuations_Drivly1() {
	// arrange db, insert some user_devices
	tokenID := "1234567890"
	udID := ksuid.New().String()
	vin := "vinny"

	s.userDeviceSvc.EXPECT().GetUserDeviceByTokenID(gomock.Any(), gomock.Any()).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userID,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	s.userDeviceSvc.EXPECT().GetUserDeviceValuationsByTokenID(gomock.Any(), gomock.Any(), "USA", 10, udID).Return(&core.DeviceValuation{
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

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/vehicles/%s/valuations", tokenID), "")
	response, _ := s.app.Test(request)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}

func (s *VehiclesControllerTestSuite) TestGetDeviceValuations_Drivly2() {
	// this is the other format we're seeing coming from drivly for pricing
	// arrange db, insert some user_devices
	tokenID := "1234567890"
	udID := ksuid.New().String()
	vin := "vinny"

	s.userDeviceSvc.EXPECT().GetUserDeviceByTokenID(gomock.Any(), gomock.Any()).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userID,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	s.userDeviceSvc.EXPECT().GetUserDeviceValuationsByTokenID(gomock.Any(), gomock.Any(), "USA", 10, udID).Return(&core.DeviceValuation{
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

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/vehicles/%s/valuations", tokenID), "")
	response, _ := s.app.Test(request)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}

func (s *VehiclesControllerTestSuite) TestGetDeviceOffers() {

	tokenID := "1234567890"
	udID := ksuid.New().String()
	vin := "vinny"

	s.userDeviceSvc.EXPECT().GetUserDeviceByTokenID(gomock.Any(), gomock.Any()).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userID,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	s.userDeviceSvc.EXPECT().GetUserDeviceOffersByTokenID(gomock.Any(), gomock.Any(), 10, udID).Return(&core.DeviceOffer{
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

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/vehicles/%s/offers", tokenID), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	require.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}
