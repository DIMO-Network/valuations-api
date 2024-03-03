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
	GetUserDeviceOffersByTokenID(ctx context.Context, tokenID *big.Int, take int) (*core.DeviceOffer, error)
	GetUserDeviceValuations(ctx context.Context, userDeviceID, countryCode string) (*core.DeviceValuation, error)
	GetUserDeviceValuationsByTokenID(ctx context.Context, tokenID *big.Int, countryCode string, take int) (*core.DeviceValuation, error)
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
		qm.Limit(1)).One(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return getUserDeviceOffers(drivlyVinData, nil)
}

func (das *userDeviceAPIService) GetUserDeviceOffersByTokenID(ctx context.Context, tokenID *big.Int, take int) (*core.DeviceOffer, error) {
	// Drivly data
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0))
	drivlyVinData, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(tid),
		models.ValuationWhere.OfferMetadata.IsNotNull(), // offer_metadata is sourced from drivly
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return getUserDeviceOffers(drivlyVinData, &take)
}

func getUserDeviceOffers(drivlyVinData *models.Valuation, take *int) (*core.DeviceOffer, error) {
	dOffer := core.DeviceOffer{
		OfferSets: []core.OfferSet{},
	}

	if drivlyVinData != nil {
		drivlyOffers := core.DecodeOfferFromJSON(drivlyVinData.OfferMetadata.JSON)

		requestJSON := drivlyVinData.RequestMetadata.JSON
		drivlyOffers.Updated = drivlyVinData.UpdatedAt.Format(time.RFC3339)
		if drivlyVinData.UpdatedAt.Add(time.Hour * 24 * 7).Before(time.Now()) {
			for i := range drivlyOffers.Offers {
				drivlyOffers.Offers[i].URL = ""
			}
		}

		requestMileage := gjson.GetBytes(requestJSON, "mileage")
		if requestMileage.Exists() {
			drivlyOffers.Mileage = int(requestMileage.Int())
		}
		requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
		if requestZipCode.Exists() {
			drivlyOffers.ZipCode = requestZipCode.String()
		}

		dOffer.OfferSets = append(dOffer.OfferSets, drivlyOffers)

		sort.Slice(dOffer.OfferSets, func(i, j int) bool {
			return dOffer.OfferSets[i].Updated > dOffer.OfferSets[j].Updated
		})

		if take != nil && len(dOffer.OfferSets) > *take {
			dOffer.OfferSets = dOffer.OfferSets[:*take]
		}

	}

	return &dOffer, nil
}

func (das *userDeviceAPIService) GetUserDeviceValuations(ctx context.Context, userDeviceID, countryCode string) (*core.DeviceValuation, error) {
	// Drivly data
	valuationData, err := models.Valuations(
		models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDeviceID)),
		qm.Where(fmt.Sprintf("%s is not null or %s is not null", models.ValuationColumns.DrivlyPricingMetadata, models.ValuationColumns.VincarioMetadata)),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return das.getUserDeviceValuations(valuationData, countryCode, nil)
}

func (das *userDeviceAPIService) GetUserDeviceValuationsByTokenID(ctx context.Context, tokenID *big.Int, countryCode string, take int) (*core.DeviceValuation, error) {
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0))
	// Drivly data
	valuationData, err := models.Valuations(
		models.ValuationWhere.TokenID.EQ(tid),
		qm.Where(fmt.Sprintf("%s is not null or %s is not null", models.ValuationColumns.DrivlyPricingMetadata, models.ValuationColumns.VincarioMetadata)),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return das.getUserDeviceValuations(valuationData, countryCode, &take)
}

