package event

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/filter"

	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/csvw"
	"github.com/ONSdigital/dp-dataset-exporter/reader"
	"github.com/ONSdigital/dp-graph/v2/observation"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/log.go/log"
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
	serviceAuthToken          string
}

// DatasetAPI contains functions to call the dataset API.
type DatasetAPI interface {
	PutVersion(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version string, m dataset.Version) error
	GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error)
	GetInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string) (m dataset.Instance, err error)
	GetMetadataURL(id, edition, version string) (url string)
	GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.Metadata, err error)
	GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, q *dataset.QueryParams) (m dataset.Options, err error)
}

// NewExportHandler returns a new instance using the given dependencies.
func NewExportHandler(
	filterStore FilterStore,
	observationStore ObservationStore,
	fileStore FileStore,
	eventProducer Producer,
	datasetAPI DatasetAPI,
	cfg *config.Config,
) *ExportHandler {

	return &ExportHandler{
		filterStore:               filterStore,
		observationStore:          observationStore,
		fileStore:                 fileStore,
		eventProducer:             eventProducer,
		datasetAPICli:             datasetAPI,
		downloadServiceURL:        cfg.DownloadServiceURL,
		apiDomainURL:              cfg.APIDomainURL,
		fullDatasetFilePrefix:     cfg.FullDatasetFilePrefix,
		filteredDatasetFilePrefix: cfg.FilteredDatasetFilePrefix,
		serviceAuthToken:          cfg.ServiceAuthToken,
	}
}

// FilterStore provides existing filter data.
type FilterStore interface {
	GetFilter(ctx context.Context, filterID string) (*filter.Model, error)
	PutCSVData(ctx context.Context, filterID string, downloadItem filter.Download) error
	PutStateAsEmpty(ctx context.Context, filterJobID string) error
	PutStateAsError(ctx context.Context, filterJobID string) error
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
	CSVExported(e *CSVExported) error
}

// Handle the export of a single filter output.
func (handler *ExportHandler) Handle(ctx context.Context, event *FilterSubmitted) error {
	var csvExported *CSVExported
	var logData log.Data

	if event.FilterID != "" {

		isPublished, err := handler.isFilterOutputPublished(ctx, event)
		if err != nil {
			return err
		}

		logData = log.Data{"filter_id": event.FilterID,
			"dataset_id": event.DatasetID,
			"edition":    event.Edition,
			"version":    event.Version,
			"published":  isPublished}

		log.Event(ctx, "filter job identified", log.INFO, logData)
		csvExported, err = handler.filterJob(ctx, event, isPublished)
		if err != nil {
			log.Event(ctx, "error handling filter job", log.ERROR, logData, log.Error(err))
			return err
		}
	} else {
		isPublished, err := handler.isVersionPublished(ctx, event)
		if err != nil {
			return err
		}

		logData = log.Data{"instance_id": event.InstanceID,
			"dataset_id": event.DatasetID,
			"edition":    event.Edition,
			"version":    event.Version,
			"published":  isPublished}

		log.Event(ctx, "dataset download job identified", log.INFO, logData)
		csvExported, err = handler.fullDownload(ctx, event, isPublished)
		if err != nil {
			log.Event(ctx, "error handling dataset download job", log.ERROR, logData, log.Error(err))
			return err
		}
	}

	if err := handler.eventProducer.CSVExported(csvExported); err != nil {
		return errors.Wrap(err, "error while attempting to send csv exported event to producer")
	}

	return nil
}

func (handler *ExportHandler) isFilterOutputPublished(ctx context.Context, event *FilterSubmitted) (bool, error) {
	filter, err := handler.filterStore.GetFilter(ctx, event.FilterID)
	if err != nil {
		return false, err
	}

	return filter.IsPublished == observation.Published, nil
}

