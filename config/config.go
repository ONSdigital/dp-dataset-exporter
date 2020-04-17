package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config values for the application.
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	KafkaAddr                  []string      `envconfig:"KAFKA_ADDR"`
	FilterConsumerGroup        string        `envconfig:"FILTER_JOB_CONSUMER_GROUP"`
	FilterConsumerTopic        string        `envconfig:"FILTER_JOB_CONSUMER_TOPIC"`
	FilterAPIURL               string        `envconfig:"FILTER_API_URL"`
	AWSRegion                  string        `envconfig:"AWS_REGION"`
	S3BucketName               string        `envconfig:"S3_BUCKET_NAME"`
	S3BucketURL                string        `envconfig:"S3_BUCKET_URL"`
	S3PrivateBucketName        string        `envconfig:"S3_PRIVATE_BUCKET_NAME"`
	CSVExportedProducerTopic   string        `envconfig:"CSV_EXPORTED_PRODUCER_TOPIC"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	ErrorProducerTopic         string        `envconfig:"ERROR_PRODUCER_TOPIC"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	VaultToken                 string        `envconfig:"VAULT_TOKEN"                   json:"-"`
	VaultAddress               string        `envconfig:"VAULT_ADDR"`
	VaultPath                  string        `envconfig:"VAULT_PATH"`
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	APIDomainURL               string        `envconfig:"API_DOMAIN_URL"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"            json:"-"`
	StartupTimeout             time.Duration `envconfig:"STARTUP_TIMEOUT"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	FullDatasetFilePrefix      string        `envconfig:"FULL_DATASET_FILE_PREFIX"`
	FilteredDatasetFilePrefix  string        `envconfig:"FILTERED_DATASET_FILE_PREFIX"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := &Config{
		BindAddr:                   ":22500",
		KafkaAddr:                  []string{"localhost:9092"},
		FilterConsumerTopic:        "filter-job-submitted",
		FilterConsumerGroup:        "dp-dataset-exporter",
		FilterAPIURL:               "http://localhost:22100",
		CSVExportedProducerTopic:   "common-output-created",
		S3BucketName:               "csv-exported",
		S3BucketURL:                "",
		S3PrivateBucketName:        "csv-exported",
		AWSRegion:                  "eu-west-1",
		DatasetAPIURL:              "http://localhost:22000",
		ErrorProducerTopic:         "filter-error",
		GracefulShutdownTimeout:    time.Second * 10,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		VaultPath:                  "secret/shared/psk",
		VaultAddress:               "http://localhost:8200",
		VaultToken:                 "",
		DownloadServiceURL:         "http://localhost:23600",
		APIDomainURL:               "http://localhost:23200",
		ServiceAuthToken:           "0f49d57b-c551-4d33-af1e-a442801dd851",
		StartupTimeout:             time.Second * 125,
		ZebedeeURL:                 "http://localhost:8082",
		FullDatasetFilePrefix:      "full-datasets/",
		FilteredDatasetFilePrefix:  "filtered-datasets/",
	}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
