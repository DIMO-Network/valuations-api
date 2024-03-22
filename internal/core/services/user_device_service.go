package services

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/db/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc"
)

//go:generate mockgen -source user_device_service.go -destination mocks/user_device_service_mock.go
type UserDeviceAPIService interface {
	GetUserDevice(ctx context.Context, userDeviceID string) (*pb.UserDevice, error)
	GetUserDeviceByTokenID(ctx context.Context, tokenID *big.Int) (*pb.UserDevice, error)
	GetUserDeviceByEthAddr(ctx context.Context, ethAddr string) (*pb.UserDevice, error)
	GetAllUserDevice(ctx context.Context, wmi string) ([]*pb.UserDevice, error)
	UpdateUserDeviceMetadata(ctx context.Context, request *pb.UpdateUserDeviceMetadataRequest) error
	GetUserDeviceOffers(ctx context.Context, userDeviceID string) (*core.DeviceOffer, error)
	GetUserDeviceOffersByTokenID(ctx context.Context, tokenID *big.Int, take int, userDeviceID string) (*core.DeviceOffer, error)
	GetUserDeviceValuations(ctx context.Context, userDeviceID, countryCode string) (*core.DeviceValuation, error)
	GetUserDeviceValuationsByTokenID(ctx context.Context, tokenID *big.Int, countryCode string, take int, userDeviceID string) (*core.DeviceValuation, error)
	CanRequestInstantOffer(ctx context.Context, userDeviceID string) (bool, error)
	CanRequestInstantOfferByTokenID(ctx context.Context, tokenID *big.Int) (bool, error)
	LastRequestDidGiveError(ctx context.Context, userDeviceID string) (bool, error)
	LastRequestDidGiveErrorByTokenID(ctx context.Context, tokenID *big.Int) (bool, error)
}

type userDeviceAPIService struct {
	devicesConn *grpc.ClientConn
	dbs         func() *db.ReaderWriter
	logger      *zerolog.Logger
}

func NewUserDeviceService(
	devicesConn *grpc.ClientConn,
	dbs func() *db.ReaderWriter,
	logger *zerolog.Logger,
) UserDeviceAPIService {
	return &userDeviceAPIService{
		devicesConn: devicesConn,
		dbs:         dbs,
		logger:      logger,
	}
}

// GetUserDevice gets the userDevice from devices-api, checks in local cache first
func (das *userDeviceAPIService) GetUserDevice(ctx context.Context, userDeviceID string) (*pb.UserDevice, error) {
	if len(userDeviceID) == 0 {
		return nil, fmt.Errorf("user device id was empty - invalid")
	}
	var err error
	deviceClient := pb.NewUserDeviceServiceClient(das.devicesConn)

	var userDevice *pb.UserDevice
	userDevice, err = deviceClient.GetUserDevice(ctx, &pb.GetUserDeviceRequest{
		Id: userDeviceID,
	})

	if err != nil {
		return nil, err
	}

	return userDevice, nil
}

// GetUserDeviceByTokenID gets the userDevice from devices-api, checks in local cache first
func (das *userDeviceAPIService) GetUserDeviceByTokenID(ctx context.Context, tokenID *big.Int) (*pb.UserDevice, error) {
	var err error
	deviceClient := pb.NewUserDeviceServiceClient(das.devicesConn)

	var userDevice *pb.UserDevice
	userDevice, err = deviceClient.GetUserDeviceByTokenId(ctx, &pb.GetUserDeviceByTokenIdRequest{
		TokenId: tokenID.Int64(),
	})

	if err != nil {
		return nil, err
	}

	return userDevice, nil
}

func (das *userDeviceAPIService) UpdateUserDeviceMetadata(ctx context.Context, request *pb.UpdateUserDeviceMetadataRequest) error {
	deviceClient := pb.NewUserDeviceServiceClient(das.devicesConn)
	_, err := deviceClient.UpdateUserDeviceMetadata(ctx, request)
	return err
}

func (das *userDeviceAPIService) GetUserDeviceByEthAddr(ctx context.Context, ethAddr string) (*pb.UserDevice, error) {
	deviceClient := pb.NewUserDeviceServiceClient(das.devicesConn)

	if len(ethAddr) > 2 && ethAddr[:2] == "0x" {
		ethAddr = ethAddr[2:]
	}

	ethAddrBytes, err := hex.DecodeString(ethAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid ethereum address: %w", err)
	}

	userDevice, err := deviceClient.GetUserDeviceByEthAddr(ctx, &pb.GetUserDeviceByEthAddrRequest{EthAddr: ethAddrBytes})
	if err != nil {
		return nil, err
	}

	return userDevice, nil
}

