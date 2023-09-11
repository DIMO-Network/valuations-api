package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

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
	GetAllUserDevice(ctx context.Context, wmi string) ([]*pb.UserDevice, error)
	UpdateUserDeviceMetadata(ctx context.Context, request *pb.UpdateUserDeviceMetadataRequest) error
	GetUserDeviceOffers(ctx context.Context, userDeviceID string) (*core.DeviceOffer, error)
	GetUserDeviceValuations(ctx context.Context, userDeviceID, countryCode string) (*core.DeviceValuation, error)
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

func (das *userDeviceAPIService) UpdateUserDeviceMetadata(ctx context.Context, request *pb.UpdateUserDeviceMetadataRequest) error {
	deviceClient := pb.NewUserDeviceServiceClient(das.devicesConn)
	_, err := deviceClient.UpdateUserDeviceMetadata(ctx, request)
	return err
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
	dOffer := core.DeviceOffer{
		OfferSets: []core.OfferSet{},
	}

	// Drivly data
	drivlyVinData, err := models.Valuations(
		models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDeviceID)),
		models.ValuationWhere.OfferMetadata.IsNotNull(), // offer_metadata is sourced from drivly
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if drivlyVinData != nil {
		drivlyOffers := core.DecodeOfferFromJSON(drivlyVinData.OfferMetadata.JSON)

		requestJSON := drivlyVinData.RequestMetadata.JSON
		drivlyOffers.Updated = drivlyVinData.UpdatedAt.Format(time.RFC3339)

		requestMileage := gjson.GetBytes(requestJSON, "mileage")
		if requestMileage.Exists() {
			drivlyOffers.Mileage = int(requestMileage.Int())
		}
		requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
		if requestZipCode.Exists() {
			drivlyOffers.ZipCode = requestZipCode.String()
		}

		dOffer.OfferSets = append(dOffer.OfferSets, drivlyOffers)
	}

	return &dOffer, nil
}

func (das *userDeviceAPIService) GetUserDeviceValuations(ctx context.Context, userDeviceID, countryCode string) (*core.DeviceValuation, error) {

	dVal := core.DeviceValuation{
		ValuationSets: []core.ValuationSet{},
	}

	// Drivly data
	valuationData, err := models.Valuations(
		models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDeviceID)),
		qm.Where(fmt.Sprintf("%s is not null or %s is not null", models.ValuationColumns.DrivlyPricingMetadata, models.ValuationColumns.VincarioMetadata)),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(ctx, das.dbs().Reader)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
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
				das.logger.Warn().Msg("did not find a drivly trade-in or retail value, or json in unexpected format")
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

	return &dVal, nil
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
	if len(pricings) > 1 {
		sum := 0
		for _, v := range pricings {
			sum += v
		}
		return sum / len(pricings)
	}

	return 0
}
