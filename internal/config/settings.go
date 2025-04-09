package config

import (
	"github.com/DIMO-Network/shared/db"
	"net/url"
)

type Settings struct {
	Environment               string      `yaml:"ENVIRONMENT"`
	LogLevel                  string      `yaml:"LOG_LEVEL"`
	Port                      string      `yaml:"PORT"`
	GRPCPort                  string      `yaml:"GRPC_PORT"`
	MonitoringPort            string      `yaml:"MONITORING_PORT"`
	DB                        db.Settings `yaml:"DB"`
	JwtKeySetURL              string      `yaml:"JWT_KEY_SET_URL"`
	ServiceName               string      `yaml:"SERVICE_NAME"`
	ServiceVersion            string      `yaml:"SERVICE_VERSION"`
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

	NATSURL                      string `yaml:"NATS_URL"`
	NATSStreamName               string `yaml:"NATS_STREAM_NAME"`
	NATSValuationSubject         string `yaml:"NATS_VALUATION_SUBJECT"`
	NATSAckTimeout               string `yaml:"NATS_ACK_TIMEOUT"`
	NATSValuationDurableConsumer string `yaml:"NATS_VALUATION_DURABLE_CONSUMER"`
	NATSOfferSubject             string `yaml:"NATS_OFFER_SUBJECT"`
	NATSOfferDurableConsumer     string `yaml:"NATS_OFFER_DURABLE_CONSUMER"`
	UsersGRPCAddr                string `yaml:"USERS_GRPC_ADDR"`
	VehicleNFTAddress            string `yaml:"VEHICLE_NFT_ADDRESS"`
	TokenExchangeJWTKeySetURL    string `yaml:"TOKEN_EXCHANGE_JWT_KEY_SET_URL"`

	// EventsTopic kafka topic to get onchain events emmitted by devices-api
	EventsTopic  string `yaml:"EVENTS_TOPIC"`
	KafkaBrokers string `yaml:"KAFKA_BROKERS"`

	IdentityAPIURL url.URL `yaml:"IDENTITY_API_URL"`
}

func (s *Settings) IsProduction() bool {
	return s.Environment == "prod"
}
