package api

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/DIMO-Network/valuations-api/internal/controllers"
	"github.com/DIMO-Network/valuations-api/internal/controllers/helpers"
	"github.com/DIMO-Network/valuations-api/internal/core/commands"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/metrics"
	"github.com/DIMO-Network/valuations-api/internal/rpc"
	pb "github.com/DIMO-Network/valuations-api/pkg/grpc"
	"github.com/gofiber/adaptor/v2"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

func Run(ctx context.Context, pdb db.Store, logger zerolog.Logger, settings *config.Settings,
	ddSvc services.DeviceDefinitionsAPIService, userDeviceSvc services.UserDeviceAPIService, deviceDataSvc services.UserDeviceDataAPIService,
	natsSvc *services.NATSService) {

	handler := commands.NewRunValuationCommandHandler(pdb.DBS, logger, settings, userDeviceSvc, ddSvc, deviceDataSvc, natsSvc)

	go func() {
		err := handler.Execute(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("unable to start nats consumer")
		}
	}()

	startMonitoringServer(logger, settings)
	go startGRCPServer(pdb, logger, settings, userDeviceSvc)
	app := startWebAPI(logger, settings, pdb, userDeviceSvc)
	// nolint
	defer app.Shutdown()

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent with length of 1
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
	<-c                                             // This blocks the main thread until an interrupt is received
	logger.Info().Msg("Gracefully shutting down and running cleanup tasks...")
	_ = ctx.Done()
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

// @title                       DIMO Vehicle Valuations API
// @description 				API to get latest valuation for a given connected vehicle belonging to user
// @version                     1.0
// @BasePath                    /v1
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb db.Store, userDeviceSvc services.UserDeviceAPIService) *fiber.App {
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

	valuationsController := controllers.NewValuationsController(&logger, pdb.DBS, userDeviceSvc)

	// secured paths
	jwtAuth := jwtware.New(jwtware.Config{
		JWKSetURLs: []string{settings.JwtKeySetURL},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid JWT.")
		},
	})

	v1Auth := app.Group("/v1", jwtAuth)
	// todo bring in udOwner stuff, but see if can put in shared - major refactor btw
	v1Auth.Get("/user/devices/:userDeviceID/valuations", valuationsController.GetValuations)
	v1Auth.Get("/user/devices/:userDeviceID/offers", valuationsController.GetOffers)

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