func (das *userDeviceAPIService) getUserDeviceValuations(valuationData *models.Valuation, countryCode string, take *int) (*core.DeviceValuation, error) {
	dVal := core.DeviceValuation{
		ValuationSets: []core.ValuationSet{},
	}

	if valuationData != nil {
		if valuationData.DrivlyPricingMetadata.Valid {
			drivlyVal := core.ValuationSet{
				Vendor:        "drivly",
				TradeInSource: "drivly",
				RetailSource:  "drivly",
				Updated:       valuationData.UpdatedAt.Format(time.RFC3339),
			}
			drivlyJSON := valuationData.DrivlyPricingMetadata.JSON
			requestJSON := valuationData.RequestMetadata.JSON
			drivlyMileage := gjson.GetBytes(drivlyJSON, "mileage")
			if drivlyMileage.Exists() {
				drivlyVal.Mileage = int(drivlyMileage.Int())
				drivlyVal.Odometer = int(drivlyMileage.Int())
				drivlyVal.OdometerUnit = "miles"
			} else {
				requestMileage := gjson.GetBytes(requestJSON, "mileage")
				if requestMileage.Exists() {
					drivlyVal.Mileage = int(requestMileage.Int())
				}
			}
			requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
			if requestZipCode.Exists() {
				drivlyVal.ZipCode = requestZipCode.String()
			}
			// Drivly Trade-In
			drivlyVal.TradeIn = extractDrivlyValuation(drivlyJSON, "trade")
			drivlyVal.TradeInAverage = drivlyVal.TradeIn
			// Drivly Retail
			drivlyVal.Retail = extractDrivlyValuation(drivlyJSON, "retail")
			drivlyVal.RetailAverage = drivlyVal.Retail
			drivlyVal.Currency = "USD"

			// often drivly saves valuations with 0 for value, if this is case do not consider it
			if drivlyVal.Retail > 0 || drivlyVal.TradeIn > 0 {
				// set the price to display to users
				drivlyVal.UserDisplayPrice = (drivlyVal.Retail + drivlyVal.TradeIn) / 2
				dVal.ValuationSets = append(dVal.ValuationSets, drivlyVal)
			} else {
				das.logger.Warn().Str("vin", valuationData.Vin).
					Str("user_device_id", valuationData.UserDeviceID.String).
					Msgf("did not find a drivly trade-in or retail value, or json in unexpected format. id: %s", valuationData.ID)
			}
		} else if valuationData.VincarioMetadata.Valid {
			ratio := 1.0

			vincarioVal := core.ValuationSet{
				Vendor:        "vincario",
				TradeInSource: "vincario",
				RetailSource:  "vincario",
				Updated:       valuationData.UpdatedAt.Format(time.RFC3339),
			}

			if strings.EqualFold(countryCode, "TUR") {
				ratio = 1.5
			}

			valJSON := valuationData.VincarioMetadata.JSON
			requestJSON := valuationData.RequestMetadata.JSON
			odometerMarket := gjson.GetBytes(valJSON, "market_odometer.odometer_avg")
			if odometerMarket.Exists() {
				vincarioVal.Mileage = int(odometerMarket.Int())
				vincarioVal.Odometer = int(odometerMarket.Int())
				vincarioVal.OdometerUnit = gjson.GetBytes(valJSON, "market_odometer.odometer_unit").String()
			}
			// TODO: this needs to be implemented in the load_valuations script
			requestPostalCode := gjson.GetBytes(requestJSON, "postalCode")
			if requestPostalCode.Exists() {
				vincarioVal.ZipCode = requestPostalCode.String()
			}
			// vincario Trade-In - just using the price below mkt mean
			vincarioVal.TradeIn = int(gjson.GetBytes(valJSON, "market_price.price_below").Float() * ratio)
			vincarioVal.TradeInAverage = vincarioVal.TradeIn
			// vincario Retail - just using the price above mkt mean
			vincarioVal.Retail = int(gjson.GetBytes(valJSON, "market_price.price_above").Float() * ratio)
			vincarioVal.RetailAverage = vincarioVal.Retail

			vincarioVal.UserDisplayPrice = int(gjson.GetBytes(valJSON, "market_price.price_avg").Float() * ratio)
			vincarioVal.Currency = gjson.GetBytes(valJSON, "market_price.price_currency").String()

			// often drivly saves valuations with 0 for value, if this is case do not consider it
			if vincarioVal.Retail > 0 || vincarioVal.TradeIn > 0 {
				dVal.ValuationSets = append(dVal.ValuationSets, vincarioVal)
			} else {
				das.logger.Warn().Msg("did not find a market value from vincario, or valJSON in unexpected format")
			}
		}
	}

	sort.Slice(dVal.ValuationSets, func(i, j int) bool {
		return dVal.ValuationSets[i].Updated > dVal.ValuationSets[j].Updated
	})

	if take != nil && len(dVal.ValuationSets) > *take {
		dVal.ValuationSets = dVal.ValuationSets[:*take]
	}

	return &dVal, nil
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
