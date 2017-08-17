package config

import (
	"github.com/ian-kent/gofigure"
)

// Config values for the application.
type Config struct {
	BindAddr                 string `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The port to bind to"`
	KafkaAddr                string `env:"KAFKA_ADDR" flag:"kafka-addr" flagDesc:"The address of the Kafka instance"`
	FilterJobConsumerGroup   string `env:"FILTER_JOB_CONSUMER_GROUP" flag:"filter-job-consumer-group" flagDesc:"The Kafka consumer group to consume filter job messages from"`
	FilterJobConsumerTopic   string `env:"FILTER_JOB_CONSUMER_TOPIC" flag:"filter-job-consumer-topic" flagDesc:"The Kafka topic to consume filter job messages from"`
	DatabaseAddress          string `env:"DATABASE_ADDRESS" flag:"database-address" flagDesc:"The address of the database to store observations"`
	FilterAPIURL             string `env:"FILTER_API_URL" flag:"filter-api-url" flagDesc:"The URL of the filter API"`
	FilterAPIAuthToken       string `env:"FILTER_API_AUTH_TOKEN" flag:"filter-api-auth-token" flagDesc:"Authentication token for access to filter API"`
	AWSRegion                string `env:"AWS_REGION" flag:"aws-region" flagDesc:"The AWS region to use"`
	S3BucketName             string `env:"S3_BUCKET_NAME" flag:"s3-bucket-name" flagDesc:"The S3 bucket name to store exported files"`
	CSVExportedProducerTopic string `env:"CSV_EXPORTED_PRODUCER_TOPIC" flag:"csv-exported-producer-topic" flagDesc:"The Kafka topic to send result messages to"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := Config{
		BindAddr:                 ":22500",
		KafkaAddr:                "localhost:9092",
		FilterJobConsumerTopic:   "filter-job-submitted",
		FilterJobConsumerGroup:   "dp-dataset-exporter",
		DatabaseAddress:          "bolt://localhost:7687",
		FilterAPIURL:             "http://localhost:22100",
		FilterAPIAuthToken:       "FD0108EA-825D-411C-9B1D-41EF7727F465",
		CSVExportedProducerTopic: "csv-exported",
		S3BucketName:             "csv-exported",
		AWSRegion:                "eu-west-1",
	}

	return &cfg, gofigure.Gofigure(&cfg)
}
