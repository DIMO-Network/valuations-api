package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/rs/zerolog"
)

// dependencyContainer way to hold different dependencies we need for our app. We could put all our deps and follow this pattern for everything.
type dependencyContainer struct {
	settings      *config.Settings
	logger        *zerolog.Logger
	userDeviceSvc services.UserDeviceAPIService
	dbs           func() *db.ReaderWriter
}

func newDependencyContainer(settings *config.Settings, logger zerolog.Logger, dbs func() *db.ReaderWriter) dependencyContainer {
	return dependencyContainer{
		settings: settings,
		logger:   &logger,
		dbs:      dbs,
	}
}

func (dc *dependencyContainer) getDeviceService() (services.UserDeviceAPIService, *grpc.ClientConn) {
	devicesConn, err := grpc.NewClient(dc.settings.DevicesGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		dc.logger.Fatal().Err(err).Msg("failed to dial devices grpc")
	}
	dc.userDeviceSvc = services.NewUserDeviceService(devicesConn, dc.dbs, dc.logger)
	return dc.userDeviceSvc, devicesConn
}
