package event

import (
	"github.com/ONSdigital/dp-dataset-exporter/observation"
	"io"
)

//go:generate moq -out eventtest/filter_store.go -pkg eventtest . FilterStore
//go:generate moq -out eventtest/observation_store.go -pkg eventtest . ObservationStore
//go:generate moq -out eventtest/file_store.go -pkg eventtest . FileStore
//go:generate moq -out eventtest/producer.go -pkg eventtest . Producer

var _ Handler = (*ExportHandler)(nil)

// ExportHandler handles a single CSV export of a filtered dataset.
type ExportHandler struct {
	filterStore      FilterStore
	observationStore ObservationStore
	fileStore        FileStore
	eventProducer    Producer
}

// NewExportHandler returns a new instance using the given dependencies.
func NewExportHandler(filterStore FilterStore,
	observationStore ObservationStore,
	fileStore FileStore,
	eventProducer Producer) *ExportHandler {

	return &ExportHandler{
		filterStore:      filterStore,
		observationStore: observationStore,
		fileStore:        fileStore,
		eventProducer:    eventProducer,
	}
}

// FilterStore provides existing filter data.
type FilterStore interface {
	GetFilter(filterJobID string) (*observation.Filter, error)
	PutCSVData(filterJobID string, csvURL string, csvSize int64) error
}

// ObservationStore provides filtered observation data in CSV rows.
type ObservationStore interface {
	GetCSVRows(filter *observation.Filter) (observation.CSVRowReader, error)
}

// FileStore provides storage for filtered output files.
type FileStore interface {
	PutFile(reader io.Reader, filter *observation.Filter) (url string, err error)
}

// Producer handles producing output events.
type Producer interface {
	CSVExported(filterJobID string) error
}

// Handle the export of a single filter job.
func (handler *ExportHandler) Handle(event *FilterJobSubmitted) error {

	filter, err := handler.filterStore.GetFilter(event.FilterJobID)
	if err != nil {
		return err
	}

	csvRowReader, err := handler.observationStore.GetCSVRows(filter)
	if err != nil {
		return err
	}

	reader := observation.NewReader(csvRowReader)
	fileURL, err := handler.fileStore.PutFile(reader, filter)
	if err != nil {
		return err
	}

	// write url and file size to filter API
	err = handler.filterStore.PutCSVData(filter.JobID, fileURL, reader.TotalBytesRead())
	if err != nil {
		return err
	}

	return handler.eventProducer.CSVExported(filter.JobID)
}