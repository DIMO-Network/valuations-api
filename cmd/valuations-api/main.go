package main

import (
	"context"
	"flag"
	"github.com/DIMO-Network/shared/pkg/settings"
	"github.com/DIMO-Network/valuations-api/internal/app"
	"github.com/DIMO-Network/valuations-api/internal/core/gateways"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"time"

	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/google/subcommands"

	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/rs/zerolog"
)

// @title                       DIMO Vehicle Valuations API
// @description 				API to get latest valuation for a given connected vehicle belonging to user. Tokens must be privilege tokens.
// @version                     1.0
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func main() {

	gitSha1 := os.Getenv("GIT_SHA1")
	ctx := context.Background()

	cfg, err := settings.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		log.Fatal("could not load settings: $s", err)
	}
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal("could not parse log level: $s", err)
	}
	logger := zerolog.New(os.Stdout).Level(level).With().
		Timestamp().
		Str("app", cfg.ServiceName).
		Str("git-sha1", gitSha1).
		Logger()

	pdb := db.NewDbConnectionFromSettings(ctx, &cfg.DB, true)
	// check db ready, this is not ideal btw, the db connection handler would be nicer if it did this.
	totalTime := 0
	for !pdb.IsReady() {
		if totalTime > 30 {
			logger.Fatal().Msg("could not connect to postgres after 30 seconds")
		}
		time.Sleep(time.Second)
		totalTime++
	}
	identityAPI := gateways.NewIdentityAPIService(&logger, &cfg)
	telemetryAPI := gateways.NewTelemetryAPI(&logger, &cfg)
	locationSvc := services.NewLocationService(pdb.DBS, &cfg, &logger)
	devicesConn, err := grpc.NewClient(cfg.DevicesGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to dial devices grpc")
	}
	userDeviceSvc := services.NewUserDeviceService(devicesConn, pdb.DBS, &logger, locationSvc, telemetryAPI)

	defer devicesConn.Close()

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&migrateDBCmd{logger: logger, settings: cfg}, "")
	subcommands.Register(&loadValuationsCmd{logger: logger,
		settings: cfg,
		pdb:      pdb,
	}, "")

	// Run API
	if len(os.Args) == 1 {
		app.Run(ctx, pdb, logger, &cfg, identityAPI, userDeviceSvc, telemetryAPI, locationSvc)
	} else {
		flag.Parse()
		os.Exit(int(subcommands.Execute(ctx)))
	}
}
