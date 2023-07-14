package main

import (
	"context"
	"flag"

	"os"

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
}

func (*loadValuationsCmd) Name() string     { return "valuations-pull" }
func (*loadValuationsCmd) Synopsis() string { return "valuations-pull args to stdout." }
func (*loadValuationsCmd) Usage() string {
	return `valuations-pull:
	valuations-pull args.
  `
}

// nolint
func (p *loadValuationsCmd) SetFlags(f *flag.FlagSet) {

}

func (p *loadValuationsCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.logger.Info().Msgf("Pull VIN info, valuations and pricing from driv.ly for USA and valuations from Vincario for EUR")
	wmi := ""
	if len(os.Args) > 2 {
		// parse out vin WMI code to filter on
		for i, a := range os.Args {
			if a == "--wmi" {
				wmi = os.Args[i+1]
				break
			}
		}
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
