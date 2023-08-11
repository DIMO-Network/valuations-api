package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	mock_services "github.com/DIMO-Network/valuations-api/internal/core/services/mocks"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/dbtest"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"io"
	"testing"
)

const migrationsDirRelPath = "../infrastructure/db/migrations"
const userId = "2TqxFTIQPZ3gnUPi3Pdb3eEZDx4"

type ValuationsControllerTestSuite struct {
	suite.Suite
	pdb           db.Store
	controller    *ValuationsController
	container     testcontainers.Container
	ctx           context.Context
	mockCtrl      *gomock.Controller
	app           *fiber.App
	testUserID    string
	userDeviceSvc *mock_services.MockUserDeviceAPIService
}

// SetupSuite starts container db
func (s *ValuationsControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = dbtest.StartContainerDatabase(s.ctx, "valuations_api", s.T(), migrationsDirRelPath)
	logger := dbtest.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl
	s.userDeviceSvc = mock_services.NewMockUserDeviceAPIService(mockCtrl)
	var err error

	if err != nil {
		s.T().Fatal(err)
	}

	controller := NewValuationsController(logger, s.pdb.DBS, s.userDeviceSvc)
	app := dbtest.SetupAppFiber(*logger)
	app.Get("/user/devices/:userDeviceID/offers", dbtest.AuthInjectorTestHandler(userId), controller.GetOffers)
	app.Get("/user/devices/:userDeviceID/valuations", dbtest.AuthInjectorTestHandler(userId), controller.GetValuations)
	s.controller = controller

	s.app = app
}

func (s *ValuationsControllerTestSuite) SetupTest() {

}

// TearDownTest after each test truncate tables
func (s *ValuationsControllerTestSuite) TearDownTest() {
	dbtest.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *ValuationsControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish() // might need to do mockctrl on every test, and refactor setup into one method
}

// Test Runner
func TestValuationsControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ValuationsControllerTestSuite))
}

//go:embed test_drivly_pricing_by_vin.json
var testDrivlyPricingJSON string

//go:embed test_drivly_pricing2.json
var testDrivlyPricing2JSON string

//go:embed test_vincario_valuation.json
var testVincarioValuationJSON string

func (s *ValuationsControllerTestSuite) TestGetDeviceValuations_Format1() {
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"

	_ = SetupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"PricingMetadata": []byte(testDrivlyPricingJSON),
	}, s.pdb)

	s.userDeviceSvc.EXPECT().GetUserDevice(gomock.Any(), udID).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userId,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", udID), "")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 1, int(gjson.GetBytes(body, "valuationSets.#").Int()))
	assert.Equal(s.T(), 49957, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).mileage").Int()))
	assert.Equal(s.T(), 49957, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).odometer").Int()))
	assert.Equal(s.T(), "miles", gjson.GetBytes(body, "valuationSets.#(vendor=drivly).odometerUnit").String())
	assert.Equal(s.T(), 54123, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).retail").Int()))
	//54123 + 50151 / 2
	assert.Equal(s.T(), 52137, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).userDisplayPrice").Int()))
	assert.Equal(s.T(), "USD", gjson.GetBytes(body, "valuationSets.#(vendor=drivly).currency").String())
	// 49040 + 52173 + 49241 / 3 = 50151
	assert.Equal(s.T(), 50151, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).tradeIn").Int()))
	assert.Equal(s.T(), 50151, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).tradeInAverage").Int()))
}

func (s *ValuationsControllerTestSuite) TestGetDeviceValuations_Format2() {
	// this is the other format we're seeing coming from drivly for pricing
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"
	_ = SetupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"PricingMetadata": []byte(testDrivlyPricing2JSON),
	}, s.pdb)
	s.userDeviceSvc.EXPECT().GetUserDevice(gomock.Any(), udID).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userId,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", udID), "")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 1, int(gjson.GetBytes(body, "valuationSets.#").Int()))
	// mileage comes from request metadata, but it is also sometimes returned by payload
	assert.Equal(s.T(), 50702, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).mileage").Int()))
	assert.Equal(s.T(), 40611, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).tradeIn").Int()))
	assert.Equal(s.T(), 50803, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).retail").Int()))
}

