package commands

import (
	"context"
	"fmt"
	"sync"

	"github.com/DIMO-Network/devices-api/pkg/grpc"

	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	core "github.com/DIMO-Network/valuations-api/internal/core/models"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source run_valuation.go -destination mocks/run_valuation_mock.go

type LoadVinVerifiedValuationCommandHandler interface {
	Execute(ctx context.Context, command *LoadVinVerifiedValuationCommandRequest) error
}

type loadVinVerifiedValuationCommandHandler struct {
	logger                   zerolog.Logger
	userDeviceService        services.UserDeviceAPIService
	vincarioValuationService services.VincarioValuationService
	drivlyValuationService   services.DrivlyValuationService
}

type LoadVinVerifiedValuationCommandRequest struct {
	WMI string `json:"wmi"`
}

func NewLoadVinVerifiedValuationCommandHandler(dbs func() *db.ReaderWriter, logger zerolog.Logger, settings *config.Settings,
	userDeviceService services.UserDeviceAPIService,
	ddSvc services.DeviceDefinitionsAPIService,
	uddSvc services.UserDeviceDataAPIService) LoadVinVerifiedValuationCommandHandler {
	return loadVinVerifiedValuationCommandHandler{
		logger:                   logger,
		userDeviceService:        userDeviceService,
		vincarioValuationService: services.NewVincarioValuationService(dbs, &logger, settings, userDeviceService),
		drivlyValuationService:   services.NewDrivlyValuationService(dbs, &logger, settings, ddSvc, uddSvc, userDeviceService),
	}
}

const maxConcurrency = 5

func (h loadVinVerifiedValuationCommandHandler) Execute(ctx context.Context, command *LoadVinVerifiedValuationCommandRequest) error {
	h.logger.Info().Msg("Starting Valuations Pull. Getting list of User Devices...")
	all, err := h.userDeviceService.GetAllUserDevice(ctx, command.WMI)
	h.logger.Info().Msgf("User Devices found: %d - Now let's work through them", len(all))

	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	statsAggr := map[core.DataPullStatusEnum]int{}
	// process requests in parallel
	for i, ud := range all {
		wg.Add(1)
		go func(c context.Context, userDevice *grpc.UserDevice, _ int) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire a token
			status, _ := h.requestValuation(c, userDevice)
			statsAggr[status]++
			<-semaphore // Release the token
		}(ctx, ud, i)
	}
	wg.Wait()

	fmt.Println("-------------------RUN SUMMARY--------------------------")
	// colorize each result
	fmt.Printf("Total VINs processed: %d \n", len(all))
	fmt.Printf("New Drivly Pulls (vin + valuations): %d \n", statsAggr[core.PulledInfoAndValuationStatus])
	fmt.Printf("Pulled New Pricing & Offers: %d \n", statsAggr[core.PulledValuationDrivlyStatus])
	fmt.Printf("Skipped VIN due to biz logic: %d \n", statsAggr[core.SkippedDataPullStatus])
	fmt.Printf("Pulled New Vincario Market Valuation: %d \n", statsAggr[core.PulledValuationVincarioStatus])
	fmt.Printf("Skipped VIN due to error: %d \n", statsAggr[""])
	fmt.Println("--------------------------------------------------------")
	return nil
}

func (h loadVinVerifiedValuationCommandHandler) requestValuation(ctx context.Context, ud *grpc.UserDevice) (core.DataPullStatusEnum, error) {
	fmt.Printf("Pulling valuation: https://admin.team.dimo.zone/user-devices/%s, country: %s, token_id: %d. Status: ", ud.Id, ud.CountryCode, ud.TokenId)
	if ud.CountryCode == "USA" || ud.CountryCode == "CAN" || ud.CountryCode == "MEX" || ud.CountryCode == "PRI" {
		status, err := h.drivlyValuationService.PullValuation(ctx, ud.Id, ud.DeviceDefinitionId, *ud.Vin)
		fmt.Printf("drivly %s \n", status)
		if err != nil {
			h.logger.Err(err).Str("vin", *ud.Vin).Msg("error pulling drivly data")
		}
		return status, nil
	}
	status, err := h.vincarioValuationService.PullValuation(ctx, ud.Id, ud.DeviceDefinitionId, *ud.Vin)
	fmt.Printf("vincario %s \n", status)
	if err != nil {
		h.logger.Err(err).Str("vin", *ud.Vin).Msg("error pulling vincario data")
	}
	return status, nil
}
