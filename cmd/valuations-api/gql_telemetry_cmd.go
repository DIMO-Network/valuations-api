package main

import (
	"context"
	"flag"
	"strconv"

	"github.com/DIMO-Network/valuations-api/internal/core/gateways"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type gqlTelemetryCmd struct {
	logger zerolog.Logger
	//settings  config.Settings
	//pdb       db.Store
	telemetry gateways.TelemetryAPI
	tokenID   string
	jwt       string
}

func (*gqlTelemetryCmd) Name() string { return "telemetry" }
func (*gqlTelemetryCmd) Synopsis() string {
	return "telemetry tests our gql client"
}
func (*gqlTelemetryCmd) Usage() string {
	return `telemetry -tokenid <tokenid> -jwt <priv jwt token>`
}

func (p *gqlTelemetryCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.tokenID, "tokenid", "", "vehciel token id")
	f.StringVar(&p.jwt, "jwt", "", "priv token jwt")
}

func (p *gqlTelemetryCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	tokenID, _ := strconv.ParseUint(p.tokenID, 10, 64)
	p.logger.Info().Msgf("tokenid: %d", tokenID)

	signals, err := p.telemetry.GetLatestSignals(ctx, tokenID, "Bearer "+p.jwt)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("could not get latest signals")
	}
	p.logger.Info().Msgf("signals: %+v", signals)

	return subcommands.ExitSuccess
}
