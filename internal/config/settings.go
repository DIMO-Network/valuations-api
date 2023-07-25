package config

import (
	"github.com/DIMO-Network/shared/db"
)

type Settings struct {
	Environment               string      `yaml:"ENVIRONMENT"`
	LogLevel                  string      `yaml:"LOG_LEVEL"`
	Port                      string      `yaml:"PORT"`
	GRPCPort                  string      `yaml:"GRPC_PORT"`
	MonitoringPort            string      `yaml:"MONITORING_PORT"`
	DB                        db.Settings `yaml:"DB"`
	ServiceName               string      `yaml:"SERVICE_NAME"`
	ServiceVersion            string      `yaml:"SERVICE_VERSION"`
	KafkaBrokers              string      `yaml:"KAFKA_BROKERS"`
	ValuationRequestTopic     string      `yaml:"VALUATION_REQUEST_TOPIC"`
	DevicesGRPCAddr           string      `yaml:"DEVICES_GRPC_ADDR"`
	DeviceDataGRPCAddr        string      `yaml:"DEVICE_DATA_GRPC_ADDR"`
	DeviceDefinitionsGRPCAddr string      `yaml:"DEVICE_DEFINITIONS_GRPC_ADDR"`
	VincarioAPIURL            string      `yaml:"VINCARIO_API_URL"`
	VincarioAPISecret         string      `yaml:"VINCARIO_API_SECRET"`
	VincarioAPIKey            string      `yaml:"VINCARIO_API_KEY"`
	DrivlyAPIKey              string      `yaml:"DRIVLY_API_KEY"`
	DrivlyVINAPIURL           string      `yaml:"DRIVLY_VIN_API_URL"`
	DrivlyOfferAPIURL         string      `yaml:"DRIVLY_OFFER_API_URL"`
	GoogleMapsAPIKey          string      `yaml:"GOOGLE_MAPS_API_KEY"`
}
