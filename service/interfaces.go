package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	vault "github.com/ONSdigital/dp-vault"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock . HealthChecker

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetDatasetAPIClient(cfg *config.Config) DatasetAPI
	DoGetFilterStore(cfg *config.Config, serviceAuthToken string) FilterStore
	DoGetObservationStore(ctx context.Context) (observationStore *graph.DB, err error)
	DoGetFileStore(cfg *config.Config, vaultClient *vault.Client) (fileStore *file.Store, err error)
	DoGetVault(cfg *config.Config, retries int) (client *vault.Client, err error)
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetKafkaConsumer(ctx context.Context, kafkaCfg *config.KafkaConfig) (kafka.IConsumerGroup, error)
	DoGetKafkaProducer(ctx context.Context, cfg *config.Config) (kafka.IProducer, error)
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

// EventConsumer defines the required methods from event Consumer
type EventConsumer interface {
	Close(ctx context.Context) (err error)
}

// FilterStore provides existing filter data.
type FilterStore interface {
	GetFilter(ctx context.Context, filterID string) (*filter.Model, error)
	PutCSVData(ctx context.Context, filterID string, downloadItem filter.Download) error
	PutStateAsEmpty(ctx context.Context, filterJobID string) error
	PutStateAsError(ctx context.Context, filterJobID string) error
	Checker(context.Context, *healthcheck.CheckState) error
}

// DatasetAPI contains functions to call the dataset API.
type DatasetAPI interface {
	PutVersion(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version string, m dataset.Version) error
	GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error)
	GetInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID, ifMatch string) (m dataset.Instance, eTag string, err error)
	GetMetadataURL(id, edition, version string) (url string)
	GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.Metadata, err error)
	GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, q *dataset.QueryParams) (m dataset.Options, err error)
	GetVersionDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.VersionDimensions, err error)
	GetOptionsInBatches(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, batchSize, maxWorkers int) (m dataset.Options, err error)
	Checker(context.Context, *healthcheck.CheckState) error
}
