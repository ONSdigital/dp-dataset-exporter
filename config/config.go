package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config values for the application.
type Config struct {
	BindAddr                 string        `envconfig:"BIND_ADDR"`
	KafkaAddr                []string      `envconfig:"KAFKA_ADDR"`
	FilterConsumerGroup      string        `envconfig:"FILTER_JOB_CONSUMER_GROUP"`
	FilterConsumerTopic      string        `envconfig:"FILTER_JOB_CONSUMER_TOPIC"`
	DatabaseAddress          string        `envconfig:"DATABASE_ADDRESS"`
	FilterAPIURL             string        `envconfig:"FILTER_API_URL"`
	FilterAPIAuthToken       string        `envconfig:"FILTER_API_AUTH_TOKEN"`
	AWSRegion                string        `envconfig:"AWS_REGION"`
	S3BucketName             string        `envconfig:"S3_BUCKET_NAME"`
	CSVExportedProducerTopic string        `envconfig:"CSV_EXPORTED_PRODUCER_TOPIC"`
	DatasetAPIURL            string        `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken      string        `envconfig:"DATASET_API_AUTH_TOKEN"`
	ErrorProducerTopic       string        `envconfig:"ERROR_PRODUCER_TOPIC"`
	GracefulShutdownTimeout  time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := &Config{
		BindAddr:                 ":22500",
		KafkaAddr:                []string{"localhost:9092"},
		FilterConsumerTopic:      "filter-job-submitted",
		FilterConsumerGroup:      "dp-dataset-exporter",
		DatabaseAddress:          "bolt://localhost:7687",
		FilterAPIURL:             "http://localhost:22100",
		FilterAPIAuthToken:       "FD0108EA-825D-411C-9B1D-41EF7727F465",
		CSVExportedProducerTopic: "common-output-created",
		S3BucketName:             "csv-exported-test",
		AWSRegion:                "eu-west-1",
		DatasetAPIURL:            "http://localhost:22000",
		DatasetAPIAuthToken:      "FD0108EA-825D-411C-9B1D-41EF7727F465",
		ErrorProducerTopic:       "filter-error",
		GracefulShutdownTimeout:  time.Second * 10,
	}

	return cfg, envconfig.Process("", cfg)
}
