package main

import (
	"context"
	"flag"

	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type loadValuationsCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
	wmi      string
}

func (*loadValuationsCmd) Name() string { return "pull-valuations" }
func (*loadValuationsCmd) Synopsis() string {
	return "pull-valuations runs through all connected cars and gets their valuation"
}
func (*loadValuationsCmd) Usage() string {
	return `pull-valuations -wmi <WMI 3 char>`
}

func (p *loadValuationsCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.wmi, "wmi", "", "WMI filter option to only get valuations for these")
}

func (p *loadValuationsCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.logger.Info().Msgf("Pull VIN info, valuations and pricing from driv.ly for USA and valuations from Vincario for EUR")

	// todo pending implement, re-think business / product value. This should be taken over by DINC.
	// Send a notification when new valuation comes up. Run valuations every 3 months.

	return subcommands.ExitSuccess
}
