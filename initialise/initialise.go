package initialise

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-graph/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	vault "github.com/ONSdigital/dp-vault"
)

// ExternalServiceList represents a list of services
type ExternalServiceList struct {
	Consumer            bool
	CSVExportedProducer bool
	ErrorProducer       bool
	EventConsumer       bool
	FileStore           bool
	ObservationStore    bool
	HealthCheck         bool
	Vault               bool
}

// KafkaProducerName represents a type for kafka producer name used by iota constants
type KafkaProducerName int

// Possible names of Kafka Producers
const (
	CSVExported = iota
	Error
)

var kafkaProducerNames = []string{"CSVExported", "Error"}

// Values of the kafka producers names
func (k KafkaProducerName) String() string {
	return kafkaProducerNames[k]
}

// GetConsumer returns a kafka consumer, which might not be initialised
func (e *ExternalServiceList) GetConsumer(cfg *config.Config) (kafkaConsumer *kafka.ConsumerGroup, err error) {
	ctx := context.Background()

	k, err := kafka.NewConsumerGroup(
		ctx,
		cfg.KafkaAddr,
		cfg.FilterConsumerTopic,
		cfg.FilterConsumerGroup,
		kafka.OffsetNewest,
		true,
		kafka.CreateConsumerGroupChannels(true),
	)
	if err != nil {
		return
	}

	e.Consumer = true

	return &k, err
}

// GetFileStore returns an initialised connection to file store
func (e *ExternalServiceList) GetFileStore(cfg *config.Config, vaultClient *vault.Client) (fileStore *file.Store, err error) {
	fileStore, err = file.NewStore(
		cfg.AWSRegion,
		cfg.S3BucketURL,
		cfg.S3BucketName,
		cfg.S3PrivateBucketName,
		cfg.VaultPath,
		vaultClient,
	)
	if err != nil {
		return
	}

	e.FileStore = true

	return
}

// GetObservationStore returns an initialised connection to observation store (graph database)
func (e *ExternalServiceList) GetObservationStore() (observationStore *graph.DB, err error) {
	observationStore, err = graph.New(context.Background(), graph.Subsets{Observation: true})
	if err != nil {
		return
	}

	e.ObservationStore = true

	return
}

// GetProducer returns a kafka producer, which might no be initialised
func (e *ExternalServiceList) GetProducer(kafkaBrokers []string, topic string, name KafkaProducerName) (kafkaProducer *kafka.Producer, err error) {

	ctx := context.Background()

	k, err := kafka.NewProducer(
		ctx,
		kafkaBrokers,
		topic,
		0,
		kafka.CreateProducerChannels(),
	)
	if err != nil {
		return
	}

	switch {
	case name == CSVExported:
		e.CSVExportedProducer = true
	case name == Error:
		e.ErrorProducer = true
	default:
		err = fmt.Errorf("Kafka producer name not recognised: '%s'. Valid names: %v", name.String(), kafkaProducerNames)
	}

	return &k, err
}

// GetVault returns a vault client
func (e *ExternalServiceList) GetVault(cfg *config.Config, retries int) (client *vault.Client, err error) {
	client, err = vault.CreateClient(cfg.VaultToken, cfg.VaultAddress, retries)
	if err != nil {
		return
	}

	e.Vault = true

	return
}

// GetHealthCheck creates a healthcheck with versionInfo
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (healthcheck.HealthCheck, error) {

	// Create healthcheck object with versionInfo
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return healthcheck.HealthCheck{}, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckRecoveryInterval, cfg.HealthCheckInterval)

	e.HealthCheck = true

	return hc, nil
}
