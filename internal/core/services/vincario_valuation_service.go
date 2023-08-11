package services

import (
	"context"

	"strings"
	"time"

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

//go:generate mockgen -source vincario_valuation_service.go -destination mocks/vincario_valuation_service_mock.go

type VincarioValuationService interface {
	PullValuation(ctx context.Context, userDeiceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error)
}

type vincarioValuationService struct {
	dbs         func() *db.ReaderWriter
	log         *zerolog.Logger
	vincarioSvc VincarioAPIService
	udSvc       UserDeviceAPIService
}

func NewVincarioValuationService(DBS func() *db.ReaderWriter, log *zerolog.Logger, settings *config.Settings, udSvc UserDeviceAPIService) VincarioValuationService {
	return &vincarioValuationService{
		dbs:         DBS,
		log:         log,
		vincarioSvc: NewVincarioAPIService(settings, log),
		udSvc:       udSvc,
	}
}

func (d *vincarioValuationService) PullValuation(ctx context.Context, userDeviceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error) {
	const repullWindow = time.Hour * 24 * 14
	if len(vin) != 17 {
		return ErrorDataPullStatus, errors.Errorf("invalid VIN %s", vin)
	}

	// make sure userdevice exists
	ud, err := d.udSvc.GetUserDevice(ctx, userDeviceID)
	if err != nil {
		return ErrorDataPullStatus, err
	}
	// do not pull for USA
	if strings.EqualFold(ud.CountryCode, "USA") {
		return SkippedDataPullStatus, nil
	}

	// check repull window
	existingPricingData, _ := models.Valuations(
		models.ValuationWhere.Vin.EQ(vin),
		models.ValuationWhere.VincarioMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)

	// just return if already pulled recently for this VIN, but still need to insert never pulled vin - should be uncommon scenario
	if existingPricingData != nil && existingPricingData.UpdatedAt.Add(repullWindow).After(time.Now()) {
		return SkippedDataPullStatus, nil
	}

	externalVinData := &models.Valuation{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(deviceDefinitionID),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}

	valuation, err := d.vincarioSvc.GetMarketValuation(vin)

	if err != nil {
		return ErrorDataPullStatus, errors.Wrap(err, "error pulling market data from vincario")
	}
	err = externalVinData.VincarioMetadata.Marshal(valuation)
	if err != nil {
		return ErrorDataPullStatus, errors.Wrap(err, "error marshalling vincario responset")
	}
	err = externalVinData.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return ErrorDataPullStatus, errors.Wrap(err, "error inserting external_vin_data for vincario")
	}

	return PulledValuationVincarioStatus, nil
}

type DataPullStatusEnum string

const (
	// PulledInfoAndValuationStatus means we pulled vin, edmunds, build and valuations
	PulledInfoAndValuationStatus DataPullStatusEnum = "PulledAll"
	// PulledValuationDrivlyStatus means we only pulled offers and pricing
	PulledValuationDrivlyStatus   DataPullStatusEnum = "PulledValuations"
	PulledValuationVincarioStatus DataPullStatusEnum = "PulledValuationVincario"
	SkippedDataPullStatus         DataPullStatusEnum = "Skipped"
	ErrorDataPullStatus           DataPullStatusEnum = "Error"
)