func (handler *ExportHandler) isVersionPublished(ctx context.Context, event *FilterSubmitted) (bool, error) {

	if len(event.InstanceID) == 0 {
		version, err := handler.datasetAPICli.GetVersion(
			ctx, "", handler.serviceAuthToken, "", "", event.DatasetID, event.Edition, event.Version)
		if err != nil {
			return false, err
		}

		return version.State == publishedState, nil
	}

	instance, err := handler.datasetAPICli.GetInstance(ctx, "", handler.serviceAuthToken, "", event.InstanceID)
	if err != nil {
		return false, err
	}

	return instance.State == publishedState, nil
}

var callCount int

func (handler *ExportHandler) filterJob(ctx context.Context, event *FilterSubmitted, isPublished bool) (*CSVExported, error) {

	log.Event(ctx, "handling filter job", log.INFO, log.Data{"filter_id": event.FilterID})
	filter, err := handler.filterStore.GetFilter(ctx, event.FilterID)
	if err != nil {
		return nil, err
	}

	// =====
	fmt.Printf("\ndoing: StreamCSVRows\n")
	//fmt.Printf("%+v\n", filter)
	start := time.Now()
	// =====
	dbFilter := mapFilter(filter)
	//spew.Dump(dbFilter)

	// get info from mongo

	for i, dimension := range dbFilter.Dimensions {

		options, err := handler.datasetAPICli.GetOptions(ctx,
			"", // userAuthToken ??
			handler.serviceAuthToken,
			"", // collectionID, // ??
			event.DatasetID,
			event.Edition,
			event.Version,
			dimension.Name,
			&dataset.QueryParams{Offset: 0, Limit: 0})
		if err != nil {
			fmt.Printf("options err: %v\n", err)
			return nil, err
		} else {
			fmt.Printf("index %d. ID: %s, name %s, options.TotalCount: %d\n", i, event.FilterID, dimension.Name, options.TotalCount)
		}
	}
	callCount++
	fmt.Printf("callCount: %d\n", callCount)

	csvRowReader, err := handler.observationStore.StreamCSVRows(ctx, filter.InstanceID, filter.FilterID, dbFilter, nil)
	if err != nil {
		return nil, err
	}
	//	return &CSVExported{FilterID: filter.FilterID, DatasetID: event.DatasetID, Edition: event.Edition, Version: event.Version, FileURL: "/bla/bla", RowCount: 10}, nil

	reader := observation.NewReader(csvRowReader)
	defer func() {
		closeErr := reader.Close(ctx)
		if closeErr != nil {
			log.Event(ctx, "error closing reader", log.ERROR, log.Error(closeErr))
		}
	}()

	timestamp := strings.ReplaceAll(time.Now().UTC().Format(time.RFC3339), ":", "-") // ISO8601, with colon ':' replaced by hyphen '-'
	s := strings.Split(timestamp, "+")                                               // Strip off timezone data
	name := event.FilterID + "/" + event.DatasetID + "-" + event.Edition + "-v" + event.Version + "-filtered-" + s[0]

	log.Event(ctx, "storing filtered dataset file", log.INFO, log.Data{"name": name})

	filename := handler.filteredDatasetFilePrefix + name + ".csv"

	// When getting the data from the reader, this will call the neo4j driver to start streaming the data
	// into the S3 library. We can only tell if data is present by reading the stream.
	fileURL, err := handler.fileStore.PutFile(ctx, reader, filename, isPublished)
	if err != nil {
		if strings.Contains(err.Error(), observation.ErrNoResultsFound.Error()) {
			log.Event(ctx, "empty results from filter job", log.INFO, log.Data{"instance_id": filter.InstanceID,
				"filter": filter})
			updateErr := handler.filterStore.PutStateAsEmpty(ctx, filter.FilterID)
			if updateErr != nil {
				return nil, updateErr
			}
			return nil, err
		} else if strings.Contains(err.Error(), observation.ErrNoInstanceFound.Error()) {
			log.Event(ctx, "instance not found", log.ERROR, log.Data{"instance_id": filter.InstanceID,
				"filter": filter}, log.Error(err))
			updateErr := handler.filterStore.PutStateAsError(ctx, filter.FilterID)
			if updateErr != nil {
				return nil, updateErr
			}
			return nil, err
		}
		return nil, err
	}

	csv := createCSVDownloadData(handler, event, isPublished, reader, fileURL)

	// write url and file size to filter API
	err = handler.filterStore.PutCSVData(ctx, filter.FilterID, csv)
	if err != nil {
		return nil, errors.Wrap(err, "error while putting CSV in filter store")
	}

	rowCount := reader.ObservationsCount()

	// =====
	endTime := time.Now()

	elapsed := endTime.Sub(start) //time.Since(start)
	res := fmt.Sprintf("done: StreamCSVRows  %s -> %s\n", filter.FilterID, elapsed)
	fmt.Printf("\n%s", res)
	e := fmt.Sprintf("%s", elapsed)
	if strings.Contains(e, "ms") {
		e = e[0 : len(e)-2]
		n, _ := strconv.ParseFloat(e, 64)
		n = n * 1000.0
		parts := strings.Split(fmt.Sprintf("%f", n), ".")
		res = fmt.Sprintf("%s, %s, %s\n", fmt.Sprint(start.Format("15:04:05.000000")), filter.FilterID, parts[0])
	} else if strings.Contains(e, "µs") {
		parts := strings.Split(e, ".")
		p := strings.Replace(parts[0], "µs", "", 1)
		res = fmt.Sprintf("%s, %s, %s\n", fmt.Sprint(start.Format("15:04:05.000000")), filter.FilterID, p)
	} else {
		e = e[0 : len(e)-1]
		n, _ := strconv.ParseFloat(e, 64)
		n = n * 1000000.0
		parts := strings.Split(fmt.Sprintf("%f", n), ".")
		res = fmt.Sprintf("%s, %s, %s\n", fmt.Sprint(start.Format("15:04:05.000000")), filter.FilterID, parts[0])
	}
	f, err := os.OpenFile("timings.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Can't access file: timings.txt\n")
	} else {
		defer f.Close()
		_, err := f.WriteString(res)
		if err != nil {
			fmt.Printf("Can't write file: timings.txt\n")
		}
	}
	// =====

	log.Event(ctx, "csv export completed", log.INFO, log.Data{"filter_id": filter.FilterID, "file_url": fileURL, "row_count": rowCount})
	return &CSVExported{FilterID: filter.FilterID, DatasetID: event.DatasetID, Edition: event.Edition, Version: event.Version, FileURL: fileURL, RowCount: rowCount}, nil
}

