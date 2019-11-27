package initialise

import (
	"context"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-graph/graph"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/vault"
)

// ExternalServiceList represents a list of services
type ExternalServiceList struct {
	Consumer            bool
	CSVExportedProducer bool
	ErrorProducer       bool
	FileStore           bool
	ObservationStore    bool
	HealthTicker        bool
	Vault               bool
}

const (
	// CSVExportedProducer represents a name for
	// the producer that writes to a csv exported topic
	CSVExportedProducer = "csv-exported-producer"
	// ErrorProducer represents a name for
	// the producer that writes to an error topic
	ErrorProducer = "error-producer"
)

// GetConsumer returns an initialised kafka consumer
func (e *ExternalServiceList) GetConsumer(kafkaBrokers []string, cfg *config.Config) (kafkaConsumer *kafka.ConsumerGroup, err error) {
	kafkaConsumer, err = kafka.NewSyncConsumer(
		kafkaBrokers,
		cfg.FilterConsumerTopic,
		cfg.FilterConsumerGroup,
		kafka.OffsetNewest,
	)

	if err == nil {
		e.Consumer = true
	}

	return
}

// GetFileStore returns an initialised connection to file store
func (e *ExternalServiceList) GetFileStore(cfg *config.Config, vaultClient *vault.VaultClient) (fileStore *file.Store, err error) {
	fileStore, err = file.NewStore(
		cfg.AWSRegion,
		cfg.S3BucketURL,
		cfg.S3BucketName,
		cfg.S3PrivateBucketName,
		cfg.VaultPath,
		vaultClient,
	)
	if err == nil {
		e.FileStore = true
	}

	return
}

// GetObservationStore returns an initialised connection to observation store (graph database)
func (e *ExternalServiceList) GetObservationStore() (observationStore *graph.DB, err error) {
	observationStore, err = graph.New(context.Background(), graph.Subsets{Observation: true})
	if err == nil {
		e.ObservationStore = true
	}

	return
}

// GetProducer returns a kafka producer
func (e *ExternalServiceList) GetProducer(kafkaBrokers []string, topic, name string) (kafkaProducer kafka.Producer, err error) {
	kafkaProducer, err = kafka.NewProducer(kafkaBrokers, topic, 0)
	if err == nil {
		switch {
		case name == CSVExportedProducer:
			e.CSVExportedProducer = true
		case name == ErrorProducer:
			e.ErrorProducer = true
		}
	}

	return
}

// GetVault returns a vault client
func (e *ExternalServiceList) GetVault(cfg *config.Config, retries int) (client *vault.VaultClient, err error) {
	client, err = vault.CreateVaultClient(cfg.VaultToken, cfg.VaultAddress, retries)
	if err == nil {
		e.Vault = true
	}

	return
}
