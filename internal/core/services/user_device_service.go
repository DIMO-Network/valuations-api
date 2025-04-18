package services

import (
	"context"
	"database/sql"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/valuations-api/internal/core/gateways"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/DIMO-Network/shared/pkg/db"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/db/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc"
)

//go:generate mockgen -source user_device_service.go -destination mocks/user_device_service_mock.go
type UserDeviceAPIService interface {
	GetOffers(ctx context.Context, tokenID uint64) (*core.DeviceOffer, error)
	GetValuations(ctx context.Context, tokenID uint64, privJWT string) (*core.DeviceValuation, error)
	CanRequestInstantOffer(ctx context.Context, tokenID uint64) (bool, error)
	LastRequestDidGiveError(ctx context.Context, tokenID uint64) (bool, error)
}

type userDeviceAPIService struct {
	devicesConn  *grpc.ClientConn
	dbs          func() *db.ReaderWriter
	logger       *zerolog.Logger
	locationSvc  LocationService
	telemetryAPI gateways.TelemetryAPI
}

func NewUserDeviceService(devicesConn *grpc.ClientConn, dbs func() *db.ReaderWriter, logger *zerolog.Logger,
	locationSvc LocationService, telemetryAPI gateways.TelemetryAPI) UserDeviceAPIService {
	return &userDeviceAPIService{
		devicesConn:  devicesConn,
		dbs:          dbs,
		logger:       logger,
		locationSvc:  locationSvc,
		telemetryAPI: telemetryAPI,
	}
}

func (das *userDeviceAPIService) GetOffers(ctx context.Context, tokenID uint64) (*core.DeviceOffer, error) {
	// Drivly data
	tokenDecimal := types.NewNullDecimal(decimal.New(int64(tokenID), 0))
	drivlyVinData, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(tokenDecimal),
		models.ValuationWhere.OfferMetadata.IsNotNull(), // offer_metadata is sourced from drivly
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).All(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	offers, err := getUserDeviceOffers(drivlyVinData)
	return offers, err
}

// GetValuations retrieves device valuation details based on the provided tokenID and private JWT token header include Bearer JWT.
// It queries valuation data, retrieves the latest telemetry signals, and determines the geo-decoded location for valuation.
func (das *userDeviceAPIService) GetValuations(ctx context.Context, tokenID uint64, privJWT string) (*core.DeviceValuation, error) {
	d := decimal.New(int64(tokenID), 0)
	valuationData, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(types.NewNullDecimal(d)),
		qm.Where("drivly_pricing_metadata is not null or vincario_metadata is not null"),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).All(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	signals, err := das.telemetryAPI.GetLatestSignals(tokenID, privJWT)
	if err != nil {
		das.logger.Error().Err(err).Msgf("failed to get latest signals for token %d, skipping", tokenID)
	}
	location, err := das.locationSvc.GetGeoDecodedLocation(ctx, signals, tokenID)
	if err != nil {
		das.logger.Error().Err(err).Msgf("failed to get geo decoded location for token %d, skipping", tokenID)
	}
	countryCode := "USA"
	if location != nil {
		countryCode = location.CountryCode
	}

	return buildValuationsFromSlice(das.logger, valuationData, countryCode)
}

func getUserDeviceOffers(drivlyVinData models.ValuationSlice) (*core.DeviceOffer, error) {
	dOffer := core.DeviceOffer{
		OfferSets: []core.OfferSet{},
	}
	for _, offer := range drivlyVinData {
		// func to project the data, if not nil add to dOffer.OfferSets
		offerSet := core.DecodeOfferFromJSON(offer.OfferMetadata.JSON)

		requestJSON := offer.RequestMetadata.JSON
		offerSet.Updated = offer.UpdatedAt.Format(time.RFC3339)
		// remove offer url if it has expired (7 days)
		if offer.UpdatedAt.Add(time.Hour * 24 * 7).Before(time.Now()) {
			for i := range offerSet.Offers {
				offerSet.Offers[i].URL = ""
			}
		}

		requestMileage := gjson.GetBytes(requestJSON, "mileage")
		if requestMileage.Exists() {
			offerSet.Mileage = int(requestMileage.Int())
		}
		requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
		if requestZipCode.Exists() {
			offerSet.ZipCode = requestZipCode.String()
		}
		// set odometer measure type
		if offerSet.Mileage%12000 == 0 {
			offerSet.OdometerMeasurementType = core.Estimated
		} else {
			offerSet.OdometerMeasurementType = core.Real
		}

		dOffer.OfferSets = append(dOffer.OfferSets, offerSet)
	}

	sort.Slice(dOffer.OfferSets, func(i, j int) bool {
		return dOffer.OfferSets[i].Updated > dOffer.OfferSets[j].Updated
	})

	return &dOffer, nil
}

