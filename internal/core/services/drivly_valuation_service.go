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
	internalmodel "github.com/DIMO-Network/valuations-api/internal/core/models"
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

func NewDrivlyValuationService(DBS func() *db.ReaderWriter, log *zerolog.Logger, settings *config.Settings, ddSvc DeviceDefinitionsAPIService, uddSvc UserDeviceDataAPIService) DrivlyValuationService {
	return &drivlyValuationService{
		dbs:       DBS,
		log:       log,
		drivlySvc: NewDrivlyAPIService(settings, DBS),
		ddSvc:     ddSvc,
		geoSvc:    NewGoogleGeoAPIService(settings),
		uddSvc:    uddSvc,
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
	localLog := d.log.With().Str("vin", vin).Str("deviceDefinitionID", deviceDefinitionID).Logger()

	existingVINData, err := models.Valuations(
		models.ValuationWhere.Vin.EQ(vin),
		models.ValuationWhere.VinMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)

	if err != nil {
		return ErrorDataPullStatus, err
	}

	// make sure userdevice exists
	ud, err := d.udSvc.GetUserDevice(ctx, userDeviceID)
	if err != nil {
		return ErrorDataPullStatus, err
	}

	// by this point we know we might need to insert drivly raw json data
	externalVinData := &models.Valuation{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(deviceDef.DeviceDefinitionId),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}

	// should probably move this up to top as our check for never pulled, then seperate call to get latest pull date for repullWindow check
	if existingVINData != nil && existingVINData.VinMetadata.Valid {
		var vinInfo map[string]interface{}
		err = existingVINData.VinMetadata.Unmarshal(&vinInfo)
		if err != nil {
			return ErrorDataPullStatus, errors.Wrap(err, "unable to unmarshal vin metadata")
		}
		// update the device attributes via gRPC
		err2 := d.ddSvc.UpdateDeviceDefAttrs(ctx, deviceDef, vinInfo)
		if err2 != nil {
			return ErrorDataPullStatus, err2
		}
	}

	// determine if want to pull pricing data
	existingPricingData, _ := models.Valuations(
		models.ValuationWhere.Vin.EQ(vin),
		models.ValuationWhere.PricingMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)
	// just return if already pulled recently for this VIN, but still need to insert never pulled vin - should be uncommon scenario
	if existingPricingData != nil && existingPricingData.UpdatedAt.Add(repullWindow).After(time.Now()) {
		localLog.Info().Msgf("already pulled pricing data for vin %s, skipping", vin)
		return SkippedDataPullStatus, nil
	}

	// get mileage for the drivly request
	userDeviceData, err := d.uddSvc.GetUserDeviceData(ctx, userDeviceID, ud.DeviceDefinitionId)
	deviceMileage, err := d.getDeviceMileage(userDeviceData, int(deviceDef.Type.Year))
	if err != nil {
		return ErrorDataPullStatus, err
	}

	reqData := ValuationRequestData{
		Mileage: deviceMileage,
	}

	udMD := new(internalmodel.UserDeviceMetadata)
	//_ = ud.Metadata.Unmarshal(udMD) TODO: edu

	if udMD.PostalCode == nil {
		lat := userDeviceData.Latitude
		long := userDeviceData.Longitude
		localLog.Info().Msgf("lat long found: %f, %f", lat, long)
		if lat != 0 && long != 0 {
			gl, err := d.geoSvc.GeoDecodeLatLong(lat, long)
			if err != nil {
				localLog.Err(err).Msgf("failed to GeoDecode lat long %f, %f", lat, long)
			}
			if gl != nil {
				// update UD, ignore if fails doesn't matter
				udMD.PostalCode = &gl.PostalCode
				udMD.GeoDecodedCountry = &gl.Country
				udMD.GeoDecodedStateProv = &gl.AdminAreaLevel1

				// TODO: edu
				//_ = ud.Metadata.Marshal(udMD)
				//_, err = ud.Update(ctx, d.dbs().Writer, boil.Whitelist(models.UserDeviceColumns.Metadata, models.UserDeviceColumns.UpdatedAt))
				//if err != nil {
				//	localLog.Err(err).Msg("failed to update user_device.metadata with geodecode info")
				//}
				localLog.Info().Msgf("GeoDecoded a lat long: %+v", gl)
			}
		}
	}

	if udMD.PostalCode != nil {
		reqData.ZipCode = udMD.PostalCode
	}
	_ = externalVinData.RequestMetadata.Marshal(reqData)

	// only pull offers and pricing on every pull.
	offer, err := d.drivlySvc.GetOffersByVIN(vin, &reqData)
	if err == nil {
		_ = externalVinData.OfferMetadata.Marshal(offer)
	}
	pricing, err := d.drivlySvc.GetVINPricing(vin, &reqData)
	if err == nil {
		_ = externalVinData.PricingMetadata.Marshal(pricing)
	}

	// check on edmunds data so we can get the style id
	edmundsExists, _ := models.Valuations(models.ValuationWhere.UserDeviceID.EQ(null.StringFrom(ud.Id)),
		models.ValuationWhere.EdmundsMetadata.IsNotNull()).Exists(ctx, d.dbs().Reader)
	if !edmundsExists {
		// extra optional data that only needs to be pulled once.
		edmunds, err := d.drivlySvc.GetEdmundsByVIN(vin) // this is source data that will only be available after pulling vin + pricing
		if err == nil {
			_ = externalVinData.EdmundsMetadata.Marshal(edmunds)
		}
		// fill in edmunds style_id in our user_device if it exists and not already set. None of these seen as bad errors so just logs
		if edmunds != nil && ud.DeviceStyleId == nil {
			d.setUserDeviceStyleFromEdmunds(ctx, edmunds, ud)
			localLog.Info().Msgf("set device_style_id for ud id %s", ud.Id)
		} else {
			localLog.Warn().Msgf("could not set edmunds style id. edmunds data exists: %v. ud style_id already set: %v", edmunds != nil, ud.DeviceStyleId)
		}
	}

	err = externalVinData.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return ErrorDataPullStatus, err
	}

	//defer appmetrics.DrivlyIngestTotalOps.Inc()

	return PulledValuationDrivlyStatus, nil
}

const EstMilesPerYear = 12000.0

func (d *drivlyValuationService) getDeviceMileage(userDeviceData *pb.UserDeviceDataResponse, modelYear int) (mileage *float64, err error) {
	var deviceMileage *float64
	if userDeviceData.Odometer > 0 {
		*deviceMileage = userDeviceData.Odometer
	}

	if userDeviceData.Odometer == 0 {
		deviceMileage = new(float64)
		yearDiff := time.Now().Year() - modelYear
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
