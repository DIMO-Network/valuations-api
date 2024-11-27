package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/rs/zerolog"
)

// dependencyContainer way to hold different dependencies we need for our app. We could put all our deps and follow this pattern for everything.
type dependencyContainer struct {
	settings      *config.Settings
	logger        *zerolog.Logger
	ddSvc         services.DeviceDefinitionsAPIService
	userDeviceSvc services.UserDeviceAPIService
	deviceDataSvc services.UserDeviceDataAPIService
	dbs           func() *db.ReaderWriter
}

func newDependencyContainer(settings *config.Settings, logger zerolog.Logger, dbs func() *db.ReaderWriter) dependencyContainer {
	return dependencyContainer{
		settings: settings,
		logger:   &logger,
		dbs:      dbs,
	}
}

func (dc *dependencyContainer) getDeviceDefinitionService() (services.DeviceDefinitionsAPIService, *grpc.ClientConn) {

	definitionsConn, err := grpc.NewClient(dc.settings.DeviceDefinitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		dc.logger.Fatal().Err(err).Str("definitions-api-grpc-addr", dc.settings.DeviceDefinitionsGRPCAddr).
			Msg("failed to dial device definitions grpc")
	}
	dc.ddSvc = services.NewDeviceDefinitionsAPIService(definitionsConn)
	return dc.ddSvc, definitionsConn
}

func (dc *dependencyContainer) getDeviceService() (services.UserDeviceAPIService, *grpc.ClientConn) {
	devicesConn, err := grpc.NewClient(dc.settings.DevicesGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		dc.logger.Fatal().Err(err).Msg("failed to dial devices grpc")
	}
	dc.userDeviceSvc = services.NewUserDeviceService(devicesConn, dc.dbs, dc.logger)
	return dc.userDeviceSvc, devicesConn
}

func (dc *dependencyContainer) getDeviceDataService() (services.UserDeviceDataAPIService, *grpc.ClientConn) {
	deviceDataConn, err := grpc.NewClient(dc.settings.DeviceDataGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		dc.logger.Fatal().Err(err).Msg("failed to dial device data grpc")
	}
	dc.deviceDataSvc = services.NewUserDeviceDataAPIService(deviceDataConn)
	return dc.deviceDataSvc, deviceDataConn
}