// GetAllUserDevice gets all userDevices from devices-api
func (das *userDeviceAPIService) GetAllUserDevice(ctx context.Context, wmi string) ([]*pb.UserDevice, error) {
	deviceClient := pb.NewUserDeviceServiceClient(das.devicesConn)
	all, err := deviceClient.GetAllUserDevice(ctx, &pb.GetAllUserDeviceRequest{Wmi: wmi})
	if err != nil {
		return nil, err
	}

	var useDevices []*pb.UserDevice
	for {
		response, err := all.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while receiving response: %v", err)
		}

		useDevices = append(useDevices, response)
	}

	return useDevices, nil
}

func (das *userDeviceAPIService) GetUserDeviceOffers(ctx context.Context, userDeviceID string) (*core.DeviceOffer, error) {
	// Drivly data
	drivlyVinData, err := models.Valuations(
		models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDeviceID)),
		models.ValuationWhere.OfferMetadata.IsNotNull(), // offer_metadata is sourced from drivly
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).All(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	offers, err := getUserDeviceOffers(drivlyVinData)
	return offers, err
}

func (das *userDeviceAPIService) GetUserDeviceOffersByTokenID(ctx context.Context, tokenID *big.Int, take int, userDeviceID string) (*core.DeviceOffer, error) {
	// Drivly data
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0))
	drivlyVinData, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(tid),
		models.ValuationWhere.OfferMetadata.IsNotNull(), // offer_metadata is sourced from drivly
		qm.OrderBy("updated_at desc"),
		qm.Limit(take)).All(ctx, das.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// fallback if nothing found to lookup by userDeviceID
			drivlyVinData, err = models.Valuations(
				models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDeviceID)),
				models.ValuationWhere.OfferMetadata.IsNotNull(),
				qm.OrderBy("updated_at desc"),
				qm.Limit(take)).All(ctx, das.dbs().Reader)
			if err != nil {
				return nil, err
			} else if len(drivlyVinData) > 0 {
				// update the found records and set the token id
				for _, datum := range drivlyVinData {
					datum.TokenID = tid
					_, err := datum.Update(ctx, das.dbs().Writer, boil.Infer())
					if err != nil {
						das.logger.Err(err).Str("vin", datum.Vin).Msgf("failed to set token_id on valuation id: %s", datum.ID)
					}
				}
			}
		} else {
			return nil, err
		}
	}

	return getUserDeviceOffers(drivlyVinData)
}

func (das *userDeviceAPIService) GetUserDeviceValuations(ctx context.Context, userDeviceID, countryCode string) (*core.DeviceValuation, error) {
	valuationData, err := models.Valuations(
		models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDeviceID)),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).All(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return das.getUserDeviceValuations(valuationData, countryCode)
}

func (das *userDeviceAPIService) GetUserDeviceValuationsByTokenID(ctx context.Context, tokenID *big.Int, countryCode string, take int, userDeviceID string) (*core.DeviceValuation, error) {
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0))
	valuations, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(tid),
		qm.OrderBy("updated_at desc"),
		qm.Limit(take)).All(ctx, das.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// fallback if nothing found to lookup by userDeviceID
			valuations, err = models.Valuations(
				models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDeviceID)),
				qm.OrderBy("updated_at desc"),
				qm.Limit(take)).All(ctx, das.dbs().Reader)
			if err != nil {
				return nil, err
			} else if len(valuations) > 0 {
				// update the found records and set the token id
				for _, datum := range valuations {
					datum.TokenID = tid
					_, err := datum.Update(ctx, das.dbs().Writer, boil.Infer())
					if err != nil {
						das.logger.Err(err).Str("vin", datum.Vin).Msgf("failed to set token_id on valuation id: %s", datum.ID)
					}
				}
			}
		} else {
			return nil, err
		}
	}

	return das.getUserDeviceValuations(valuations, countryCode)
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

func (das *userDeviceAPIService) getUserDeviceValuations(valuations models.ValuationSlice, countryCode string) (*core.DeviceValuation, error) {
	dVal := core.DeviceValuation{
		ValuationSets: []core.ValuationSet{},
	}

	for _, valuation := range valuations {
		valSet := das.projectValuation(valuation, countryCode)
		if valSet != nil {
			dVal.ValuationSets = append(dVal.ValuationSets, *valSet)
		}
	}
	sort.Slice(dVal.ValuationSets, func(i, j int) bool {
		return dVal.ValuationSets[i].Updated > dVal.ValuationSets[j].Updated
	})

	return &dVal, nil
}

