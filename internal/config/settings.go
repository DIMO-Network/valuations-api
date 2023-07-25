package config

import (
	"github.com/DIMO-Network/shared/db"
)

type Settings struct {
	Environment               string      `yaml:"ENVIRONMENT"`
	Port                      string      `yaml:"PORT"`
	LogLevel                  string      `yaml:"LOG_LEVEL"`
	DB                        db.Settings `yaml:"DB"`
	ServiceName               string      `yaml:"SERVICE_NAME"`
	ServiceVersion            string      `yaml:"SERVICE_VERSION"`
	GRPCPort                  string      `yaml:"GRPC_PORT"`
	KafkaBrokers              string      `yaml:"KAFKA_BROKERS"`
	ValuationRequestTopic     string      `yaml:"TASK_STATUS_TOPIC"`
	MonitoringPort            string      `yaml:"MONITORING_PORT"`
	DeviceGRPCAddr            string      `yaml:"DEVICE_GRPC_ADDR"`
	DeviceDefinitionsGRPCAddr string      `yaml:"DEVICE_DEFINITIONS_GRPC_ADDR"`
	VincarioAPIURL            string      `yaml:"VINCARIO_API_URL"`
	VincarioAPISecret         string      `yaml:"VINCARIO_API_SECRET"`
	VincarioAPIKey            string      `yaml:"VINCARIO_API_KEY"`
	DrivlyAPIKey              string      `yaml:"DRIVLY_API_KEY"`
	DrivlyVINAPIURL           string      `yaml:"DRIVLY_VIN_API_URL"`
	DrivlyOfferAPIURL         string      `yaml:"DRIVLY_OFFER_API_URL"`
	GoogleMapsAPIKey          string      `yaml:"GOOGLE_MAPS_API_KEY"`
}
