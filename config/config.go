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
		CSVExportedProducerTopic: "csv-exported",
	}

	err := gofigure.Gofigure(&cfg)

	return &cfg, err
}
