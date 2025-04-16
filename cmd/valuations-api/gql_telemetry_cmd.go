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

	signals, err := p.telemetry.GetLatestSignals(ctx, tokenID, "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjhiNTRkYTRmNDk2MWIwYzYzYjdiMTU3ZmMzZGM3MzgyYzFiYzdkMjAifQ.eyJhdWQiOlsiZGltby56b25lIl0sImNvbnRyYWN0X2FkZHJlc3MiOiIweGJhNTczOGExOGQ4M2Q0MTg0N2RmZmJkYzYxMDFkMzdjNjljOWIwY2YiLCJleHAiOjE3NDQ4NDc2MjMsImlhdCI6MTc0NDg0NzAyMywiaXNzIjoiaHR0cHM6Ly9hdXRoLXJvbGVzLXJpZ2h0cy5kaW1vLnpvbmUiLCJwcml2aWxlZ2VfaWRzIjpbMSwyLDMsNCw1LDZdLCJzdWIiOiIweGJBNTczOGExOGQ4M0Q0MTg0N2RmRmJEQzYxMDFkMzdDNjljOUIwY0YvMzY4MSIsInRva2VuX2lkIjoiMzY4MSJ9.NcqWyznhwQbkNmdVrW-p_f7S6eQcEFsLkk8UfBnMfAAbDWX7krcmmiNdoACvkMZb2RKLyYrmreF4lGi3LC-GJrApXTlH4ca51yOTq3aOTYylc0HApMdj9GQcW7RiNVbNNVyKe3JX2Mvuek20Z9ld0K5jWACdqIY7rLihPp-LS5NyXq-yEW-iGGTN6-CI-kyaVYvkxKaaM7CyoXP7NXmapNHTxQRynZ-Dkm9Ng5oukNqRmp5uD08f91q05AR9mn9Mv-LrNJXJjKtBKE-BclYH6WqiYDCr3F_B7Jz6OOidyBupDIDLUNVnnoaMhgQIepyxXLt8M0lAAa6ZXWu34bqyyQ")
	if err != nil {
		p.logger.Fatal().Err(err).Msg("could not get latest signals")
	}
	p.logger.Info().Msgf("signals: %+v", signals)

	return subcommands.ExitSuccess
}
