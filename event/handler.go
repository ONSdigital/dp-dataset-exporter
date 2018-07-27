package event

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	"github.com/ONSdigital/dp-dataset-exporter/csvw"
	"github.com/ONSdigital/dp-dataset-exporter/reader"
	"github.com/ONSdigital/dp-filter/observation"

	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/log"
)

//go:generate moq -out eventtest/filter_store.go -pkg eventtest . FilterStore
//go:generate moq -out eventtest/observation_store.go -pkg eventtest . ObservationStore
//go:generate moq -out eventtest/file_store.go -pkg eventtest . FileStore
//go:generate moq -out eventtest/producer.go -pkg eventtest . Producer
//go:generate moq -out eventtest/datasetapi.go -pkg eventtest . DatasetAPI

var _ Handler = (*ExportHandler)(nil)

const publishedState = "published"
const metadataExtension = "-metadata.json"

// ExportHandler handles a single CSV export of a filtered dataset.
type ExportHandler struct {
	filterStore               FilterStore
	observationStore          ObservationStore
	fileStore                 FileStore
	eventProducer             Producer
	datasetAPICli             DatasetAPI
	downloadServiceURL        string
	apiDomainURL              string
	fullDatasetFilePrefix     string
	filteredDatasetFilePrefix string
}

// DatasetAPI contains functions to call the dataset API.
type DatasetAPI interface {
	PutVersion(ctx context.Context, id, edition, version string, m dataset.Version) error
	GetVersion(ctx context.Context, id, edition, version string) (m dataset.Version, err error)
	GetInstance(ctx context.Context, instanceID string) (m dataset.Instance, err error)
	GetMetadataURL(id, edition, version string) (url string)
	GetVersionMetadata(ctx context.Context, id, edition, version string) (m dataset.Metadata, err error)
}

// NewExportHandler returns a new instance using the given dependencies.
func NewExportHandler(
	filterStore FilterStore,
	observationStore ObservationStore,
	fileStore FileStore,
	eventProducer Producer,
	datasetAPI DatasetAPI,
	downloadServiceURL,
	apiDomainURL,
	fullDatasetFilePrefix,
	filteredDatasetFilePrefix string,
) *ExportHandler {

	return &ExportHandler{
		filterStore:               filterStore,
		observationStore:          observationStore,
		fileStore:                 fileStore,
		eventProducer:             eventProducer,
		datasetAPICli:             datasetAPI,
		downloadServiceURL:        downloadServiceURL,
		apiDomainURL:              apiDomainURL,
		fullDatasetFilePrefix:     fullDatasetFilePrefix,
		filteredDatasetFilePrefix: filteredDatasetFilePrefix,
	}
}

// FilterStore provides existing filter data.
type FilterStore interface {
	GetFilter(filterID string) (*observation.Filter, error)
	PutCSVData(filterID string, downloadItem observation.DownloadItem) error
	PutStateAsEmpty(filterJobID string) error
	PutStateAsError(filterJobID string) error
}

// ObservationStore provides filtered observation data in CSV rows.
type ObservationStore interface {
	GetCSVRows(filter *observation.Filter, limit *int) (observation.CSVRowReader, error)
}

// FileStore provides storage for filtered output files.
type FileStore interface {
	PutFile(reader io.Reader, filename string, isPublished bool) (url string, err error)
}

// Producer handles producing output events.
type Producer interface {
	CSVExported(e *CSVExported) error
}

// Handle the export of a single filter output.
func (handler *ExportHandler) Handle(ctx context.Context, event *FilterSubmitted) error {
	var csvExported *CSVExported

	if event.FilterID != "" {

		isPublished, err := handler.isFilterOutputPublished(event)
		if err != nil {
			return err
		}

		logData := log.Data{"filter_id": event.FilterID, "published": isPublished}

		log.Debug("filter job identified", logData)
		csvExported, err = handler.filterJob(event, isPublished)
		if err != nil {
			log.ErrorC("error handling filter job", err, logData)
			return err
		}
	} else {
		isPublished, err := handler.isVersionPublished(ctx, event)
		if err != nil {
			return err
		}

		logData := log.Data{"instance_id": event.InstanceID,
			"dataset_id": event.DatasetID,
			"edition":    event.Edition,
			"version":    event.Version,
			"published":  isPublished}

		log.Debug("dataset download job identified", logData)
		csvExported, err = handler.fullDownload(ctx, event, isPublished)
		if err != nil {
			log.ErrorC("error handling dataset download job", err, logData)
			return err
		}
	}

	if err := handler.eventProducer.CSVExported(csvExported); err != nil {
		return errors.Wrap(err, "error while attempting to send csv exported event to producer")
	}
	return nil
}

