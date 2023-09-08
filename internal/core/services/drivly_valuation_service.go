package services

import (
	"context"
	"encoding/json"

	"github.com/tidwall/gjson"

	"time"

	pb "github.com/DIMO-Network/device-data-api/pkg/grpc"
	pbdeviceapi "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/db/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

//go:generate mockgen -source drivly_valuation_service.go -destination mocks/drivly_valuation_service_mock.go

type DrivlyValuationService interface {
	PullValuation(ctx context.Context, userDeiceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error)
}

type drivlyValuationService struct {
	dbs       func() *db.ReaderWriter
	ddSvc     DeviceDefinitionsAPIService
	drivlySvc DrivlyAPIService
	udSvc     UserDeviceAPIService
	geoSvc    GoogleGeoAPIService
	uddSvc    UserDeviceDataAPIService
	log       *zerolog.Logger
}

func NewDrivlyValuationService(DBS func() *db.ReaderWriter, log *zerolog.Logger, settings *config.Settings, ddSvc DeviceDefinitionsAPIService, uddSvc UserDeviceDataAPIService, udSvc UserDeviceAPIService) DrivlyValuationService {
	return &drivlyValuationService{
		dbs:       DBS,
		log:       log,
		drivlySvc: NewDrivlyAPIService(settings, DBS),
		ddSvc:     ddSvc,
		geoSvc:    NewGoogleGeoAPIService(settings),
		uddSvc:    uddSvc,
		udSvc:     udSvc,
	}
}

func (d *drivlyValuationService) PullValuation(ctx context.Context, userDeviceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error) {
	const repullWindow = time.Hour * 24 * 14
	if len(vin) != 17 {
		return ErrorDataPullStatus, errors.Errorf("invalid VIN %s", vin)
	}

	deviceDef, err := d.ddSvc.GetDeviceDefinitionByID(ctx, deviceDefinitionID)
	if err != nil {
		return ErrorDataPullStatus, err
	}
	localLog := d.log.With().Str("vin", vin).Str("device_definition_id", deviceDefinitionID).Str("user_device_id", userDeviceID).Logger()

	// make sure userdevice exists
	userDevice, err := d.udSvc.GetUserDevice(ctx, userDeviceID)
	if err != nil {
		return ErrorDataPullStatus, err
	}

	// determine if want to pull pricing data
	existingPricingData, _ := models.Valuations(
		models.ValuationWhere.Vin.EQ(vin),
		models.ValuationWhere.DrivlyPricingMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)
	// just return if already pulled recently for this VIN, but still need to insert never pulled vin - should be uncommon scenario
	if existingPricingData != nil && existingPricingData.UpdatedAt.Add(repullWindow).After(time.Now()) {
		localLog.Info().Msgf("already pulled pricing data for vin %s, skipping", vin)
		return SkippedDataPullStatus, nil
	}

	// by this point we know we might need to insert drivly valuation
	valuation := &models.Valuation{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(deviceDef.DeviceDefinitionId),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}

	// get mileage for the drivly request
	userDeviceData, err := d.uddSvc.GetUserDeviceData(ctx, userDeviceID, userDevice.DeviceDefinitionId)
	if err != nil {
		// just warn if can't get data
		localLog.Warn().Err(err).Msgf("could not find any user device data to obtain mileage or location - continuing without")
	}
	deviceMileage, err := getDeviceMileage(userDeviceData, int(deviceDef.Type.Year), time.Now().Year())
	if err != nil {
		return ErrorDataPullStatus, err
	}

	reqData := ValuationRequestData{
		Mileage: deviceMileage,
	}

	if userDevice.PostalCode == "" && userDeviceData != nil {
		// need to geodecode the postal code
		lat := userDeviceData.Latitude
		long := userDeviceData.Longitude
		localLog.Info().Msgf("lat long found: %f, %f", safePtrFloat(lat), safePtrFloat(long))
		if lat != nil && long != nil {
			gl, err := d.geoSvc.GeoDecodeLatLong(safePtrFloat(lat), safePtrFloat(long))
			if err != nil {
				localLog.Err(err).Msgf("failed to GeoDecode lat long %f, %f", safePtrFloat(lat), safePtrFloat(long))
			}
			if gl != nil {
				userDevice.PostalCode = gl.PostalCode
				// update UD, ignore if fails doesn't matter
				err := d.udSvc.UpdateUserDeviceMetadata(ctx, &pbdeviceapi.UpdateUserDeviceMetadataRequest{
					UserDeviceId:        userDeviceID,
					PostalCode:          &gl.PostalCode,
					GeoDecodedCountry:   &gl.Country,
					GeoDecodedStateProv: &gl.AdminAreaLevel1,
				})
				if err != nil {
					localLog.Err(err).Msgf("failed to update user device metadata for postal code")
				}

				localLog.Info().Msgf("GeoDecoded a lat long: %+v", gl)
			}
		}
	}

	if userDevice.PostalCode != "" {
		reqData.ZipCode = &userDevice.PostalCode
	}
	_ = valuation.RequestMetadata.Marshal(reqData)

	pricing, err := d.drivlySvc.GetVINPricing(vin, &reqData)
	if err == nil {
		_ = valuation.DrivlyPricingMetadata.Marshal(pricing)
	}

	// check on edmunds data so we can get the style id
	edmundsExists, _ := models.Valuations(models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(userDevice.Id)),
		models.ValuationWhere.EdmundsMetadata.IsNotNull()).Exists(ctx, d.dbs().Reader)
	if !edmundsExists {
		// extra optional data that only needs to be pulled once.
		edmunds, err := d.drivlySvc.GetEdmundsByVIN(vin) // this is source data that will only be available after pulling vin + pricing
		if err == nil {
			_ = valuation.EdmundsMetadata.Marshal(edmunds)
		}
		// fill in edmunds style_id in our user_device if it exists and not already set. None of these seen as bad errors so just logs
		if edmunds != nil && userDevice.DeviceStyleId == nil {
			d.setUserDeviceStyleFromEdmunds(ctx, edmunds, userDevice)
			localLog.Info().Msgf("set device_style_id for userDevice id %s", userDevice.Id)
		} else {
			localLog.Warn().Msgf("could not set edmunds style id. edmunds data exists: %v. userDevice style_id already set: %v", edmunds != nil, userDevice.DeviceStyleId)
		}
	}

	err = valuation.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return ErrorDataPullStatus, err
	}

	//defer appmetrics.DrivlyIngestTotalOps.Inc()

	return PulledValuationDrivlyStatus, nil
}

