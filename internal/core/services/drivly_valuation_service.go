package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/valuations-api/internal/core/gateways"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"time"

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
	PullValuation(ctx context.Context, tokenID uint64, definitionID, vin string) (core.DataPullStatusEnum, error)
	PullOffer(ctx context.Context, tokenID uint64, vin string) (core.DataPullStatusEnum, error)
}

type drivlyValuationService struct {
	dbs          func() *db.ReaderWriter
	drivlySvc    DrivlyAPIService
	udSvc        UserDeviceAPIService
	geoSvc       GoogleGeoAPIService
	identityAPI  gateways.IdentityAPI
	telemetryAPI gateways.TelemetryAPI
	log          *zerolog.Logger
}

func NewDrivlyValuationService(DBS func() *db.ReaderWriter, log *zerolog.Logger, settings *config.Settings, udSvc UserDeviceAPIService) DrivlyValuationService {
	return &drivlyValuationService{
		dbs:          DBS,
		log:          log,
		drivlySvc:    NewDrivlyAPIService(settings, DBS),
		geoSvc:       NewGoogleGeoAPIService(settings),
		udSvc:        udSvc,
		identityAPI:  gateways.NewIdentityAPIService(log, settings, nil),
		telemetryAPI: gateways.NewTelemetryAPI(log, settings, nil),
	}
}

// PullValuation performs a data pull for a vehicle valuation. It retrieves pricing and
// other relevant data for a given VIN. Not necessary for the userDevice to exist, VIN is what matters
func (d *drivlyValuationService) PullValuation(ctx context.Context, tokenID uint64, definitionID, vin string) (core.DataPullStatusEnum, error) {
	const repullWindow = time.Hour * 24 * 14
	if len(vin) != 17 {
		return core.ErrorDataPullStatus, fmt.Errorf("invalid VIN %s", vin)
	}

	// todo switch to use identity-api
	deviceDef, err := d.identityAPI.GetDefinition(definitionID)
	if err != nil {
		return core.ErrorDataPullStatus, err
	}
	localLog := d.log.With().Str("vin", vin).Str("definition_id", definitionID).Uint64("token_id", tokenID).Logger()

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
		ID:           ksuid.New().String(),
		Vin:          vin,
		TokenID:      types.NewNullDecimal(decimal.New(int64(tokenID), 0)),
		DefinitionID: null.StringFrom(definitionID), // todo make this not nullable
	}

	// get mileage for the drivly request
	signals, err := d.telemetryAPI.GetLatestSignals(tokenID)
	if err != nil {
		d.log.Warn().Err(err).Uint64("token_id", tokenID).Msgf("could not get telemetry latest signals for token %d", tokenID)
	}

	deviceMileage := getDeviceMileage(signals, deviceDef.Year, time.Now().Year())
	if deviceMileage == 0 {
		localLog.Warn().Msg("vehicle mileage found was 0 for valuation pull request")
	}

	reqData := ValuationRequestData{
		Mileage: &deviceMileage,
	}
	// handle postal code information to send to drivly and store for future use
	gloc, err := models.GeodecodedLocations(models.GeodecodedLocationWhere.TokenID.EQ(int64(tokenID))).One(ctx, d.dbs().Reader)
	if err != nil {
		d.log.Warn().Err(err).Msgf("failed to get geodecoded location for token %d", tokenID)
	}
	if gloc != nil {
		reqData.ZipCode = &gloc.PostalCode.String
	} else {
		if signals != nil && signals.CurrentLocationLatitude.Value > 0 && signals.CurrentLocationLongitude.Value > 0 {
			// decode the lat long if we have it
			gl, err := d.geoSvc.GeoDecodeLatLong(signals.CurrentLocationLatitude.Value, signals.CurrentLocationLongitude.Value)
			if err != nil {
				d.log.Warn().Err(err).Msgf("failed to GeoDecode lat long %f, %f", signals.CurrentLocationLatitude.Value, signals.CurrentLocationLongitude.Value)
			} else {
				reqData.ZipCode = &gl.PostalCode
				// persist the info for future
				gloc = &models.GeodecodedLocation{
					TokenID:    int64(tokenID),
					PostalCode: null.StringFrom(gl.PostalCode),
					Country:    null.StringFrom(gl.Country),
				}
				err = gloc.Insert(ctx, d.dbs().Writer, boil.Infer())
				if err != nil {
					d.log.Err(err).Msgf("failed to insert geodecoded location for token %d", tokenID)
				}
			}
		}
	}
	// add the request data to the valution record
	_ = valuation.RequestMetadata.Marshal(reqData)
	// cal drivly for pricing
	pricing, err := d.drivlySvc.GetVINPricing(vin, &reqData)
	if err == nil {
		_ = valuation.DrivlyPricingMetadata.Marshal(pricing)
	}

	err = valuation.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return core.ErrorDataPullStatus, err
	}

	//defer appmetrics.DrivlyIngestTotalOps.Inc()

	return core.PulledValuationDrivlyStatus, nil
}