func (handler *ExportHandler) isFilterOutputPublished(event *FilterSubmitted) (bool, error) {
	filter, err := handler.filterStore.GetFilter(event.FilterID)
	if err != nil {
		return false, err
	}

	return filter.Published != nil && *filter.Published == observation.Published, nil
}

func (handler *ExportHandler) isVersionPublished(ctx context.Context, event *FilterSubmitted) (bool, error) {
	if len(event.InstanceID) == 0 {
		version, err := handler.datasetAPICli.GetVersion(ctx, event.DatasetID, event.Edition, event.Version)
		if err != nil {
			return false, err
		}

		return version.State == publishedState, nil
	}

	instance, err := handler.datasetAPICli.GetInstance(ctx, event.InstanceID)
	if err != nil {
		return false, err
	}

	return instance.State == publishedState, nil
}

func (handler *ExportHandler) filterJob(event *FilterSubmitted, isPublished bool) (*CSVExported, error) {
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

	filename := handler.filteredDatasetFilePrefix + filter.FilterID + ".csv"

	// When getting the data from the reader, this will call the neo4j driver to start streaming the data
	// into the S3 library. We can only tell if data is present by reading the stream.
	fileURL, err := handler.fileStore.PutFile(reader, filename, isPublished)
	if err != nil {
		if strings.Contains(err.Error(), observation.ErrNoResultsFound.Error()) {
			log.Debug("empty results from filter job", log.Data{"instance_id": filter.InstanceID,
				"filter": filter})
			updateErr := handler.filterStore.PutStateAsEmpty(filter.FilterID)
			if updateErr != nil {
				return nil, updateErr
			}
			return nil, err
		} else if strings.Contains(err.Error(), observation.ErrNoInstanceFound.Error()) {
			log.Error(err, log.Data{"instance_id": filter.InstanceID,
				"filter": filter})
			updateErr := handler.filterStore.PutStateAsError(filter.FilterID)
			if updateErr != nil {
				return nil, updateErr
			}
			return nil, err
		}
		return nil, err
	}

	var csv observation.DownloadItem
	downloadURL := fmt.Sprintf("%s/downloads/filter-outputs/%s.csv",
		handler.downloadServiceURL,
		event.FilterID,
	)

	if isPublished {
		csv = observation.DownloadItem{Size: strconv.Itoa(int(reader.TotalBytesRead())), Public: fileURL, HRef: downloadURL}
	} else {
		csv = observation.DownloadItem{Size: strconv.Itoa(int(reader.TotalBytesRead())), Private: fileURL, HRef: downloadURL}
	}

	// write url and file size to filter API
	err = handler.filterStore.PutCSVData(filter.FilterID, csv)
	if err != nil {
		return nil, errors.Wrap(err, "error while putting CSV in filter store")
	}

	log.Info("CSV export completed", log.Data{"filter_id": filter.FilterID, "file_url": fileURL})
	return &CSVExported{FilterID: filter.FilterID, FileURL: fileURL}, nil
}