func (s *ValuationsControllerTestSuite) TestGetDeviceValuations_Vincario() {
	// this is the other format we're seeing coming from drivly for pricing
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"
	_ = SetupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"VincarioMetadata": []byte(testVincarioValuationJSON),
	}, s.pdb)
	s.userDeviceSvc.EXPECT().GetUserDevice(gomock.Any(), udID).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userId,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", udID), "")
	response, _ := s.app.Test(request, 2000)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 1, int(gjson.GetBytes(body, "valuationSets.#").Int()))
	// mileage comes from request metadata, but it is also sometimes returned by payload
	assert.Equal(s.T(), 30137, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).mileage").Int()))
	assert.Equal(s.T(), 30137, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).odometer").Int()))
	assert.Equal(s.T(), "km", gjson.GetBytes(body, "valuationSets.#(vendor=vincario).odometerUnit").String())
	assert.Equal(s.T(), "EUR", gjson.GetBytes(body, "valuationSets.#(vendor=vincario).currency").String())

	assert.Equal(s.T(), 44800, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).tradeIn").Int()))
	assert.Equal(s.T(), 55200, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).retail").Int()))
	assert.Equal(s.T(), 51440, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).userDisplayPrice").Int()))
}

//go:embed test_drivly_offers_by_vin.json
var testDrivlyOffersJSON string

func (s *ValuationsControllerTestSuite) TestGetDeviceOffers() {
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"
	_ = SetupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"OfferMetadata": []byte(testDrivlyOffersJSON),
	}, s.pdb)
	s.userDeviceSvc.EXPECT().GetUserDevice(gomock.Any(), udID).Return(&grpc.UserDevice{
		Id:           udID,
		UserId:       userId,
		VinConfirmed: true,
		Vin:          &vin,
		CountryCode:  "USA",
	}, nil)

	request := dbtest.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/offers", udID), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	require.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), 1, int(gjson.GetBytes(body, "offerSets.#").Int()))
	assert.Equal(s.T(), "drivly", gjson.GetBytes(body, "offerSets.0.source").String())
	assert.Equal(s.T(), 3, int(gjson.GetBytes(body, "offerSets.0.offers.#").Int()))
	assert.Equal(s.T(), "Error in v1/acquisition/appraisal POST",
		gjson.GetBytes(body, "offerSets.0.offers.#(vendor=vroom).error").String())
	assert.Equal(s.T(), 10123, int(gjson.GetBytes(body, "offerSets.0.offers.#(vendor=carvana).price").Int()))
	assert.Equal(s.T(), "Make[Ford],Model[Mustang Mach-E],Year[2022] is not eligible for offer.",
		gjson.GetBytes(body, "offerSets.0.offers.#(vendor=carmax).declineReason").String())
}

func SetupCreateValuationsData(t *testing.T, ddID, userDeviceID, vin string, md map[string][]byte, pdb db.Store) *models.Valuation {
	val := models.Valuation{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(ddID),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
		RequestMetadata:    null.JSONFrom([]byte(`{"mileage":49957,"zipCode":"48216"}`)), // default request metadata
	}
	if rmd, ok := md["RequestMetadata"]; ok {
		val.RequestMetadata = null.JSONFrom(rmd)
	}
	if omd, ok := md["OfferMetadata"]; ok {
		val.OfferMetadata = null.JSONFrom(omd)
	}
	if pmd, ok := md["PricingMetadata"]; ok {
		val.PricingMetadata = null.JSONFrom(pmd)
	}
	if vmd, ok := md["VincarioMetadata"]; ok {
		val.VincarioMetadata = null.JSONFrom(vmd)
	}
	if bmd, ok := md["BlackbookMetadata"]; ok {
		val.BlackbookMetadata = null.JSONFrom(bmd)
	}
	err := val.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	return &val
}
