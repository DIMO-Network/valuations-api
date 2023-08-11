package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/valuations-api/internal/api"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/rs/zerolog"
)

func main() {

	gitSha1 := os.Getenv("GIT_SHA1")
	ctx := context.Background()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		log.Fatal("could not load settings: $s", err)
	}
	level, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		log.Fatal("could not parse log level: $s", err)
	}
	logger := zerolog.New(os.Stdout).Level(level).With().
		Timestamp().
		Str("app", settings.ServiceName).
		Str("git-sha1", gitSha1).
		Logger()

	pdb := db.NewDbConnectionFromSettings(ctx, &settings.DB, true)
	// check db ready, this is not ideal btw, the db connection handler would be nicer if it did this.
	totalTime := 0
	for !pdb.IsReady() {
		if totalTime > 30 {
			logger.Fatal().Msg("could not connect to postgres after 30 seconds")
		}
		time.Sleep(time.Second)
		totalTime++
	}

	deps := newDependencyContainer(&settings, logger, pdb.DBS)

	deviceDefsSvc, deviceDefsConn := deps.getDeviceDefinitionService()
	defer deviceDefsConn.Close()
	devicesSvc, devicesConn := deps.getDeviceService()
	defer devicesConn.Close()
	deviceDataSvc, devicedataConn := deps.getDeviceDataService()
	defer devicedataConn.Close()

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&migrateDBCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&loadValuationsCmd{logger: logger,
		settings:      settings,
		deviceDataSvc: deviceDataSvc,
		userDeviceSvc: devicesSvc,
		ddSvc:         deviceDefsSvc,
		pdb:           pdb,
	}, "")

	// Run API
	if len(os.Args) == 1 {

		api.Run(ctx, pdb, logger, &settings, deviceDefsSvc, devicesSvc, deviceDataSvc, deps.getNATSService())
	} else {
		flag.Parse()
		os.Exit(int(subcommands.Execute(ctx)))
	}

}