func buildValuationsFromSlice(logger *zerolog.Logger, valuations models.ValuationSlice, countryCode string) (*core.DeviceValuation, error) {
	dVal := core.DeviceValuation{
		ValuationSets: []core.ValuationSet{},
	}

	for _, valuation := range valuations {
		valSet := projectValuation(logger, valuation, countryCode)
		if valSet != nil {
			dVal.ValuationSets = append(dVal.ValuationSets, *valSet)
		}
	}
	sort.Slice(dVal.ValuationSets, func(i, j int) bool {
		return dVal.ValuationSets[i].Updated > dVal.ValuationSets[j].Updated
	})

	return &dVal, nil
}

func projectValuation(logger *zerolog.Logger, valuation *models.Valuation, countryCode string) *core.ValuationSet {
	if !valuation.DrivlyPricingMetadata.Valid && !valuation.VincarioMetadata.Valid {
		return nil
	}
	valSet := core.ValuationSet{
		Updated: valuation.UpdatedAt.Format(time.RFC3339),
	}
	if valuation.DrivlyPricingMetadata.Valid {
		valSet.Vendor = "drivly"
		valSet.TradeInSource = "drivly"
		valSet.RetailSource = "drivly"

		drivlyJSON := valuation.DrivlyPricingMetadata.JSON
		requestJSON := valuation.RequestMetadata.JSON
		drivlyMileage := gjson.GetBytes(drivlyJSON, "mileage")
		if drivlyMileage.Exists() {
			valSet.Mileage = int(drivlyMileage.Int())
			valSet.Odometer = int(drivlyMileage.Int())
			valSet.OdometerUnit = "miles"
		} else {
			requestMileage := gjson.GetBytes(requestJSON, "mileage")
			if requestMileage.Exists() {
				valSet.Mileage = int(requestMileage.Int())
			}
		}
		requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
		if requestZipCode.Exists() {
			valSet.ZipCode = requestZipCode.String()
		}
		// Drivly Trade-In
		valSet.TradeIn = extractDrivlyValuation(drivlyJSON, "trade")
		valSet.TradeInAverage = valSet.TradeIn
		// Drivly Retail
		valSet.Retail = extractDrivlyValuation(drivlyJSON, "retail")
		valSet.RetailAverage = valSet.Retail
		valSet.Currency = "USD"

		// set the price to display to users
		valSet.UserDisplayPrice = (valSet.Retail + valSet.TradeIn) / 2
	} else if valuation.VincarioMetadata.Valid {
		ratio := 1.0
		if strings.EqualFold(countryCode, "TUR") {
			ratio = 1.5
		}
		valSet.Vendor = "vincario"
		valSet.TradeInSource = "vincario"
		valSet.RetailSource = "vincario"

		valJSON := valuation.VincarioMetadata.JSON
		requestJSON := valuation.RequestMetadata.JSON
		// vincario now suports two markets
		odometerRegion := gjson.GetBytes(valJSON, "market_odometer.europe")
		if !odometerRegion.Exists() {
			odometerRegion = gjson.GetBytes(valJSON, "market_odometer.north_america")
		}
		odometerMarket := odometerRegion.Get("odometer_avg")

		if odometerMarket.Exists() {
			valSet.Mileage = int(odometerMarket.Int())
			valSet.Odometer = int(odometerMarket.Int())
			valSet.OdometerUnit = odometerRegion.Get("odometer_unit").String()
		}
		// TODO: this needs to be implemented in the load_valuations script
		requestPostalCode := gjson.GetBytes(requestJSON, "postalCode")
		if requestPostalCode.Exists() {
			valSet.ZipCode = requestPostalCode.String()
		}
		priceRegion := gjson.GetBytes(valJSON, "market_price.europe")
		if !priceRegion.Exists() {
			priceRegion = gjson.GetBytes(valJSON, "market_price.north_america")
		}
		// vincario Trade-In - just using the price below mkt mean
		valSet.TradeIn = int(priceRegion.Get("price_below").Float() * ratio)
		valSet.TradeInAverage = valSet.TradeIn
		// vincario Retail - just using the price above mkt mean
		valSet.Retail = int(priceRegion.Get("price_above").Float() * ratio)
		valSet.RetailAverage = valSet.Retail

		valSet.UserDisplayPrice = int(priceRegion.Get("price_avg").Float() * ratio)
		valSet.Currency = priceRegion.Get("price_currency").String()
	}
	// make sure valid data & set odo type
	if valSet.Retail > 0 || valSet.TradeIn > 0 {
		if valSet.Vendor == "vincario" {
			valSet.OdometerMeasurementType = core.Market
		} else if valSet.Odometer%12000 == 0 {
			valSet.OdometerMeasurementType = core.Estimated
		} else {
			valSet.OdometerMeasurementType = core.Real
		}
		return &valSet
	}
	logger.Debug().Str("vin", valuation.Vin).Msgf("did not find a market value from %s, or valJSON in unexpected format", valSet.Vendor)
	return nil
}

