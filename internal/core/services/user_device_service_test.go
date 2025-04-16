package services

import (
	"context"
	_ "embed"
	"fmt"
	mock_gateways "github.com/DIMO-Network/valuations-api/internal/core/gateways/mocks"
	mock_services "github.com/DIMO-Network/valuations-api/internal/core/services/mocks"
	"go.uber.org/mock/gomock"
	"os"
	"testing"

	"github.com/ericlagergren/decimal"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/DIMO-Network/shared/pkg/db"
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
	pdb         db.Store
	container   testcontainers.Container
	ctx         context.Context
	svc         UserDeviceAPIService
	locationSvc *mock_services.MockLocationService
	telemetry   *mock_gateways.MockTelemetryAPI
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
	mockCtrl := gomock.NewController(s.T())
	s.locationSvc = mock_services.NewMockLocationService(mockCtrl)
	s.telemetry = mock_gateways.NewMockTelemetryAPI(mockCtrl)

	s.svc = NewUserDeviceService(nil, s.pdb.DBS, logger, s.locationSvc, s.telemetry)
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
	tokenID := uint64(12334)
	vin := "vinny"
	// this is the one that in prod would return a valuation that was below what made sense, like if not averaging right
	valuation := setupCreateValuationsData(t, tokenID, ddID, vin, map[string][]byte{
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
	vin := "vinny"
	tokenID := uint64(12334)

	v := models.Valuation{
		ID:            ksuid.New().String(),
		DefinitionID:  null.StringFrom(ddID),
		Vin:           vin,
		TokenID:       types.NewNullDecimal(new(decimal.Big).SetUint64(tokenID)),
		OfferMetadata: null.JSONFrom([]byte(`{}`)),
	}
	val := projectValuation(&logger, &v, "USA")
	assert.Nil(t, val, "if no valuations should return nil")
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceValuations_Format1() {
	// setup
	ddID := ksuid.New().String()
	vin := "vinny"
	tokenID := uint64(12334)

	_ = setupCreateValuationsData(s.T(), tokenID, ddID, vin, map[string][]byte{
		"DrivlyPricingMetadata": []byte(testDrivlyPricingJSON),
	}, &s.pdb)
	signals := core.SignalsLatest{
		PowertrainTransmissionTravelledDistance: core.TimeFloatValue{Value: 49040},
		CurrentLocationLatitude:                 core.TimeFloatValue{Value: 49.241},
		CurrentLocationLongitude:                core.TimeFloatValue{Value: -123.521},
	}
	s.telemetry.EXPECT().GetLatestSignals(gomock.Any(), tokenID, "caca").Return(&signals, nil)
	s.locationSvc.EXPECT().GetGeoDecodedLocation(gomock.Any(), &signals, tokenID).Return(&core.LocationResponse{
		CountryCode: "USA",
	}, nil)

	// test
	valuations, err := s.svc.GetValuations(s.ctx, tokenID, "caca")

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

func (s *UserDeviceServiceTestSuite) TestGetValuations_setsTokenIDFromUDID() {
	// setup
	ddID := ksuid.New().String()
	vin := "vinny"
	tokenID := uint64(12334)

	_ = setupCreateValuationsData(s.T(), tokenID, ddID, vin, map[string][]byte{
		"DrivlyPricingMetadata": []byte(testDrivlyPricingJSON),
	}, &s.pdb)
	signals := core.SignalsLatest{
		PowertrainTransmissionTravelledDistance: core.TimeFloatValue{Value: 49040},
		CurrentLocationLatitude:                 core.TimeFloatValue{Value: 49.241},
		CurrentLocationLongitude:                core.TimeFloatValue{Value: -123.521},
	}
	s.telemetry.EXPECT().GetLatestSignals(gomock.Any(), tokenID, "caca").Return(&signals, nil)
	s.locationSvc.EXPECT().GetGeoDecodedLocation(gomock.Any(), &signals, tokenID).Return(&core.LocationResponse{
		CountryCode: "USA",
	}, nil)

	// tokenId not being set
	valuations, err := s.svc.GetValuations(s.ctx, tokenID, "caca")

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
	tid := types.NewNullDecimal(new(decimal.Big).SetUint64(tokenID))
	valuation, err := models.Valuations(models.ValuationWhere.TokenID.EQ(tid)).One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), valuation)
	u, _ := valuation.TokenID.Uint64()
	assert.Equal(s.T(), u, tokenID)
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceValuations_Format2() {
	ddID := ksuid.New().String()
	tokenID := uint64(12334)
	vin := "vinny"
	_ = setupCreateValuationsData(s.T(), tokenID, ddID, vin, map[string][]byte{
		"DrivlyPricingMetadata": []byte(testDrivlyPricing2JSON),
	}, &s.pdb)
	s.telemetry.EXPECT().GetLatestSignals(gomock.Any(), tokenID, "caca").Return(nil, nil)
	s.locationSvc.EXPECT().GetGeoDecodedLocation(gomock.Any(), nil, tokenID).Return(&core.LocationResponse{
		CountryCode: "USA",
	}, nil)

	valuations, err := s.svc.GetValuations(s.ctx, tokenID, "caca")

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(valuations.ValuationSets))
	// mileage comes from request metadata, but it is also sometimes returned by payload
	assert.Equal(s.T(), 50702, valuations.ValuationSets[0].Mileage)
	assert.Equal(s.T(), 40611, valuations.ValuationSets[0].TradeIn)
	assert.Equal(s.T(), 50803, valuations.ValuationSets[0].Retail)
}

func (s *UserDeviceServiceTestSuite) TestGetUserDeviceValuations_Vincario() {
	ddID := ksuid.New().String()
	tokenID := uint64(12334)
	vin := "vinny"

	_ = setupCreateValuationsData(s.T(), tokenID, ddID, vin, map[string][]byte{
		"VincarioMetadata": []byte(testVincarioValuationJSON),
	}, &s.pdb)
	signals := core.SignalsLatest{
		PowertrainTransmissionTravelledDistance: core.TimeFloatValue{Value: 49040},
		CurrentLocationLatitude:                 core.TimeFloatValue{Value: 49.241},
		CurrentLocationLongitude:                core.TimeFloatValue{Value: -123.521},
	}
	s.telemetry.EXPECT().GetLatestSignals(gomock.Any(), tokenID, "caca").Return(&signals, nil)
	s.locationSvc.EXPECT().GetGeoDecodedLocation(gomock.Any(), &signals, tokenID).Return(&core.LocationResponse{
		CountryCode: "USA",
	}, nil)

	valuations, err := s.svc.GetValuations(s.ctx, tokenID, "caca")

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
	vin := "vinny"
	tokenID := uint64(12334)

	_ = setupCreateValuationsData(s.T(), tokenID, ddID, vin, map[string][]byte{
		"OfferMetadata": []byte(testDrivlyOffersJSON),
	}, &s.pdb)
	all, err2 := models.Valuations().All(s.ctx, s.pdb.DBS().Reader)
	if err2 != nil {
		s.T().Fatal(err2)
	}
	for _, v := range all {
		u, _ := v.TokenID.Uint64()

		fmt.Printf("tokenId: %d\n", u)
	}

	deviceOffers, err := s.svc.GetOffers(s.ctx, tokenID)

	assert.NoError(s.T(), err)

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
}

// setupCreateValuationsData creates valuation requests with some standards. request mileage: 49957, zip: 48216. if pdb nil just returns
func setupCreateValuationsData(t *testing.T, tokenID uint64, ddID, vin string, md map[string][]byte, pdb *db.Store) *models.Valuation {
	val := models.Valuation{
		ID:              ksuid.New().String(),
		DefinitionID:    null.StringFrom(ddID),
		TokenID:         types.NewNullDecimal(new(decimal.Big).SetUint64(tokenID)),
		Vin:             vin,
		RequestMetadata: null.JSONFrom([]byte(`{"mileage":49957,"zipCode":"48216"}`)), // default request metadata
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
