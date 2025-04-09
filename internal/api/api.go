package api

import (
	"context"
	"encoding/json"
	"github.com/DIMO-Network/valuations-api/internal/core/gateways"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/DIMO-Network/shared"
	"github.com/IBM/sarama"
	"github.com/burdiyan/kafkautil"
	"github.com/lovoo/goka"

	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/ethereum/go-ethereum/common"

	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/DIMO-Network/valuations-api/internal/controllers"
	"github.com/DIMO-Network/valuations-api/internal/controllers/helpers"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/metrics"
	"github.com/DIMO-Network/valuations-api/internal/rpc"
	pb "github.com/DIMO-Network/valuations-api/pkg/grpc"
	"github.com/gofiber/adaptor/v2"
	jwtware "github.com/gofiber/contrib/jwt"
	fiber "github.com/gofiber/fiber/v2"
	cors "github.com/gofiber/fiber/v2/middleware/cors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	// docs are generated by Swag CLI, you have to import them.
	_ "github.com/DIMO-Network/valuations-api/docs"
)

func Run(ctx context.Context, pdb db.Store, logger zerolog.Logger, settings *config.Settings, identity gateways.IdentityAPI,
	userDeviceSvc services.UserDeviceAPIService) {

	// mint events consumer to request valuations and offers for new paired vehicles
	// removing this for now b/c the events topic produces way too many messages & duplicates, we need something that only emits once on new mints
	startEventsConsumer(settings, logger, pdb, userDeviceSvc, identity, deviceDataSvc)

	startMonitoringServer(logger, settings)
	go startGRCPServer(pdb, logger, settings, userDeviceSvc)

	drivlySvc := services.NewDrivlyValuationService(pdb.DBS, &logger, settings, ddSvc, deviceDataSvc, userDeviceSvc)
	vincarioSvc := services.NewVincarioValuationService(pdb.DBS, &logger, settings, userDeviceSvc)
	app := startWebAPI(logger, settings, userDeviceSvc, drivlySvc, vincarioSvc)
	// nolint
	defer app.Shutdown()

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent with length of 1
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
	<-c                                             // This blocks the main thread until an interrupt is received
	logger.Info().Msg("Gracefully shutting down and running cleanup tasks...")
	_ = ctx.Done()
}

// startEventsConsumer listens to kafka topic configured by EVENTS_TOPIC and processes vehicle nft mint events to trigger new valuations
func startEventsConsumer(settings *config.Settings, logger zerolog.Logger, pdb db.Store, userDeviceSvc services.UserDeviceAPIService,
	ddSvc services.DeviceDefinitionsAPIService, deviceDataSvc services.UserDeviceDataAPIService) {

	ingestSvc := services.NewVehicleMintValuationIngest(pdb.DBS, logger, settings, userDeviceSvc, ddSvc, deviceDataSvc)
	//goka setup
	sc := goka.DefaultConfig()
	sc.Version = sarama.V2_8_1_0
	goka.ReplaceGlobalConfig(sc)

	group := goka.DefineGroup("valuation-trigger-consumer",
		goka.Input(goka.Stream(settings.EventsTopic), new(shared.JSONCodec[shared.CloudEvent[json.RawMessage]]), ingestSvc.ProcessVehicleMintMsg),
	)

	processor, err := goka.NewProcessor(strings.Split(settings.KafkaBrokers, ","),
		group,
		goka.WithHasher(kafkautil.MurmurHasher),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start valuations trigger processor")
	}

	go func() {
		err = processor.Run(context.Background())
		if err != nil {
			logger.Fatal().Err(err).Msg("could not run device status processor")
		}
	}()

	logger.Info().Msg("valuations trigger from vehicle mint consumer started")
}

// startMonitoringServer start server for monitoring endpoints. Could likely be moved to shared lib.
func startMonitoringServer(logger zerolog.Logger, settings *config.Settings) {
	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})
	monApp.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("healthy")
	})

	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	go func() {
		// 8888 is our standard port for exposing metrics in DIMO infra
		if err := monApp.Listen(":" + settings.MonitoringPort); err != nil {
			logger.Fatal().Err(err).Str("port", settings.MonitoringPort).Msg("Failed to start monitoring web server.")
		}
	}()

	logger.Info().Str("port", "8888").Msg("Started monitoring web server.")
}

func startGRCPServer(pdb db.Store, logger zerolog.Logger, settings *config.Settings, userDeviceSvc services.UserDeviceAPIService) {
	lis, err := net.Listen("tcp", ":"+settings.GRPCPort)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't listen on gRPC port %s", settings.GRPCPort)
	}

	logger.Info().Msgf("Starting gRPC server on port %s", settings.GRPCPort)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			metrics.GRPCMetricsAndLogMiddleware(&logger),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		)),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)
	pb.RegisterValuationsServiceServer(server, rpc.NewValuationsService(pdb.DBS, &logger, userDeviceSvc))

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server terminated unexpectedly")
	}
}

func startWebAPI(logger zerolog.Logger, settings *config.Settings, userDeviceSvc services.UserDeviceAPIService,
	drivlySvc services.DrivlyValuationService, vincarioSvc services.VincarioValuationService) *fiber.App {

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return helpers.ErrorHandler(c, err, &logger, settings.IsProduction())
		},
		DisableStartupMessage: true,
		ReadBufferSize:        16000,
	})

	app.Use(metrics.HTTPMetricsMiddleware)

	app.Use(fiberrecover.New(fiberrecover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: nil,
	}))
	app.Use(cors.New())

	app.Get("/", healthCheck)
	app.Get("/v1/swagger/*", swagger.HandlerDefault)

	vehiclesController := controllers.NewVehiclesController(&logger, userDeviceSvc, drivlySvc, vincarioSvc)

	// secured paths
	privilegeAuth := jwtware.New(jwtware.Config{
		JWKSetURLs: []string{settings.TokenExchangeJWTKeySetURL},
		ErrorHandler: func(_ *fiber.Ctx, _ error) error {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid privilege token.")
		},
	})
	tk := privilegetoken.New(privilegetoken.Config{
		Log: &logger,
	})
	vehicleAddr := common.HexToAddress(settings.VehicleNFTAddress)

	vOwner := app.Group("/v2/vehicles/:tokenId", privilegeAuth)
	vOwner.Get("/valuations", tk.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleNonLocationData}), vehiclesController.GetValuations)
	vOwner.Get("/offers", tk.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleNonLocationData}), vehiclesController.GetOffers)
	// request an offer of valuation
	vOwner.Post("/instant-offer", tk.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleNonLocationData}), vehiclesController.RequestInstantOffer)
	vOwner.Post("/valuation", tk.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleNonLocationData}), vehiclesController.RequestValuationOnly)
	// same as above but it causes confusion so
	vOwner.Post("/valuations", tk.OneOf(vehicleAddr, []privileges.Privilege{privileges.VehicleNonLocationData}), vehiclesController.RequestValuationOnly)

	logger.Info().Msg("HTTP web server started on port " + settings.Port)
	// Start Server from a different go routine
	go func() {
		if err := app.Listen(":" + settings.Port); err != nil {
			logger.Fatal().Err(err).Send()
		}
	}()
	return app
}

type CodeResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(CodeResp{Code: 200, Message: "Server is up."})
}
