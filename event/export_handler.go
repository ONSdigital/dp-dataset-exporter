package event

import (
	"github.com/ONSdigital/dp-dataset-exporter/observation"
	"io"
)

//go:generate moq -out eventtest/filter_store.go -pkg eventtest . FilterStore
//go:generate moq -out eventtest/observation_store.go -pkg eventtest . ObservationStore
//go:generate moq -out eventtest/file_store.go -pkg eventtest . FileStore

var _ Handler = (*ExportHandler)(nil)

// ExportHandler handles a single CSV export of a filtered dataset.
type ExportHandler struct {
	filterStore      FilterStore
	observationStore ObservationStore
	fileStore        FileStore
}

// NewExportHandler returns a new instance using the given dependencies.
func NewExportHandler(filterStore FilterStore, observationStore ObservationStore, fileStore FileStore) *ExportHandler {
	return &ExportHandler{
		filterStore:      filterStore,
		observationStore: observationStore,
		fileStore:        fileStore,
	}
}

// FilterStore provides existing filter data.
type FilterStore interface {
	GetFilters(filterJobID string) (*observation.Filter, error)
	PutCSVData(filterJobID string, url string, size int64) error
}

// ObservationStore provides filtered observation data in CSV rows.
type ObservationStore interface {
	GetCSVRows(filter *observation.Filter) (observation.CSVRowReader, error)
}

// FileStore provides storage for filtered output files.
type FileStore interface {
	PutFile(reader io.Reader) error
}

// Handle the export of a single filter job.
func (handler *ExportHandler) Handle(event *FilterJobSubmitted) error {

	// get filters from filter store
	filter, err := handler.filterStore.GetFilters(event.FilterJobID)
	if err != nil {
		return err
	}

	// get observation rows from observation store.
	csvRowReader, err := handler.observationStore.GetCSVRows(filter)
	if err != nil {
		return err
	}

	// upload to file store
	reader := observation.NewReader(csvRowReader)
	err = handler.fileStore.PutFile(reader)
	if err != nil {
		return err
	}

	// write url and file size to filter API

	// write output message

	return nil
}
