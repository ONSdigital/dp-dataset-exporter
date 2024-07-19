package event

import (
	"context"
	"io"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-graph/v2/observation"
	kafka "github.com/ONSdigital/dp-kafka/v4"
)

// FilterStore provides existing filter data.
type FilterStore interface {
	GetFilter(ctx context.Context, filterID string) (*filter.Model, error)
	PutCSVData(ctx context.Context, filterID string, downloadItem filter.Download) error
	PutStateAsEmpty(ctx context.Context, filterJobID string) error
	PutStateAsError(ctx context.Context, filterJobID string) error
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
}

// ObservationStore provides filtered observation data in CSV rows.
type ObservationStore interface {
	StreamCSVRows(ctx context.Context, instanceID, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error)
}

// FileStore provides storage for filtered output files.
type FileStore interface {
	PutFile(ctx context.Context, reader io.Reader, filename string, isPublished bool) (url string, err error)
}

// Producer handles producing output events.
type Producer interface {
	CSVExported(ctx context.Context, e *CSVExported) error
}

type KafkaProducer interface {
	Output() chan kafka.BytesMessage
}

type GenerateOutputCreatedEvent interface {
	Marshal(s interface{}) ([]byte, error)
}