func (d *drivlyValuationService) PullOffer(ctx context.Context, tokenID uint64, vin string) (core.DataPullStatusEnum, error) {
	// make sure userdevice exists
	vehicle, err := d.identityAPI.GetVehicle(tokenID)
	if err != nil {
		return core.ErrorDataPullStatus, err
	}

	if len(vin) != 17 {
		return core.ErrorDataPullStatus, fmt.Errorf("invalid VIN %s", vin)
	}

	localLog := d.log.With().Str("vin", vin).Str("device_definition_id", vehicle.Definition.Id).Uint64("token_id", tokenID).Logger()

	existingOfferData, _ := models.Valuations(
		models.ValuationWhere.Vin.EQ(vin),
		models.ValuationWhere.OfferMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(ctx, d.dbs().Writer)

	if existingOfferData != nil {
		if existingOfferData.CreatedAt.After(time.Now().Add(-time.Hour * 24 * 30)) {
			return core.SkippedDataPullStatus, fmt.Errorf("instant offer already request in last 30 days")
		}
	}
	// future: pull by tokenID from identity-api
	deviceDef, err := d.identityAPI.GetDefinition(vehicle.Definition.Id)
	if err != nil {
		return core.ErrorDataPullStatus, err
	}

	// get mileage for the drivly request
	signals, err := d.telemetryAPI.GetLatestSignals(tokenID)
	if err != nil {
		// just warn if can't get data
		localLog.Warn().Err(err).Msgf("could not find any telemtry data to obtain mileage or location - continuing without")
	}
	deviceMileage := getDeviceMileage(signals, deviceDef.Year, time.Now().Year())

	if deviceMileage == 0 {
		localLog.Warn().Msg("vehicle mileage found was 0")
	}

	params := ValuationRequestData{
		Mileage: &deviceMileage,
	}
	gloc, _ := models.GeodecodedLocations(models.GeodecodedLocationWhere.TokenID.EQ(int64(tokenID))).One(ctx, d.dbs().Reader)
	if gloc != nil {
		params.ZipCode = &gloc.PostalCode.String
	}

	offer, err := d.drivlySvc.GetOffersByVIN(vin, &params)

	if err != nil {
		localLog.Err(err).Msg("error pulling drivly offer data")
		return core.ErrorDataPullStatus, err
	}

	// insert new offer record
	newOffer := &models.Valuation{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(vehicle.Definition.Id),
		Vin:                vin,
		TokenID:            types.NewNullDecimal(decimal.New(int64(tokenID), 0)),
	}
	pj, err := json.Marshal(params)
	if err == nil {
		newOffer.OfferMetadata = null.JSONFrom(pj)
	}
	_ = newOffer.OfferMetadata.Marshal(offer)

	err = newOffer.Insert(ctx, d.dbs().Writer, boil.Infer())

	if err != nil {
		return core.ErrorDataPullStatus, err
	}
	return core.PulledValuationDrivlyStatus, nil
}

const EstMilesPerYear = 12000.0

func getDeviceMileage(signals *gateways.SignalsLatest, modelYear int, currentYear int) (mileage float64) {
	if signals != nil {
		odoKm := signals.PowertrainTransmissionTravelledDistance.Value
		if odoKm > 0 {
			return odoKm * 0.621271
		} // convert to miles
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