func (handler *ExportHandler) fullDownload(ctx context.Context, event *FilterSubmitted, isPublished bool) (*CSVExported, error) {
	log.Info("handling pre canned job", log.Data{
		"instance_id": event.InstanceID,
		"dataset_id":  event.DatasetID,
		"edition":     event.Edition,
		"version":     event.Version,
	})

	fileID := uuid.NewV4().String()
	log.Info("storing pre-publish file", log.Data{"fileID": fileID})

	filename := handler.fullDatasetFilePrefix + fileID + ".csv"

	csvDownload, csvS3URL, header, err := handler.generateFullCSV(event, filename, isPublished)
	if err != nil {
		return nil, err
	}

	downloadURL := fmt.Sprintf("%s/downloads/datasets/%s/editions/%s/versions/%s.csv",
		handler.downloadServiceURL,
		event.DatasetID,
		event.Edition,
		event.Version,
	)

	metadataDownload, err := handler.generateMetadata(event, filename, header, downloadURL, isPublished)
	if err != nil {
		return nil, err
	}

	if err := handler.updateVersionLinks(event, isPublished, csvDownload, metadataDownload, downloadURL); err != nil {
		return nil, err
	}

	log.Info("pre publish version csv download file generation completed", log.Data{
		"dataset_id":        event.DatasetID,
		"edition":           event.Edition,
		"version":           event.Version,
		"csv_download":      csvDownload,
		"metadata_download": metadataDownload,
	})

	return &CSVExported{
		InstanceID: event.InstanceID,
		DatasetID:  event.DatasetID,
		Edition:    event.Edition,
		Version:    event.Version,
		FileURL:    csvS3URL,
		Filename:   fileID,
	}, nil
}

func (handler *ExportHandler) generateFullCSV(event *FilterSubmitted, filename string, isPublished bool) (*dataset.Download, string, string, error) {

	csvRowReader, err := handler.observationStore.GetCSVRows(&observation.Filter{InstanceID: event.InstanceID}, nil)
	if err != nil {
		return nil, "", "", err
	}

	rReader := observation.NewReader(csvRowReader)
	defer func() (*dataset.Download, string, string, error) {
		closeErr := rReader.Close()
		if closeErr != nil {
			log.Error(closeErr, nil)
		}
		return nil, "", "", closeErr
	}()

	hReader := reader.New(rReader)

	header, err := hReader.PeekBytes('\n')
	if err != nil {
		return nil, "", "", errors.Wrap(err, "could not peek")
	}

	log.Info("header extracted from csv", log.Data{"header": header})

	url, err := handler.fileStore.PutFile(hReader, filename, isPublished)
	if err != nil {
		return nil, "", "", err
	}

	download := &dataset.Download{
		Size: strconv.Itoa(hReader.TotalBytesRead()),
	}

	if isPublished {
		download.Public = url
	} else {
		download.Private = url
	}

	return download, url, header, nil
}

func (handler *ExportHandler) generateMetadata(event *FilterSubmitted, s3path, header, downloadURL string, isPublished bool) (*dataset.Download, error) {
	m, err := handler.datasetAPICli.GetVersionMetadata(context.Background(), event.DatasetID, event.Edition, event.Version)
	if err != nil {
		return nil, err
	}

	aboutURL := handler.datasetAPICli.GetMetadataURL(event.DatasetID, event.Edition, event.Version)

	csvwFile, err := csvw.Generate(&m, header, downloadURL, aboutURL, handler.apiDomainURL)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(csvwFile)

	url, err := handler.fileStore.PutFile(r, s3path+metadataExtension, isPublished)
	if err != nil {
		return nil, err
	}

	download := &dataset.Download{
		Size: strconv.Itoa(int(r.Size())),
	}

	if isPublished {
		download.Public = url
	} else {
		download.Private = url
	}

	return download, nil
}

func (handler *ExportHandler) updateVersionLinks(event *FilterSubmitted, isPublished bool, csv, md *dataset.Download, downloadURL string) error {
	csv.URL = downloadURL
	md.URL = downloadURL + metadataExtension

	log.Info("updating dataset api with download link", log.Data{
		"isPublished":      isPublished,
		"csvDownload":      csv,
		"metadataDownload": md,
	})

	v := dataset.Version{
		Downloads: map[string]dataset.Download{
			"CSV":  *csv,
			"CSVW": *md,
		},
	}

	if err := handler.datasetAPICli.PutVersion(context.Background(), event.DatasetID, event.Edition, event.Version, v); err != nil {
		return errors.Wrap(err, "error while attempting update version downloads")
	}

	return nil
}