func mapFilter(filter *filter.Model) *observation.DimensionFilters {

	dbFilterDimensions := mapFilterDimensions(filter)

	dbFilter := &observation.DimensionFilters{
		Dimensions: dbFilterDimensions,
		Published:  &filter.IsPublished,
	}

	return dbFilter
}

func mapFilterDimensions(filter *filter.Model) []*observation.Dimension {

	var dbFilterDimensions []*observation.Dimension

	for _, dimension := range filter.Dimensions {
		dbFilterDimensions = append(dbFilterDimensions, &observation.Dimension{
			Name:    dimension.Name,
			Options: dimension.Options,
		})
	}

	return dbFilterDimensions
}

func createCSVDownloadData(handler *ExportHandler, event *FilterSubmitted, isPublished bool, reader *observation.Reader, fileURL string) filter.Download {

	downloadURL := fmt.Sprintf("%s/downloads/filter-outputs/%s.csv",
		handler.downloadServiceURL,
		event.FilterID,
	)

	if isPublished {
		return filter.Download{Size: strconv.Itoa(int(reader.TotalBytesRead())), Public: fileURL, URL: downloadURL}
	}

	return filter.Download{Size: strconv.Itoa(int(reader.TotalBytesRead())), Private: fileURL, URL: downloadURL}
}

