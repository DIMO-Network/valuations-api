package main

import (
	"context"
	"flag"
	"strconv"

	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/valuations-api/internal/core/services"

	"github.com/DIMO-Network/valuations-api/internal/config"

	"github.com/DIMO-Network/valuations-api/internal/core/gateways"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type gqlTelemetryCmd struct {
	logger    zerolog.Logger
	telemetry gateways.TelemetryAPI
	identity  gateways.IdentityAPI
	tokenID   string
	jwt       string
	command   string
	settings  *config.Settings
	dbs       func() *db.ReaderWriter
}

func (*gqlTelemetryCmd) Name() string { return "telemetry" }
func (*gqlTelemetryCmd) Synopsis() string {
	return "telemetry tests our gql client"
}
func (*gqlTelemetryCmd) Usage() string {
	return `telemetry -tokenid <tokenid> -jwt <priv jwt token>`
}

func (p *gqlTelemetryCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.command, "command", "", "command to run: vehicle | telemetry | location")
	f.StringVar(&p.tokenID, "tokenid", "", "vehciel token id")
	f.StringVar(&p.jwt, "jwt", "", "priv token jwt without bearer prefix")
}

func (p *gqlTelemetryCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	tokenID, _ := strconv.ParseUint(p.tokenID, 10, 64)
	p.logger.Info().Msgf("Identity API Url: %s", p.settings.IdentityAPIURL.String())
	p.logger.Info().Msgf("Telemetry API Url: %s", p.settings.TelemetryAPIURL.String())
	p.logger.Info().Msgf("tokenid: %d", tokenID)
	p.logger.Info().Msgf("jwt: %s", p.jwt)

	if p.command == "vehicle" {
		vehicle, err := p.identity.GetVehicle(tokenID)
		if err != nil {
			p.logger.Fatal().Err(err).Msg("could not get vehicle")
		}
		p.logger.Info().Msgf("vehicle: %+v", vehicle)
	}
	if p.command == "telemetry" {
		signals, err := p.telemetry.GetLatestSignals(tokenID, "Bearer "+p.jwt)
		if err != nil {
			p.logger.Fatal().Err(err).Msg("could not get latest signals")
		}
		p.logger.Info().Msgf("signals: %+v", signals)
	}
	if p.command == "location" {
		signals, err := p.telemetry.GetLatestSignals(tokenID, "Bearer "+p.jwt)
		if err != nil {
			p.logger.Fatal().Err(err).Msg("could not get latest signals")
		}
		locationService := services.NewLocationService(p.dbs, p.settings, &p.logger)
		location, err := locationService.GetGeoDecodedLocation(ctx, signals, tokenID)
		if err != nil {
			p.logger.Fatal().Err(err).Msg("could not get location")
		}
		p.logger.Info().Msgf("location: %+v", location)
	}

	return subcommands.ExitSuccess
}
