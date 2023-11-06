package main

import (
	"context"
	"flag"

	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/DIMO-Network/valuations-api/internal/core/commands"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type loadValuationsCmd struct {
	logger        zerolog.Logger
	settings      config.Settings
	pdb           db.Store
	ddSvc         services.DeviceDefinitionsAPIService
	userDeviceSvc services.UserDeviceAPIService
	deviceDataSvc services.UserDeviceDataAPIService
	wmi           *string
}

func (*loadValuationsCmd) Name() string { return "pull-valuations" }
func (*loadValuationsCmd) Synopsis() string {
	return "pull-valuations runs through all connected cars and gets their valuation"
}
func (*loadValuationsCmd) Usage() string {
	return `pull-valuations -wmi <WMI 3 char>`
}

// nolint
func (p *loadValuationsCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(p.wmi, "wmi", "", "WMI filter option to only get valuations for these")
}

func (p *loadValuationsCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.logger.Info().Msgf("Pull VIN info, valuations and pricing from driv.ly for USA and valuations from Vincario for EUR")
	wmi := ""
	if p.wmi != nil {
		wmi = *p.wmi
	}

	handler := commands.NewLoadVinVerifiedValuationCommandHandler(p.pdb.DBS, p.logger, &p.settings, p.userDeviceSvc, p.ddSvc, p.deviceDataSvc)
	err := handler.Execute(ctx, &commands.LoadVinVerifiedValuationCommandRequest{
		WMI: wmi,
	})
	if err != nil {
		p.logger.Fatal().Err(err).Msg("error trying to pull valuations")
	}

	return subcommands.ExitSuccess
}
