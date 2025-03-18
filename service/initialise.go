package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	filterCli "github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpkafka "github.com/ONSdigital/dp-kafka/v4"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	HealthCheck      bool
	KafkaConsumer    bool
	KafkaProducer    bool
	DatasetAPI       bool
	FilterStore      bool
	ObservationStore bool
	FileStore        bool
	Init             Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		HealthCheck:      false,
		KafkaConsumer:    false,
		KafkaProducer:    false,
		DatasetAPI:       false,
		FilterStore:      false,
		ObservationStore: false,
		FileStore:        false,
		Init:             initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetObservationStore returns an initialised connection to observation store (graph database)
func (e *ExternalServiceList) GetObservationStore(ctx context.Context) (observationStore *graph.DB, err error) {
	o, err := e.Init.DoGetObservationStore(ctx)
	if err != nil {
		return nil, err
	}
	e.ObservationStore = true

	return o, nil
}

// GetDatasetAPIClient gets and initialises the DatasetAPI Client
func (e *ExternalServiceList) GetDatasetAPIClient(cfg *config.Config) (DatasetAPI, error) {
	d := e.Init.DoGetDatasetAPIClient(cfg)

	e.DatasetAPI = true
	return d, nil
}

// GetFileStore returns an initialised connection to file store
func (e *ExternalServiceList) GetFileStore(ctx context.Context, cfg *config.Config) (fileStore *file.Store, err error) {

	f, err := e.Init.DoGetFileStore(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.FileStore = true

	return f, nil
}

// GetFilterStore returns a filter client
func (e *ExternalServiceList) GetFilterStore(cfg *config.Config, serviceAuthToken string) (FilterStore, error) {
	d := e.Init.DoGetFilterStore(cfg, serviceAuthToken)
	e.FilterStore = true
	return d, nil
}

// GetHTTPServer creates an http server and sets the Server flag to true
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetKafkaConsumer creates a Kafka consumer and sets the consumer flag to true
func (e *ExternalServiceList) GetKafkaConsumer(ctx context.Context, cfg *config.Config) (dpkafka.IConsumerGroup, error) {
	consumer, err := e.Init.DoGetKafkaConsumer(ctx, &cfg.KafkaConfig)
	if err != nil {
		return nil, err
	}
	e.KafkaConsumer = true
	return consumer, nil
}

// GetKafkaConsumer creates a Kafka producer and sets the producer flag to true
func (e *ExternalServiceList) GetKafkaProducer(ctx context.Context, cfg *config.Config) (dpkafka.IProducer, error) {
	producer, err := e.Init.DoGetKafkaProducer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaProducer = true
	return producer, nil
}

// GetHealthCheck creates a healthcheck with versionInfo and sets the HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// DoGetDatasetAPIClient creates a datasetAPI client
func (e *Init) DoGetDatasetAPIClient(cfg *config.Config) DatasetAPI {
	d := dataset.NewAPIClient(cfg.DatasetAPIURL)
	return d
}

// DoGetFileStore creates a connection to s3
func (e *Init) DoGetFileStore(ctx context.Context, cfg *config.Config) (fileStore *file.Store, err error) {

	fileStore, err = file.NewStore(
		ctx,
		cfg.AWSRegion,
		cfg.S3BucketURL,
		cfg.S3BucketName,
		cfg.S3PrivateBucketName,
		cfg.LocalstackHost,
	)
	if err != nil {
		return
	}
	return fileStore, nil
}

// DoGetFilterStore creates a connection to the Filter data store
func (e *Init) DoGetFilterStore(cfg *config.Config, serviceAuthToken string) FilterStore {
	filterAPIClient := filterCli.New(cfg.FilterAPIURL)
	f := filter.NewStore(filterAPIClient, serviceAuthToken)
	return f
}

// DoGetObservationStore creates a graph DB client
func (e *Init) DoGetObservationStore(ctx context.Context) (observationStore *graph.DB, err error) {
	o, err := graph.New(ctx, graph.Subsets{Observation: true})
	if err != nil {
		return
	}
	return o, nil
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetKafkaConsumer returns a Kafka Consumer group
func (e *Init) DoGetKafkaConsumer(ctx context.Context, kafkaCfg *config.KafkaConfig) (dpkafka.IConsumerGroup, error) {
	kafkaOffset := dpkafka.OffsetNewest
	if kafkaCfg.OffsetOldest {
		kafkaOffset = dpkafka.OffsetOldest
	}
	cgConfig := &dpkafka.ConsumerGroupConfig{
		BrokerAddrs:  kafkaCfg.Brokers,
		KafkaVersion: &kafkaCfg.Version,
		Offset:       &kafkaOffset,
		Topic:        kafkaCfg.FilterConsumerTopic,
		GroupName:    kafkaCfg.FilterConsumerGroup,
	}
	if kafkaCfg.SecProtocol == config.KafkaTLSProtocolFlag {
		cgConfig.SecurityConfig = dpkafka.GetSecurityConfig(
			kafkaCfg.SecCACerts,
			kafkaCfg.SecClientCert,
			kafkaCfg.SecClientKey,
			kafkaCfg.SecSkipVerify,
		)
	}
	kafkaConsumer, err := dpkafka.NewConsumerGroup(
		ctx,
		cgConfig,
	)
	if err != nil {
		return nil, err
	}

	return kafkaConsumer, nil
}

// DoGetKafkaProducer creates a kafka producer
func (e *Init) DoGetKafkaProducer(ctx context.Context, cfg *config.Config) (dpkafka.IProducer, error) {
	pConfig := &dpkafka.ProducerConfig{
		BrokerAddrs:       cfg.KafkaConfig.Brokers,
		Topic:             cfg.KafkaConfig.CSVExportedProducerTopic,
		KafkaVersion:      &cfg.KafkaConfig.Version,
		MaxMessageBytes:   &cfg.KafkaConfig.MaxBytes,
		MinBrokersHealthy: &cfg.KafkaConfig.ProducerMinBrokersHealthy,
	}

	if cfg.KafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
		pConfig.SecurityConfig = dpkafka.GetSecurityConfig(
			cfg.KafkaConfig.SecCACerts,
			cfg.KafkaConfig.SecClientCert,
			cfg.KafkaConfig.SecClientKey,
			cfg.KafkaConfig.SecSkipVerify,
		)
	}

	return dpkafka.NewProducer(ctx, pConfig)
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}
