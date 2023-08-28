package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/DIMO-Network/devices-api/pkg/grpc"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	mock_services "github.com/DIMO-Network/valuations-api/internal/core/services/mocks"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/dbtest"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const userID = "2TqxFTIQPZ3gnUPi3Pdb3eEZDx4"

type ValuationsControllerTestSuite struct {
	suite.Suite
	controller    *ValuationsController
	ctx           context.Context
	mockCtrl      *gomock.Controller
	app           *fiber.App
	userDeviceSvc *mock_services.MockUserDeviceAPIService
}

// SetupSuite starts container db
func (s *ValuationsControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	logger := dbtest.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl
	s.userDeviceSvc = mock_services.NewMockUserDeviceAPIService(mockCtrl)
	var err error

	if err != nil {
		s.T().Fatal(err)
	}

	controller := NewValuationsController(logger, s.userDeviceSvc)
	app := dbtest.SetupAppFiber(*logger)
	app.Get("/user/devices/:userDeviceID/offers", dbtest.AuthInjectorTestHandler(userID), controller.GetOffers)
	app.Get("/user/devices/:userDeviceID/valuations", dbtest.AuthInjectorTestHandler(userID), controller.GetValuations)
	s.controller = controller

	s.app = app
}

func (s *ValuationsControllerTestSuite) SetupTest() {

}

// TearDownTest after each test truncate tables
func (s *ValuationsControllerTestSuite) TearDownTest() {
}

// TearDownSuite cleanup at end by terminating container
func (s *ValuationsControllerTestSuite) TearDownSuite() {
	s.mockCtrl.Finish() // might need to do mockctrl on every test, and refactor setup into one method
}

// Test Runner
func TestValuationsControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ValuationsControllerTestSuite))
}

func (s *ValuationsControllerTestSuite) TestGetDeviceValuations_Format1() {
	// arrange db, insert some user_devices
	udID := ksuid.New().String()
	vin := "vinny"

	s.userDeviceSvc.EXPECT().GetUserDevice(gomock.Any(), udID).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userID,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	s.userDeviceSvc.EXPECT().GetUserDeviceValuations(gomock.Any(), udID, "USA").Return(&core.DeviceValuation{
		ValuationSets: []core.ValuationSet{
			{
				Vendor:           "vincario",
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
				Currency:         "EUR",
			},
		},
	}, nil)

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", udID), "")
	response, _ := s.app.Test(request)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}

func (s *ValuationsControllerTestSuite) TestGetDeviceValuations_Format2() {
	// this is the other format we're seeing coming from drivly for pricing
	// arrange db, insert some user_devices
	udID := ksuid.New().String()
	vin := "vinny"

	s.userDeviceSvc.EXPECT().GetUserDevice(gomock.Any(), udID).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userID,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	s.userDeviceSvc.EXPECT().GetUserDeviceValuations(gomock.Any(), udID, "USA").Return(&core.DeviceValuation{
		ValuationSets: []core.ValuationSet{
			{
				Vendor:           "vincario",
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
				Currency:         "EUR",
			},
		},
	}, nil)

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", udID), "")
	response, _ := s.app.Test(request)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}

func (s *ValuationsControllerTestSuite) TestGetDeviceValuations_Vincario() {
	// this is the other format we're seeing coming from drivly for pricing
	// arrange db, insert some user_devices
	udID := ksuid.New().String()
	vin := "vinny"

	s.userDeviceSvc.EXPECT().GetUserDevice(gomock.Any(), udID).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userID,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	// ? TODO: check how to retrieve this from db in configuration
	s.userDeviceSvc.EXPECT().GetUserDeviceValuations(gomock.Any(), udID, "USA").Return(&core.DeviceValuation{
		ValuationSets: []core.ValuationSet{
			{
				Vendor:           "vincario",
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
				Currency:         "EUR",
			},
		},
	}, nil)

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", udID), "")
	response, _ := s.app.Test(request, 2000)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}

func (s *ValuationsControllerTestSuite) TestGetDeviceOffers() {

	udID := ksuid.New().String()
	vin := "vinny"

	s.userDeviceSvc.EXPECT().GetUserDevice(gomock.Any(), udID).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userID,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	// ? TODO: check how to retrieve this from db in configuration
	s.userDeviceSvc.EXPECT().GetUserDeviceOffers(gomock.Any(), udID).Return(&core.DeviceOffer{
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

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/offers", udID), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	require.Equal(s.T(), fiber.StatusOK, response.StatusCode)
}
