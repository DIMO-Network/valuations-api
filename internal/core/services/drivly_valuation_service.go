package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/tidwall/gjson"

	"time"

	pb "github.com/DIMO-Network/device-data-api/pkg/grpc"
	pbdeviceapi "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/db/models"

	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

//go:generate mockgen -source drivly_valuation_service.go -destination mocks/drivly_valuation_service_mock.go

type DrivlyValuationService interface {
	PullValuation(ctx context.Context, userDeiceID, deviceDefinitionID, vin string) (core.DataPullStatusEnum, error)
	PullOffer(ctx context.Context, userDeviceID string) (core.DataPullStatusEnum, error)
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

func (d *drivlyValuationService) PullValuation(ctx context.Context, userDeviceID, deviceDefinitionID, vin string) (core.DataPullStatusEnum, error) {
	const repullWindow = time.Hour * 24 * 14
	if len(vin) != 17 {
		return core.ErrorDataPullStatus, fmt.Errorf("invalid VIN %s", vin)
	}

	deviceDef, err := d.ddSvc.GetDeviceDefinitionByID(ctx, deviceDefinitionID)
	if err != nil {
		return core.ErrorDataPullStatus, err
	}
	localLog := d.log.With().Str("vin", vin).Str("device_definition_id", deviceDefinitionID).Str("user_device_id", userDeviceID).Logger()

	// make sure userdevice exists
	userDevice, err := d.udSvc.GetUserDevice(ctx, userDeviceID)
	if err != nil {
		return core.ErrorDataPullStatus, err
	}
	if userDevice.TokenId == nil {
		return core.ErrorDataPullStatus, fmt.Errorf("valuation pull requires vehicle to have a TokenID. userDeviceID: %s", userDeviceID)
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
		return core.SkippedDataPullStatus, nil
	}

	// by this point we know we might need to insert drivly valuation
	valuation := &models.Valuation{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(deviceDef.DeviceDefinitionId),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
		TokenID:            types.NewNullDecimal(decimal.New(int64(*userDevice.TokenId), 0)),
	}

	// get mileage for the drivly request
	userDeviceData, err := d.uddSvc.GetVehicleRawData(ctx, userDeviceID)
	if err != nil {
		// just warn if can't get data
		localLog.Warn().Err(err).Msgf("could not find any user device data to obtain mileage or location - continuing without")
	}
	deviceMileage := getDeviceMileage(userDeviceData, int(deviceDef.Type.Year), time.Now().Year())
	if deviceMileage == 0 {
		localLog.Warn().Msg("vehicle mileage found was 0 for valuation pull request")
	}

	reqData := ValuationRequestData{
		Mileage: &deviceMileage,
	}

	if userDevice.PostalCode == "" && userDeviceData != nil && len(userDeviceData.Items) > 0 {
		// need to geodecode the postal code
		lat := gjson.GetBytes(userDeviceData.Items[0].SignalsJsonData, "latitude.value").Float()
		long := gjson.GetBytes(userDeviceData.Items[0].SignalsJsonData, "longitude.value").Float()
		localLog.Info().Msgf("lat long found: %f, %f", lat, long)
		if lat > 0 && long > 0 {
			gl, err := d.geoSvc.GeoDecodeLatLong(lat, long)
			if err != nil {
				localLog.Err(err).Msgf("failed to GeoDecode lat long %f, %f", lat, long)
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
		return core.ErrorDataPullStatus, err
	}

	//defer appmetrics.DrivlyIngestTotalOps.Inc()

	return core.PulledValuationDrivlyStatus, nil
}

func (d *drivlyValuationService) PullOffer(ctx context.Context, userDeviceID string) (core.DataPullStatusEnum, error) {
	// make sure userdevice exists
	userDevice, err := d.udSvc.GetUserDevice(ctx, userDeviceID)
	if err != nil {
		return core.ErrorDataPullStatus, err
	}

	if userDevice.Vin == nil || !userDevice.VinConfirmed {
		return core.ErrorDataPullStatus, fmt.Errorf("instant offer feature requires a confirmed VIN")
	}
	if userDevice.TokenId == nil {
		return core.ErrorDataPullStatus, fmt.Errorf("instant offer requires vehicle to have a TokenID. userDeviceID: %s", userDeviceID)
	}
	localLog := d.log.With().Str("vin", *userDevice.Vin).Str("device_definition_id", userDevice.DeviceDefinitionId).Str("user_device_id", userDeviceID).Logger()

	existingOfferData, _ := models.Valuations(
		models.ValuationWhere.Vin.EQ(*userDevice.Vin),
		models.ValuationWhere.OfferMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(ctx, d.dbs().Writer)

	if existingOfferData != nil {
		if existingOfferData.CreatedAt.After(time.Now().Add(-time.Hour * 24 * 30)) {
			return core.SkippedDataPullStatus, fmt.Errorf("instant offer already request in last 30 days")
		}
	}

	deviceDef, err := d.ddSvc.GetDeviceDefinitionByID(ctx, userDevice.DeviceDefinitionId)
	if err != nil {
		return core.ErrorDataPullStatus, err
	}

	// get mileage for the drivly request
	userDeviceData, err := d.uddSvc.GetVehicleRawData(ctx, userDeviceID)
	if err != nil {
		// just warn if can't get data
		localLog.Warn().Err(err).Msgf("could not find any user device data to obtain mileage or location - continuing without")
	}
	deviceMileage := getDeviceMileage(userDeviceData, int(deviceDef.Type.Year), time.Now().Year())

	if deviceMileage == 0 {
		localLog.Warn().Msg("vehicle mileage found was 0")
	}

	params := ValuationRequestData{
		Mileage: &deviceMileage,
		ZipCode: &userDevice.PostalCode,
	}

	offer, err := d.drivlySvc.GetOffersByVIN(*userDevice.Vin, &params)

	if err != nil {
		localLog.Err(err).Msg("error pulling drivly offer data")
		return core.ErrorDataPullStatus, err
	}

	// insert new offer record
	newOffer := &models.Valuation{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(userDevice.DeviceDefinitionId),
		Vin:                *userDevice.Vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
		RequestMetadata:    null.JSON{},
		TokenID:            types.NewNullDecimal(decimal.New(int64(*userDevice.TokenId), 0)),
	}
	_ = newOffer.OfferMetadata.Marshal(offer)

	err = newOffer.Insert(ctx, d.dbs().Writer, boil.Infer())

	if err != nil {
		return core.ErrorDataPullStatus, err
	}
	return core.PulledValuationDrivlyStatus, nil
}

const EstMilesPerYear = 12000.0

func getDeviceMileage(userDeviceData *pb.RawDeviceDataResponse, modelYear int, currentYear int) (mileage float64) {

	if userDeviceData != nil {
		// get the highest odometer found
		odoKm := float64(0)
		for _, item := range userDeviceData.Items {
			odo := gjson.GetBytes(item.SignalsJsonData, "odometer.value").Float()
			if odo > odoKm {
				odoKm = odo
			}
		}
		if odoKm > 0 {
			return odoKm * 0.621271
		}
	}
	// if get here means need to just estimate
	deviceMileage := float64(0)
	yearDiff := currentYear - modelYear
	switch {
	case yearDiff > 0:
		// Past model year
		deviceMileage = float64(yearDiff) * EstMilesPerYear
	case yearDiff == 0:
		// Current model year
		deviceMileage = EstMilesPerYear / 2
	default:
		// Next model year
		deviceMileage = 0
	}

	return deviceMileage
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
