package rpc

import (
	"context"

	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	pb "github.com/DIMO-Network/valuations-api/pkg/grpc"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Euro to USD conversion rate, used for calculating the price of the device, hardcoded for now
const (
	euroToUsd float64 = 1.10
)

type valuationsService struct {
	pb.UnimplementedValuationsServiceServer
	dbs      func() *db.ReaderWriter
	settings *config.Settings
	logger   *zerolog.Logger
}

func NewValuationsService(dbs func() *db.ReaderWriter, settings *config.Settings, logger *zerolog.Logger) pb.ValuationsServiceServer {
	return &valuationsService{
		dbs:      dbs,
		settings: settings,
		logger:   logger,
	}
}

func (s *valuationsService) GetAllUserDeviceValuation(ctx context.Context, _ *emptypb.Empty) (*pb.ValuationResponse, error) {

	query := `select sum(evd.retail_price) as total_retail,
					 sum(evd.vincario_price) as total_vincario
					 from
                             (
								select distinct on (vin) vin, 
														pricing_metadata, 
														jsonb_path_query(evd.pricing_metadata, '$.retail.kelley.book')::decimal as retail_price,
														jsonb_path_query(evd.vincario_metadata, '$.market_price.price_avg')::decimal as vincario_price,
														created_at
       							from external_vin_data evd 
								order by vin, created_at desc
							) as evd;`

	queryGrowth := `select sum(evd.retail_price) as total_retail,
					 sum(evd.vincario_price) as total_vincario
					 from
						(
							select distinct on (vin) vin, 
													pricing_metadata, 
													jsonb_path_query(evd.pricing_metadata, '$.retail.kelley.book')::decimal as retail_price, 
													jsonb_path_query(evd.vincario_metadata, '$.market_price.price_avg')::decimal as vincario_price,
													created_at
							from external_vin_data evd 
							where created_at > current_date - 7
							order by vin, created_at desc
						) as evd;`

	type Result struct {
		TotalRetail   null.Float64 `boil:"total_retail"`
		TotalVincario null.Float64 `boil:"total_vincario"`
	}
	var total Result
	var lastWeek Result

	err := queries.Raw(query).Bind(ctx, s.dbs().Reader, &total)
	if err != nil {
		s.logger.Err(err).Msg("Database failure retrieving total valuation.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	err = queries.Raw(queryGrowth).Bind(ctx, s.dbs().Reader, &lastWeek)
	if err != nil {
		s.logger.Err(err).Msg("Database failure retrieving last week valuation.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	totalValuation := total.TotalRetail.Float64 // 0 by default
	growthPercentage := 0.0

	if !total.TotalVincario.IsZero() {
		totalValuation += total.TotalVincario.Float64 * euroToUsd
	}

	if totalValuation > 0 {
		totalLastWeek := lastWeek.TotalRetail.Float64

		if !lastWeek.TotalVincario.IsZero() {
			totalLastWeek += lastWeek.TotalVincario.Float64 * euroToUsd
		}
		growthPercentage = (totalLastWeek / totalValuation) * 100
	}

	// todo: get an average valuation per vehicle, and multiply for whatever count of vehicles we did not get value for

	return &pb.ValuationResponse{
		Total:            float32(totalValuation),
		GrowthPercentage: float32(growthPercentage),
	}, nil
}
