package services

import (
	"context"
	"github.com/DIMO-Network/valuations-api/internal/core/gateways"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"strings"
	"time"

	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
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
	PullValuation(ctx context.Context, tokenID uint64, definitionID, vin string) (core.DataPullStatusEnum, error)
}

type vincarioValuationService struct {
	dbs         func() *db.ReaderWriter
	log         *zerolog.Logger
	vincarioSvc VincarioAPIService
	identityAPI gateways.IdentityAPI
}

func NewVincarioValuationService(DBS func() *db.ReaderWriter, log *zerolog.Logger, settings *config.Settings, identityAPI gateways.IdentityAPI) VincarioValuationService {
	return &vincarioValuationService{
		dbs:         DBS,
		log:         log,
		vincarioSvc: NewVincarioAPIService(settings, log),
		identityAPI: identityAPI,
	}
}

// PullValuation ideally we pass country code into here
func (d *vincarioValuationService) PullValuation(ctx context.Context, tokenID uint64, definitionID, vin string) (core.DataPullStatusEnum, error) {
	const repullWindow = time.Hour * 24 * 30 // one month
	if len(vin) != 17 {
		return core.ErrorDataPullStatus, errors.Errorf("invalid VIN %s", vin)
	}

	// make sure userdevice exists
	vehicle, err := d.identityAPI.GetVehicle(tokenID)
	if err != nil {
		return core.ErrorDataPullStatus, err
	}
	// do not pull for USA
	gloc, err := models.GeodecodedLocations(models.GeodecodedLocationWhere.TokenID.EQ(int64(tokenID))).One(ctx, d.dbs().Reader)
	countryCode := ""
	if gloc != nil {
		countryCode = gloc.Country.String
	}
	if strings.EqualFold(countryCode, "USA") {
		return core.SkippedDataPullStatus, nil
	}

	// check repull window
	existingPricingData, _ := models.Valuations(
		models.ValuationWhere.Vin.EQ(vin),
		models.ValuationWhere.VincarioMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)

	// just return if already pulled recently for this VIN, but still need to insert never pulled vin - should be uncommon scenario
	if existingPricingData != nil && existingPricingData.UpdatedAt.Add(repullWindow).After(time.Now()) {
		return core.SkippedDataPullStatus, nil
	}

	externalVinData := &models.Valuation{
		ID:           ksuid.New().String(),
		DefinitionID: null.StringFrom(vehicle.Definition.Id),
		Vin:          vin,
		// at some point change the db datatype to bigint
		TokenID: types.NewNullDecimal(decimal.New(int64(tokenID), 0)),
	}

	valuation, err := d.vincarioSvc.GetMarketValuation(vin)
	if err != nil {
		return core.ErrorDataPullStatus, errors.Wrap(err, "error pulling market data from vincario")
	}

	err = externalVinData.VincarioMetadata.Marshal(valuation)
	if err != nil {
		return core.ErrorDataPullStatus, errors.Wrap(err, "error marshalling vincario responset")
	}

	err = externalVinData.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return core.ErrorDataPullStatus, errors.Wrap(err, "error inserting external_vin_data for vincario")
	}

	return core.PulledValuationVincarioStatus, nil
}