func safePtrFloat(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

const EstMilesPerYear = 12000.0

func getDeviceMileage(userDeviceData *pb.UserDeviceDataResponse, modelYear int, currentYear int) (mileage *float64, err error) {
	var deviceMileage *float64

	if userDeviceData != nil && userDeviceData.Odometer != nil && *userDeviceData.Odometer > 0 {
		deviceMileage = userDeviceData.Odometer
	}

	if userDeviceData == nil || userDeviceData.Odometer == nil {
		deviceMileage = new(float64)
		yearDiff := currentYear - modelYear
		switch {
		case yearDiff > 0:
			// Past model year
			*deviceMileage = float64(yearDiff) * EstMilesPerYear
		case yearDiff == 0:
			// Current model year
			*deviceMileage = EstMilesPerYear / 2
		default:
			// Next model year
			*deviceMileage = 0
		}
	}

	return deviceMileage, nil
}

func (d *drivlyValuationService) setUserDeviceStyleFromEdmunds(ctx context.Context, edmunds map[string]interface{}, ud *pbdeviceapi.UserDevice) {
	edmundsJSON, err := json.Marshal(edmunds)
	if err != nil {
		d.log.Err(err).Msg("could not marshal edmunds response to json")
		return
	}
	styleIDResult := gjson.GetBytes(edmundsJSON, "edmundsStyle.data.style.id")
	styleID := styleIDResult.String()
	if styleIDResult.Exists() && len(styleID) > 0 {

		deviceStyle, err := d.ddSvc.GetDeviceStyleByExternalID(ctx, styleID)

		if err != nil {
			d.log.Err(err).Msgf("unable to find device_style for edmunds style_id %s", styleID)
			return
		}
		//TODO: edu
		ud.DeviceStyleId = &deviceStyle.Id // set foreign key
		//_, err = ud.Update(ctx, d.dbs().Writer, boil.Whitelist("updated_at", "device_style_id"))
		//if err != nil {
		//	d.log.Err(err).Msgf("unable to update user_device_id %s with styleID %s", ud.Id, deviceStyle.Id)
		//	return
		//}
	}
}