func (das *userDeviceAPIService) CanRequestInstantOffer(ctx context.Context, tokenID uint64) (bool, error) {
	tokenDecimal := types.NewNullDecimal(decimal.New(int64(tokenID), 0))
	existingOfferData, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(tokenDecimal),
		models.ValuationWhere.OfferMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return canRequestInstantOffer(existingOfferData)
}

func canRequestInstantOffer(existingOfferData *models.Valuation) (bool, error) {
	if existingOfferData != nil {
		if existingOfferData.CreatedAt.After(time.Now().Add(-time.Hour * 24 * 7)) {
			return false, nil
		}
	}

	return true, nil
}

func (das *userDeviceAPIService) LastRequestDidGiveError(ctx context.Context, tokenID uint64) (bool, error) {
	tokenDecimal := types.NewNullDecimal(decimal.New(int64(tokenID), 0))
	existingOfferData, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(tokenDecimal),
		models.ValuationWhere.OfferMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(ctx, das.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	if existingOfferData != nil {
		notReturnedError := true
		offersSet := core.DecodeOfferFromJSON(existingOfferData.OfferMetadata.JSON)

		for _, offer := range offersSet.Offers {
			isErrorEmpty := offer.Error == "" || offer.DeclineReason == ""
			notReturnedError = notReturnedError && isErrorEmpty
		}

		return notReturnedError, nil
	}

	return true, nil
}

func (das *userDeviceAPIService) LastRequestDidGiveErrorByTokenID(ctx context.Context, tokenID *big.Int) (bool, error) {
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0))
	existingOfferData, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(tid),
		models.ValuationWhere.OfferMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(ctx, das.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return lastRequestDidGiveError(existingOfferData)
}

func lastRequestDidGiveError(existingOfferData *models.Valuation) (bool, error) {
	if existingOfferData != nil {
		notReturnedError := true
		offersSet := core.DecodeOfferFromJSON(existingOfferData.OfferMetadata.JSON)

		for _, offer := range offersSet.Offers {
			isErrorEmpty := offer.Error == "" || offer.DeclineReason == ""
			notReturnedError = notReturnedError && isErrorEmpty
		}

		return notReturnedError, nil
	}

	return true, nil
}

// extractDrivlyValuation pulls out the price from the drivly json, based on the passed in key, eg. trade or retail. calculates average if no root property found
func extractDrivlyValuation(drivlyJSON []byte, key string) int {
	// handle when value is just set at top level
	if gjson.GetBytes(drivlyJSON, key).Exists() && !gjson.GetBytes(drivlyJSON, key).IsObject() {
		v := gjson.GetBytes(drivlyJSON, key).String()
		vf, _ := strconv.ParseFloat(v, 64)
		return int(vf)
	}
	// if no specific value, make an average of all values drivly offers
	pricings := map[string]int64{}
	if gjson.GetBytes(drivlyJSON, key+".blackBook.totalAvg").Exists() {
		pricings["blackbook"] = gjson.GetBytes(drivlyJSON, key+".blackBook.totalAvg").Int()
	}
	if gjson.GetBytes(drivlyJSON, key+".kelley.good").Exists() {
		pricings["kbb"] = gjson.GetBytes(drivlyJSON, key+".kelley.good").Int()
	} else if gjson.GetBytes(drivlyJSON, key+".kelley.book").Exists() {
		pricings["kbb"] = gjson.GetBytes(drivlyJSON, key+".kelley.book").Int()
	}
	if gjson.GetBytes(drivlyJSON, key+".edmunds.average").Exists() {
		pricings["edmunds"] = gjson.GetBytes(drivlyJSON, key+".edmunds.average").Int()
	}
	if gjson.GetBytes(drivlyJSON, key+".nada.book").Exists() {
		pricings["nada"] = gjson.GetBytes(drivlyJSON, key+".nada.book").Int()
	}
	if gjson.GetBytes(drivlyJSON, key+".cargurus").Exists() {
		pricings["cargurus"] = gjson.GetBytes(drivlyJSON, key+".cargurus").Int()
	}
	if len(pricings) > 0 {
		sum := int64(0)
		denominator := int64(0)
		for _, v := range pricings {
			if v > 100 {
				sum += v
				denominator++
			}
		}
		return int(sum / denominator)
	}

	return 0
}
