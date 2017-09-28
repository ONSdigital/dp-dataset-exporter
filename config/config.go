package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config values for the application.
type Config struct {
	BindAddr                 string   `envconfig:"BIND_ADDR"`
	KafkaAddr                []string `envconfig:"KAFKA_ADDR"`
	FilterJobConsumerGroup   string   `envconfig:"FILTER_JOB_CONSUMER_GROUP"`
	FilterJobConsumerTopic   string   `envconfig:"FILTER_JOB_CONSUMER_TOPIC"`
	DatabaseAddress          string   `envconfig:"DATABASE_ADDRESS"`
	FilterAPIURL             string   `envconfig:"FILTER_API_URL"`
	FilterAPIAuthToken       string   `envconfig:"FILTER_API_AUTH_TOKEN"`
	AWSRegion                string   `envconfig:"AWS_REGION"`
	S3BucketName             string   `envconfig:"S3_BUCKET_NAME"`
	CSVExportedProducerTopic string   `envconfig:"CSV_EXPORTED_PRODUCER_TOPIC"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := &Config{
		BindAddr:                 ":22500",
		KafkaAddr:                []string{"localhost:9092"},
		FilterJobConsumerTopic:   "filter-job-submitted",
		FilterJobConsumerGroup:   "dp-dataset-exporter",
		DatabaseAddress:          "bolt://localhost:7687",
		FilterAPIURL:             "http://localhost:22100",
		FilterAPIAuthToken:       "FD0108EA-825D-411C-9B1D-41EF7727F465",
		CSVExportedProducerTopic: "common-output-created",
		S3BucketName:             "csv-exported",
		AWSRegion:                "eu-west-1",
	}

	return cfg, envconfig.Process("", cfg)
}