func (das *userDeviceAPIService) projectValuation(valuation *models.Valuation, countryCode string) *core.ValuationSet {
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
		valSet.Updated = valuation.UpdatedAt.Format(time.RFC3339)

		valJSON := valuation.VincarioMetadata.JSON
		requestJSON := valuation.RequestMetadata.JSON
		odometerMarket := gjson.GetBytes(valJSON, "market_odometer.odometer_avg")
		if odometerMarket.Exists() {
			valSet.Mileage = int(odometerMarket.Int())
			valSet.Odometer = int(odometerMarket.Int())
			valSet.OdometerUnit = gjson.GetBytes(valJSON, "market_odometer.odometer_unit").String()
		}
		// TODO: this needs to be implemented in the load_valuations script
		requestPostalCode := gjson.GetBytes(requestJSON, "postalCode")
		if requestPostalCode.Exists() {
			valSet.ZipCode = requestPostalCode.String()
		}
		// vincario Trade-In - just using the price below mkt mean
		valSet.TradeIn = int(gjson.GetBytes(valJSON, "market_price.price_below").Float() * ratio)
		valSet.TradeInAverage = valSet.TradeIn
		// vincario Retail - just using the price above mkt mean
		valSet.Retail = int(gjson.GetBytes(valJSON, "market_price.price_above").Float() * ratio)
		valSet.RetailAverage = valSet.Retail

		valSet.UserDisplayPrice = int(gjson.GetBytes(valJSON, "market_price.price_avg").Float() * ratio)
		valSet.Currency = gjson.GetBytes(valJSON, "market_price.price_currency").String()
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
	das.logger.Warn().Str("vin", valuation.Vin).Msgf("did not find a market value from %s, or valJSON in unexpected format", valSet.Vendor)
	return nil
}

func (das *userDeviceAPIService) CanRequestInstantOffer(ctx context.Context, userDeviceID string) (bool, error) {

	existingOfferData, err := models.Valuations(
		models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDeviceID)),
		models.ValuationWhere.OfferMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return canRequestInstantOffer(existingOfferData)
}

func (das *userDeviceAPIService) CanRequestInstantOfferByTokenID(ctx context.Context, tokenID *big.Int) (bool, error) {
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0))
	existingOfferData, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(tid),
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

func (das *userDeviceAPIService) LastRequestDidGiveError(ctx context.Context, userDeviceID string) (bool, error) {

	existingOfferData, err := models.Valuations(
		models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDeviceID)),
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
	if gjson.GetBytes(drivlyJSON, key).Exists() && !gjson.GetBytes(drivlyJSON, key).IsObject() {
		v := gjson.GetBytes(drivlyJSON, key).String()
		vf, _ := strconv.ParseFloat(v, 64)
		return int(vf)
	}
	// get all values
	pricings := map[string]int{}
	if gjson.GetBytes(drivlyJSON, key+".blackBook.totalAvg").Exists() {
		values := gjson.GetManyBytes(drivlyJSON, key+".blackBook.totalRough", key+".blackBook.totalAvg", key+".blackBook.totalClean")
		pricings["blackbook"] = int(values[1].Int())
	}
	if gjson.GetBytes(drivlyJSON, key+".kelley.good").Exists() {
		pricings["kbb"] = int(gjson.GetBytes(drivlyJSON, key+".kelley.good").Int())
	}
	if gjson.GetBytes(drivlyJSON, key+".edmunds.average").Exists() {
		values := gjson.GetManyBytes(drivlyJSON, key+".edmunds.rough", key+".edmunds.average", key+".edmunds.clean")
		pricings["edmunds"] = int(values[1].Int())
	}
	if gjson.GetBytes(drivlyJSON, key+".nada.book").Exists() {
		values := gjson.GetManyBytes(drivlyJSON, key+".nada.base", key+".nada.avgBook", key+".nada.book")
		pricings["nada"] = int(values[1].Int())
	}
	if gjson.GetBytes(drivlyJSON, key+".cargurus").Exists() {
		values := gjson.GetManyBytes(drivlyJSON, key+".cargurus")
		pricings["cargurus"] = int(values[0].Int())
	}
	if len(pricings) > 1 {
		sum := 0
		for _, v := range pricings {
			sum += v
		}
		return sum / len(pricings)
	}

	return 0
}
