package services

import (
	"context"
	"database/sql"

	"github.com/DIMO-Network/shared/pkg/logfields"
	"github.com/pkg/errors"

	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	coremodels "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/db/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

//go:generate mockgen -source location_service.go -destination mocks/location_service_mock.go
type LocationService interface {
	GetGeoDecodedLocation(ctx context.Context, signals *coremodels.SignalsLatest, tokenID uint64) (*coremodels.LocationResponse, error)
}

type locationService struct {
	dbs    func() *db.ReaderWriter
	geoSvc GoogleGeoAPIService
	logger *zerolog.Logger
}

func NewLocationService(db func() *db.ReaderWriter, settings *config.Settings, logger *zerolog.Logger) LocationService {
	return &locationService{dbs: db, geoSvc: NewGoogleGeoAPIService(settings, logger), logger: logger}

}

// GetGeoDecodedLocation checks in database if we've already decoded this location, if not pulls new from google and stores in db
func (ls *locationService) GetGeoDecodedLocation(ctx context.Context, signals *coremodels.SignalsLatest, tokenID uint64) (*coremodels.LocationResponse, error) {
	gloc, err := models.GeodecodedLocations(models.GeodecodedLocationWhere.TokenID.EQ(int64(tokenID))).One(ctx, ls.dbs().Reader)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, "failed to query database for geodecoded location")
		}
	}
	if gloc != nil {
		return &coremodels.LocationResponse{
			PostalCode:  gloc.PostalCode.String,
			CountryCode: gloc.Country.String,
		}, nil
	}
	// guard
	if signals == nil {
		return nil, errors.New("no signals provided")
	}
	if signals.CurrentLocationLatitude.Value == 0 && signals.CurrentLocationLongitude.Value == 0 {
		return nil, errors.New("no location provided, lat long zero")
	}
	// decode the lat long with google
	gl, err := ls.geoSvc.GeoDecodeLatLong(signals.CurrentLocationLatitude.Value, signals.CurrentLocationLongitude.Value)
	if err != nil {
		return nil, err
	}
	if gl == nil {
		return nil, errors.New("no information found when decoding lat long to postal code for valuation request")
	}
	gloc = &models.GeodecodedLocation{
		TokenID:    int64(tokenID),
		PostalCode: null.StringFrom(gl.PostalCode),
		Country:    null.StringFrom(gl.Country),
	}
	err = gloc.Insert(ctx, ls.dbs().Writer, boil.Infer())
	if err != nil {
		ls.logger.Err(err).Uint64(logfields.VehicleTokenID, tokenID).Msgf("failed to insert geodecoded location")
	}
	return &coremodels.LocationResponse{
		PostalCode:  gl.PostalCode,
		CountryCode: gl.Country,
	}, nil
}
