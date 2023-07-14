package api

import (
	"context"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/valuations-api/internal/core/services"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/DIMO-Network/valuations-api/internal/core/commands"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/kafka"
	"github.com/Shopify/sarama"

	"github.com/DIMO-Network/valuations-api/internal/config"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

func Run(ctx context.Context, pdb db.Store, logger zerolog.Logger, settings *config.Settings, ddSvc services.DeviceDefinitionsAPIService, userDeviceSvc services.UserDeviceAPIService, deviceDataSvc services.UserDeviceDataAPIService) {

	startMonitoringServer(logger, settings)
	startValuationConsumer(pdb, logger, settings, ddSvc, userDeviceSvc, deviceDataSvc)

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

func startValuationConsumer(pdb db.Store, logger zerolog.Logger, settings *config.Settings,
	ddSvc services.DeviceDefinitionsAPIService,
	userDeviceSvc services.UserDeviceAPIService,
	deviceDataSvc services.UserDeviceDataAPIService) {

	if len(settings.KafkaBrokers) == 0 {
		return
	}

	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_8_1_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.DBCDecodingTopic,
		GroupID:         "valuations-api",
		MaxInFlight:     int64(5),
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start credential update consumer")
	}

	natSvc, _ := services.NewNATSService(settings, &logger)
	handler := commands.NewRunValuationCommandHandler(pdb.DBS, logger, settings, userDeviceSvc, ddSvc, deviceDataSvc, *natSvc)
	service := NewWorkerListenerService(logger, handler)

	consumer.Start(context.Background(), service.ProcessWorker)

	logger.Info().Msg("Vehicle Signal Decoding consumer started")
}
