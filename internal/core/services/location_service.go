package services

import (
	"context"

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
	return &locationService{dbs: db, geoSvc: NewGoogleGeoAPIService(settings), logger: logger}

}

// GetGeoDecodedLocation checks in database if we've already decoded this location, if not pulls new from google and stores in db
func (ls *locationService) GetGeoDecodedLocation(ctx context.Context, signals *coremodels.SignalsLatest, tokenID uint64) (*coremodels.LocationResponse, error) {
	gloc, err := models.GeodecodedLocations(models.GeodecodedLocationWhere.TokenID.EQ(int64(tokenID))).One(ctx, ls.dbs().Reader)
	if err != nil {
		return nil, err
	}
	if gloc != nil {
		return &coremodels.LocationResponse{
			PostalCode:  gloc.PostalCode.String,
			CountryCode: gloc.Country.String,
		}, nil
	}
	if signals != nil && signals.CurrentLocationLatitude.Value > 0 && signals.CurrentLocationLongitude.Value > 0 {
		// decode the lat long if we have it
		gl, err := ls.geoSvc.GeoDecodeLatLong(signals.CurrentLocationLatitude.Value, signals.CurrentLocationLongitude.Value)
		if err != nil {
			return nil, err
		}
		gloc = &models.GeodecodedLocation{
			TokenID:    int64(tokenID),
			PostalCode: null.StringFrom(gl.PostalCode),
			Country:    null.StringFrom(gl.Country),
		}
		err = gloc.Insert(ctx, ls.dbs().Writer, boil.Infer())
		if err != nil {
			ls.logger.Err(err).Msgf("failed to insert geodecoded location for token %d", tokenID)
		}
		return &coremodels.LocationResponse{
			PostalCode:  gl.PostalCode,
			CountryCode: gl.Country,
		}, nil
	}

	return nil, nil
}
