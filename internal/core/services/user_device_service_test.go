package services

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/DIMO-Network/shared/db"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/dbtest"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const migrationsDirRelPath = "../../infrastructure/db/migrations"

type UserDeviceServiceTestSuite struct {
	suite.Suite
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
	svc       UserDeviceAPIService
}

//go:embed test_drivly_offers_by_vin.json
var testDrivlyOffersJSON string

//go:embed test_drivly_pricing_by_vin.json
var testDrivlyPricingJSON string

//go:embed test_drivly_pricing2.json
var testDrivlyPricing2JSON string

//go:embed test_vincario_valuation.json
var testVincarioValuationJSON string

// SetupSuite starts container db
func (s *UserDeviceServiceTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = dbtest.StartContainerDatabase(s.ctx, "valuations_api", s.T(), migrationsDirRelPath)
	logger := dbtest.Logger()

	s.svc = NewUserDeviceService(nil, s.pdb.DBS, logger)

	var err error

	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserDeviceServiceTestSuite) SetupTest() {

}

// TearDownTest after each test truncate tables
func (s *UserDeviceServiceTestSuite) TearDownTest() {
	dbtest.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *UserDeviceServiceTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

// Test Runner
func TestUserDeviceServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserDeviceServiceTestSuite))
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceValuations_Format1() {
	// setup
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"

	_ = SetupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"DrivlyPricingMetadata": []byte(testDrivlyPricingJSON),
	}, s.pdb)

	// test
	valuations, err := s.svc.GetUserDeviceValuations(s.ctx, udID, vin)

	assert.NoError(s.T(), err)

	assert.Equal(s.T(), 1, len(valuations.ValuationSets))
	assert.Equal(s.T(), 49957, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), 49957, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), "miles", valuations.ValuationSets[0].OdometerUnit)
	assert.Equal(s.T(), 54123, valuations.ValuationSets[0].Retail)
	//54123 + 50151 / 2
	assert.Equal(s.T(), 52049, valuations.ValuationSets[0].UserDisplayPrice)
	assert.Equal(s.T(), "USD", valuations.ValuationSets[0].Currency)
	// 49040 + 52173 + 49241 / 3 = 50151
	assert.Equal(s.T(), 49976, valuations.ValuationSets[0].TradeIn)
	assert.Equal(s.T(), 49976, valuations.ValuationSets[0].TradeInAverage)
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceValuations_Format2() {
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"
	_ = SetupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"DrivlyPricingMetadata": []byte(testDrivlyPricing2JSON),
	}, s.pdb)

	valuations, err := s.svc.GetUserDeviceValuations(s.ctx, udID, vin)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(valuations.ValuationSets))
	// mileage comes from request metadata, but it is also sometimes returned by payload
	assert.Equal(s.T(), 50702, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), 40611, valuations.ValuationSets[0].TradeIn)
	assert.Equal(s.T(), 50803, valuations.ValuationSets[0].Retail)
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceValuations_Vincario() {
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"

	_ = SetupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"VincarioMetadata": []byte(testVincarioValuationJSON),
	}, s.pdb)

	valuations, err := s.svc.GetUserDeviceValuations(s.ctx, udID, vin)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(valuations.ValuationSets))
	// mileage comes from request metadata, but it is also sometimes returned by payload
	assert.Equal(s.T(), 30137, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), 30137, valuations.ValuationSets[0].Odometer)
	assert.Equal(s.T(), "km", valuations.ValuationSets[0].OdometerUnit)
	assert.Equal(s.T(), "EUR", valuations.ValuationSets[0].Currency)

	assert.Equal(s.T(), 44800, valuations.ValuationSets[0].TradeIn)
	assert.Equal(s.T(), 55200, valuations.ValuationSets[0].Retail)
	assert.Equal(s.T(), 51440, valuations.ValuationSets[0].UserDisplayPrice)
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceOffers() {
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"

	_ = SetupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"OfferMetadata": []byte(testDrivlyOffersJSON),
	}, s.pdb)

	deviceOffers, err := s.svc.GetUserDeviceOffers(s.ctx, udID)

	assert.NoError(s.T(), err)

	assert.Equal(s.T(), 1, len(deviceOffers.OfferSets))
	assert.Equal(s.T(), "drivly", deviceOffers.OfferSets[0].Source)
	assert.Equal(s.T(), 3, len(deviceOffers.OfferSets[0].Offers))

	var vroomOffer core.Offer
	var carvanaOffer core.Offer
	var carmaxOffer core.Offer

	for _, offer := range deviceOffers.OfferSets[0].Offers {
		switch offer.Vendor {
		case "vroom":
			vroomOffer = offer
		case "carvana":
			carvanaOffer = offer
		case "carmax":
			carmaxOffer = offer
		}
	}

	assert.Equal(s.T(), "Error in v1/acquisition/appraisal POST",
		vroomOffer.Error)
	assert.Equal(s.T(), 10123, carvanaOffer.Price)
	assert.Equal(s.T(), "Make[Ford],Model[Mustang Mach-E],Year[2022] is not eligible for offer.",
		carmaxOffer.DeclineReason)
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
	if vmd, ok := md["VincarioMetadata"]; ok {
		val.VincarioMetadata = null.JSONFrom(vmd)
	}
	if vmd, ok := md["DrivlyPricingMetadata"]; ok {
		val.DrivlyPricingMetadata = null.JSONFrom(vmd)
	}

	err := val.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	return &val
}