func (handler *ExportHandler) fullDownload(ctx context.Context, event *FilterSubmitted, isPublished bool) (*CSVExported, error) {
	log.Event(ctx, "handling pre canned job", log.INFO, log.Data{
		"instance_id": event.InstanceID,
		"dataset_id":  event.DatasetID,
		"edition":     event.Edition,
		"version":     event.Version,
	})

	name := event.DatasetID + "-" + event.Edition + "-v" + event.Version
	log.Event(ctx, "storing pre-publish file", log.INFO, log.Data{"name": name})

	filename := handler.fullDatasetFilePrefix + name + ".csv"

	csvDownload, csvS3URL, header, rowCount, err := handler.generateFullCSV(ctx, event, filename, isPublished)
	if err != nil {
		return nil, err
	}

	downloadURL := fmt.Sprintf("%s/downloads/datasets/%s/editions/%s/versions/%s.csv",
		handler.downloadServiceURL,
		event.DatasetID,
		event.Edition,
		event.Version,
	)

	metadataDownload, err := handler.generateMetadata(ctx, event, filename, header, downloadURL, isPublished)
	if err != nil {
		return nil, err
	}

	if err := handler.updateVersionLinks(ctx, event, isPublished, csvDownload, metadataDownload, downloadURL); err != nil {
		return nil, err
	}

	log.Event(ctx, "pre publish version csv download file generation completed", log.INFO, log.Data{
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
		Filename:   name,
		RowCount:   rowCount,
	}, nil
}

func (handler *ExportHandler) generateFullCSV(ctx context.Context, event *FilterSubmitted, filename string, isPublished bool) (*dataset.Download, string, string, int32, error) {
	csvRowReader, err := handler.observationStore.StreamCSVRows(ctx, event.InstanceID, "", &observation.DimensionFilters{}, nil)
	if err != nil {
		return nil, "", "", 0, err
	}

	rReader := observation.NewReader(csvRowReader)
	defer func() (*dataset.Download, string, string, int32, error) {
		closeErr := rReader.Close(ctx)
		if closeErr != nil {
			log.Event(ctx, "error closing observation reader", log.ERROR, log.Error(closeErr))
		}
		return nil, "", "", 0, closeErr
	}()

	hReader := reader.New(rReader)

	header, err := hReader.PeekBytes('\n')
	if err != nil {
		return nil, "", "", 0, errors.Wrap(err, "could not peek")
	}

	log.Event(ctx, "header extracted from csv", log.INFO, log.Data{"header": header})

	url, err := handler.fileStore.PutFile(ctx, hReader, filename, isPublished)
	if err != nil {
		return nil, "", "", 0, err
	}

	download := &dataset.Download{
		Size: strconv.Itoa(hReader.TotalBytesRead()),
	}

	if isPublished {
		download.Public = url
	} else {
		download.Private = url
	}

	return download, url, header, rReader.ObservationsCount(), nil
}

func (handler *ExportHandler) generateMetadata(ctx context.Context, event *FilterSubmitted, s3path, header, downloadURL string, isPublished bool) (*dataset.Download, error) {

	m, err := handler.datasetAPICli.GetVersionMetadata(ctx, "", handler.serviceAuthToken, "", event.DatasetID, event.Edition, event.Version)
	if err != nil {
		return nil, err
	}

	aboutURL := handler.datasetAPICli.GetMetadataURL(event.DatasetID, event.Edition, event.Version)

	csvwFile, err := csvw.Generate(ctx, &m, header, downloadURL, aboutURL, handler.apiDomainURL)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(csvwFile)

	url, err := handler.fileStore.PutFile(ctx, r, s3path+metadataExtension, isPublished)
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

func (handler *ExportHandler) updateVersionLinks(ctx context.Context, event *FilterSubmitted, isPublished bool, csv, md *dataset.Download, downloadURL string) error {
	csv.URL = downloadURL
	md.URL = downloadURL + metadataExtension

	log.Event(ctx, "updating dataset api with download link", log.INFO, log.Data{
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

	err := handler.datasetAPICli.PutVersion(
		ctx, "", handler.serviceAuthToken, "", event.DatasetID, event.Edition, event.Version, v)
	if err != nil {
		return errors.Wrap(err, "error while attempting update version downloads")
	}

	return nil
}
