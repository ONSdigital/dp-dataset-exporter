package event

import (
	"io"
	"strings"

	"github.com/satori/go.uuid"

	"strconv"

	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

//go:generate moq -out eventtest/filter_store.go -pkg eventtest . FilterStore
//go:generate moq -out eventtest/observation_store.go -pkg eventtest . ObservationStore
//go:generate moq -out eventtest/file_store.go -pkg eventtest . FileStore
//go:generate moq -out eventtest/producer.go -pkg eventtest . Producer
//go:generate moq -out eventtest/datasetapi.go -pkg eventtest . DatasetAPI

var _ Handler = (*ExportHandler)(nil)

// ExportHandler handles a single CSV export of a filtered dataset.
type ExportHandler struct {
	filterStore      FilterStore
	observationStore ObservationStore
	fileStore        FileStore
	eventProducer    Producer
	datasetAPICli    DatasetAPI
}

// DatasetAPI contains functions to call the dataset API.
type DatasetAPI interface {
	PutVersion(id, edition, version string, m dataset.Version) error
}

// NewExportHandler returns a new instance using the given dependencies.
func NewExportHandler(filterStore FilterStore,
	observationStore ObservationStore,
	fileStore FileStore,
	eventProducer Producer,
	datasetAPI DatasetAPI) *ExportHandler {

	return &ExportHandler{
		filterStore:      filterStore,
		observationStore: observationStore,
		fileStore:        fileStore,
		eventProducer:    eventProducer,
		datasetAPICli:    datasetAPI,
	}
}

// FilterStore provides existing filter data.
type FilterStore interface {
	GetFilter(filterID string) (*observation.Filter, error)
	PutCSVData(filterID string, csvURL string, csvSize int64) error
	PutStateAsEmpty(filterJobID string) error
	PutStateAsError(filterJobID string) error
}

// ObservationStore provides filtered observation data in CSV rows.
type ObservationStore interface {
	GetCSVRows(filter *observation.Filter, limit *int) (observation.CSVRowReader, error)
}

// FileStore provides storage for filtered output files.
type FileStore interface {
	PutFile(reader io.Reader, fileID string) (url string, err error)
}

// Producer handles producing output events.
type Producer interface {
	CSVExported(e *CSVExported) error
}

// Handle the export of a single filter output.
func (handler *ExportHandler) Handle(event *FilterSubmitted) error {
	var csvExported *CSVExported
	var err error

	if event.FilterID != "" {
		csvExported, err = handler.filterJob(event)
	} else {
		csvExported, err = handler.prePublishJob(event)
	}

	if err != nil {
		return err
	}

	if err := handler.eventProducer.CSVExported(csvExported); err != nil {
		return errors.Wrap(err, "error while attempting to send csv exported event to producer")
	}
	return nil
}

func (handler *ExportHandler) filterJob(event *FilterSubmitted) (*CSVExported, error) {
	log.Info("handling filter job", log.Data{"filter_id": event.FilterID})
	filter, err := handler.filterStore.GetFilter(event.FilterID)
	if err != nil {
		return nil, err
	}

	csvRowReader, err := handler.observationStore.GetCSVRows(filter, nil)
	if err != nil {
		return nil, err
	}

	reader := observation.NewReader(csvRowReader)
	defer func() {
		closeErr := reader.Close()
		if closeErr != nil {
			log.Error(closeErr, nil)
		}
	}()

	fileURL, err := handler.fileStore.PutFile(reader, filter.FilterID)
	if err != nil {
		return nil, err
		if strings.Contains(err.Error(), observation.ErrNoResultsFound.Error()) {
			log.Debug("empty results from filter job", log.Data{"instance_id": filter.InstanceID,
				"filter": filter})
			return nil, handler.filterStore.PutStateAsEmpty(filter.FilterID)
		} else if strings.Contains(err.Error(), observation.ErrNoInstanceFound.Error()) {
			log.Error(err, log.Data{"instance_id": filter.InstanceID,
				"filter": filter})
			return nil, handler.filterStore.PutStateAsError(filter.FilterID)
		}
		return nil, err
	}

	// write url and file size to filter API
	err = handler.filterStore.PutCSVData(filter.FilterID, fileURL, reader.TotalBytesRead())
	if err != nil {
		return nil, errors.Wrap(err, "error while putting CSV in filter store")
	}

	log.Info("CSV export completed", log.Data{"filter_id": filter.FilterID, "file_url": fileURL})
	return &CSVExported{FilterID: filter.FilterID, FileURL: fileURL}, nil
}

func (handler *ExportHandler) prePublishJob(event *FilterSubmitted) (*CSVExported, error) {
	log.Info("handling pre canned job", log.Data{
		"instance_id": event.InstanceID,
		"dataset_id":  event.DatasetID,
		"edition":     event.Edition,
		"version":     event.Version,
	})

	csvRowReader, err := handler.observationStore.GetCSVRows(&observation.Filter{InstanceID: event.InstanceID}, nil)
	if err != nil {
		return nil, err
	}

	reader := observation.NewReader(csvRowReader)
	defer func() {
		closeErr := reader.Close()
		if closeErr != nil {
			log.Error(closeErr, nil)
		}
	}()

	fileID := uuid.NewV4().String()
	log.Info("storing pre-publish file", log.Data{"fileID": fileID})

	fileURL, err := handler.fileStore.PutFile(reader, fileID)
	if err != nil {
		return nil, err
	}

	downloads := map[string]dataset.Download{
		"CSV": {Size: strconv.Itoa(int(reader.TotalBytesRead())), URL: fileURL},
	}

	v := dataset.Version{Downloads: downloads}

	if err := handler.datasetAPICli.PutVersion(event.DatasetID, event.Edition, event.Version, v); err != nil {
		return nil, errors.Wrap(err, "error while attempting update version downloads")
	}

	log.Info("pre publish version csv download file generation completed", log.Data{
		"dataset_id": event.DatasetID,
		"edition":    event.Edition,
		"version":    event.Version,
		"url":        fileURL,
	})

	return &CSVExported{
		InstanceID: event.InstanceID,
		DatasetID:  event.DatasetID,
		Edition:    event.Edition,
		Version:    event.Version,
		FileURL:    fileURL,
		Filename:   fileID,
	}, nil
}
