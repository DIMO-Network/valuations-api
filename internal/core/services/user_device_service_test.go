package services

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/ericlagergren/decimal"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/types"

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

//go:embed test_drivly_valuation3.json
var testDrivlyValuations3JSON string

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

// *** Valuations *** //

func Test_projectValuation(t *testing.T) {
	// change to test projectValuation
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()

	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"
	// this is the one that in prod would return a valuation that was below what made sense, like if not averaging right
	valuation := setupCreateValuationsData(t, ddID, udID, vin, map[string][]byte{
		"DrivlyPricingMetadata": []byte(testDrivlyValuations3JSON),
	}, nil)

	valuationSet := projectValuation(&logger, valuation, "USA")

	// mileage comes from request metadata, but it is also sometimes returned by payload
	assert.Equal(t, 24000, valuationSet.Mileage, "mileage must be what is in the mileage json node from drivly, ideally matches request")
	assert.Equal(t, 26718, valuationSet.TradeIn)
	assert.Equal(t, 32442, valuationSet.Retail)
	assert.Equal(t, 29580, valuationSet.UserDisplayPrice)
	assert.Equal(t, core.Estimated, valuationSet.OdometerMeasurementType)
}
func Test_projectValuation_empty(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()

	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"

	v := models.Valuation{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(ddID),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(udID),
		OfferMetadata:      null.JSONFrom([]byte(`{}`)),
	}
	val := projectValuation(&logger, &v, "USA")
	assert.Nil(t, val, "if no valuations should return nil")
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceValuations_Format1() {
	// setup
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"

	_ = setupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"DrivlyPricingMetadata": []byte(testDrivlyPricingJSON),
	}, &s.pdb)

	// test
	valuations, err := s.svc.GetUserDeviceValuations(s.ctx, udID, vin)

	assert.NoError(s.T(), err)

	assert.Equal(s.T(), 1, len(valuations.ValuationSets))
	assert.Equal(s.T(), 49957, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), 49957, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), "miles", valuations.ValuationSets[0].OdometerUnit)
	assert.Equal(s.T(), 54123, valuations.ValuationSets[0].Retail)
	//54123 + 50151 / 2
	assert.Equal(s.T(), 52259, valuations.ValuationSets[0].UserDisplayPrice)
	assert.Equal(s.T(), "USD", valuations.ValuationSets[0].Currency)
	// 49040 + 52173 + 49241 / 3 = 50151
	assert.Equal(s.T(), 50396, valuations.ValuationSets[0].TradeIn)
	assert.Equal(s.T(), 50396, valuations.ValuationSets[0].TradeInAverage)
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceValuationsByTokenID_setsTokenIDFromUDID() {
	// setup
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"
	tID := big.NewInt(123)

	_ = setupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"DrivlyPricingMetadata": []byte(testDrivlyPricingJSON),
	}, &s.pdb)

	// tokenId not being set
	valuations, err := s.svc.GetUserDeviceValuationsByTokenID(s.ctx, tID, "USA", 10, udID)

	assert.NoError(s.T(), err)

	assert.Equal(s.T(), 1, len(valuations.ValuationSets))
	assert.Equal(s.T(), 49957, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), 49957, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), "miles", valuations.ValuationSets[0].OdometerUnit)
	assert.Equal(s.T(), 54123, valuations.ValuationSets[0].Retail)
	//54123 + 50151 / 2
	assert.Equal(s.T(), 52259, valuations.ValuationSets[0].UserDisplayPrice)
	assert.Equal(s.T(), "USD", valuations.ValuationSets[0].Currency)
	// 49040 + 52173 + 49241 / 3 = 50151
	assert.Equal(s.T(), 50396, valuations.ValuationSets[0].TradeIn)
	assert.Equal(s.T(), 50396, valuations.ValuationSets[0].TradeInAverage)

	// lookup in db by tokenId and should exist
	tokenID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tID, 0))
	valuation, err := models.Valuations(models.ValuationWhere.TokenID.EQ(tokenID)).One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), valuation)
	assert.Equal(s.T(), valuation.UserDeviceID.String, udID)
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceValuations_Format2() {
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"
	_ = setupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"DrivlyPricingMetadata": []byte(testDrivlyPricing2JSON),
	}, &s.pdb)

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

	_ = setupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"VincarioMetadata": []byte(testVincarioValuationJSON),
	}, &s.pdb)

	valuations, err := s.svc.GetUserDeviceValuations(s.ctx, udID, vin)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(valuations.ValuationSets))
	// mileage comes from request metadata, but it is also sometimes returned by payload
	assert.Equal(s.T(), 74926, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), 74926, valuations.ValuationSets[0].Odometer)
	assert.Equal(s.T(), "km", valuations.ValuationSets[0].OdometerUnit)
	assert.Equal(s.T(), "EUR", valuations.ValuationSets[0].Currency)

	assert.Equal(s.T(), 27900, valuations.ValuationSets[0].TradeIn)
	assert.Equal(s.T(), 35000, valuations.ValuationSets[0].Retail)
	assert.Equal(s.T(), 32115, valuations.ValuationSets[0].UserDisplayPrice)
}

// *** Instant Offers (USA only) *** //

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceOffers() {
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"

	_ = setupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"OfferMetadata": []byte(testDrivlyOffersJSON),
	}, &s.pdb)

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

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceOffersByTokenID_setsTokenIdFromUDID() {
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	udID := ksuid.New().String()
	vin := "vinny"
	tID := big.NewInt(123)

	// tokenId not being set
	_ = setupCreateValuationsData(s.T(), ddID, udID, vin, map[string][]byte{
		"OfferMetadata": []byte(testDrivlyOffersJSON),
	}, &s.pdb)

	deviceOffers, err := s.svc.GetUserDeviceOffersByTokenID(s.ctx, tID, 10, udID)
	require.NoError(s.T(), err)

	require.Equal(s.T(), 1, len(deviceOffers.OfferSets))
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
	// lookup in db by tokenId and should exist
	tokenID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tID, 0))
	offer, err := models.Valuations(models.ValuationWhere.TokenID.EQ(tokenID)).One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), offer)
	assert.Equal(s.T(), offer.UserDeviceID.String, udID)
}

// setupCreateValuationsData creates valuation requests with some standards. request mileage: 49957, zip: 48216. if pdb nil just returns
func setupCreateValuationsData(t *testing.T, ddID, userDeviceID, vin string, md map[string][]byte, pdb *db.Store) *models.Valuation {
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
	if pdb != nil {
		err := val.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	return &val
}
