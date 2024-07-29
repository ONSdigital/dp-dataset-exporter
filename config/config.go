package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const KafkaTLSProtocolFlag = "TLS"

// Config values for the application.
type Config struct {
	BindAddr                   string `envconfig:"BIND_ADDR"`
	KafkaConfig                KafkaConfig
	FilterAPIURL               string        `envconfig:"FILTER_API_URL"`
	AWSRegion                  string        `envconfig:"AWS_REGION"`
	S3BucketName               string        `envconfig:"S3_BUCKET_NAME"`
	S3BucketURL                string        `envconfig:"S3_BUCKET_URL"`
	S3PrivateBucketName        string        `envconfig:"S3_PRIVATE_BUCKET_NAME"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	ErrorProducerTopic         string        `envconfig:"ERROR_PRODUCER_TOPIC"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	APIDomainURL               string        `envconfig:"API_DOMAIN_URL"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"            json:"-"`
	StartupTimeout             time.Duration `envconfig:"STARTUP_TIMEOUT"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	FullDatasetFilePrefix      string        `envconfig:"FULL_DATASET_FILE_PREFIX"`
	FilteredDatasetFilePrefix  string        `envconfig:"FILTERED_DATASET_FILE_PREFIX"`
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OTBatchTimeout             time.Duration `envconfig:"OTEL_BATCH_TIMEOUT"`
	OtelEnabled                bool          `envconfig:"OTEL_ENABLED"`
	LocalstackHost             string        `envconfig:"LOCALSTACK_HOST"`
}

type KafkaConfig struct {
	Brokers                   []string `envconfig:"KAFKA_ADDR"`
	Version                   string   `envconfig:"KAFKA_VERSION"`
	OffsetOldest              bool     `envconfig:"KAFKA_OFFSET_OLDEST"`
	SecProtocol               string   `envconfig:"KAFKA_SEC_PROTO"`
	SecCACerts                string   `envconfig:"KAFKA_SEC_CA_CERTS"`
	SecClientKey              string   `envconfig:"KAFKA_SEC_CLIENT_KEY"    json:"-"`
	SecClientCert             string   `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	SecSkipVerify             bool     `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	NumWorkers                int      `envconfig:"KAFKA_NUM_WORKERS"`
	FilterConsumerGroup       string   `envconfig:"HELLO_CALLED_GROUP"`
	FilterConsumerTopic       string   `envconfig:"HELLO_CALLED_TOPIC"`
	CSVExportedProducerTopic  string   `envconfig:"CSV_EXPORTED_PRODUCER_TOPIC"`
	ConsumerMinBrokersHealthy int      `envconfig:"KAFKA_CONSUMER_MIN_BROKERS_HEALTHY"`
	MaxBytes                  int      `envconfig:"KAFKA_MAX_BYTES"`
	ProducerMinBrokersHealthy int      `envconfig:"KAFKA_PRODUCER_MIN_BROKERS_HEALTHY"`
}

// Get the configuration values from the environment or provide the defaults.
func Get() (*Config, error) {

	cfg := &Config{
		BindAddr:                   ":22500",
		FilterAPIURL:               "http://localhost:22100",
		S3BucketName:               "csv-exported",
		S3BucketURL:                "",
		S3PrivateBucketName:        "csv-exported",
		AWSRegion:                  "eu-west-1",
		DatasetAPIURL:              "http://localhost:22000",
		ErrorProducerTopic:         "filter-error",
		GracefulShutdownTimeout:    time.Second * 10,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		DownloadServiceURL:         "http://localhost:23600",
		APIDomainURL:               "http://localhost:23200",
		ServiceAuthToken:           "0f49d57b-c551-4d33-af1e-a442801dd851",
		StartupTimeout:             time.Second * 125,
		ZebedeeURL:                 "http://localhost:8082",
		FullDatasetFilePrefix:      "full-datasets/",
		FilteredDatasetFilePrefix:  "filtered-datasets/",
		OTExporterOTLPEndpoint:     "localhost:4317",
		OTServiceName:              "dp-dataset-exporter",
		OTBatchTimeout:             5 * time.Second,
		OtelEnabled:                false,
		KafkaConfig: KafkaConfig{
			Brokers:                   []string{"localhost:9092", "localhost:9093", "localhost:9094"},
			Version:                   "1.0.2",
			OffsetOldest:              true,
			NumWorkers:                1,
			FilterConsumerTopic:       "filter-job-submitted",
			FilterConsumerGroup:       "dp-dataset-exporter",
			CSVExportedProducerTopic:  "common-output-created",
			ConsumerMinBrokersHealthy: 1,
			MaxBytes:                  2000000,
			ProducerMinBrokersHealthy: 2,
			SecCACerts:                "",
			SecClientCert:             "",
			SecClientKey:              "",
			SecProtocol:               "",
			SecSkipVerify:             false,
		},
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
	jsonStr, _ := json.Marshal(config)
	return string(jsonStr)
}
