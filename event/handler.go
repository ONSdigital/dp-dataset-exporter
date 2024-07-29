package event

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/csvw"
	"github.com/ONSdigital/dp-dataset-exporter/reader"
	"github.com/ONSdigital/dp-graph/v2/observation"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

var _ Handler = (*ExportHandler)(nil)

const publishedState = "published"
const metadataExtension = "-metadata.json"

type EventGenerator interface {
	GenerateOutput(ctx context.Context, event *CSVExported) error
}

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

// Handle takes a single event.
func (handler *ExportHandler) Handle(ctx context.Context, cfg *config.Config, event *FilterSubmitted) (err error) {
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

		log.Info(ctx, "filter job identified", logData)
		csvExported, err = handler.filterJob(ctx, event, isPublished)
		if err != nil {
			log.Error(ctx, "error handling filter job", err, logData)
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

		log.Info(ctx, "dataset download job identified", logData)
		csvExported, err = handler.fullDownload(ctx, event, isPublished)
		if err != nil {
			log.Error(ctx, "error handling dataset download job", err, logData)
			return err
		}
	}

	if err := handler.eventProducer.CSVExported(ctx, csvExported); err != nil {
		return errors.Wrap(err, "error while attempting to send csv exported event to producer")
	}

	return nil
}

func (handler *ExportHandler) isFilterOutputPublished(ctx context.Context, event *FilterSubmitted) (bool, error) {
	filterStruct, err := handler.filterStore.GetFilter(ctx, event.FilterID)
	if err != nil {
		return false, err
	}

	return filterStruct.IsPublished == observation.Published, nil
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

	instance, _, err := handler.datasetAPICli.GetInstance(ctx, "", handler.serviceAuthToken, "", event.InstanceID, headers.IfMatchAnyETag)
	if err != nil {
		return false, err
	}

	return instance.State == publishedState, nil
}

// SortFilter by Dimension size. Largest first, to make Neptune searches faster
// The sort is done here because the sizes are retrieved from Mongo and
// its best not to have the dp-graph library acquiring such coupling to its caller.
var SortFilter = func(ctx context.Context, handler *ExportHandler, event *FilterSubmitted, dbFilter *observation.DimensionFilters) {
	nofDimensions := len(dbFilter.Dimensions)
	if nofDimensions <= 1 {
		return
	}
	// Create a slice of sorted dimension sizes
	type dim struct {
		index         int
		dimensionSize int
	}

	dimSizes := make([]dim, 0, nofDimensions)
	var dimSizesMutex sync.Mutex

	// get info from mongo
	var getErrorCount int32
	var concurrent = 10 // limit number of go routines to not put too much on heap
	var semaphoreChan = make(chan struct{}, concurrent)
	var wg sync.WaitGroup // number of working goroutines

	for i, dimension := range dbFilter.Dimensions {
		if atomic.LoadInt32(&getErrorCount) != 0 {
			break
		}
		semaphoreChan <- struct{}{} // block while full

		wg.Add(1)

		// Get dimension sizes in parallel
		go func(i int, dimension *observation.Dimension) {
			defer func() {
				<-semaphoreChan // read to release a slot
			}()

			defer wg.Done()

			// passing a 'Limit' of 0 makes GetOptions skip getting the documents
			// and to return only what we are interested in: TotalCount
			options, err := handler.datasetAPICli.GetOptions(ctx,
				"", // userAuthToken
				handler.serviceAuthToken,
				"", // collectionID
				event.DatasetID, event.Edition, event.Version, dimension.Name,
				&dataset.QueryParams{Offset: 0, Limit: 0})

			if err != nil {
				if atomic.AddInt32(&getErrorCount, 1) <= 2 {
					// only show a few of possibly hundreds of errors, as once someone
					// looks into the one error they may fix all associated errors
					logData := log.Data{"dataset_id": event.DatasetID, "edition": event.Edition, "version": event.Version, "dimension name": dimension.Name}
					log.Info(ctx, "SortFilter: GetOptions failed for dataset and dimension", logData)
				}
			} else {
				d := dim{dimensionSize: options.TotalCount, index: i}
				dimSizesMutex.Lock()
				dimSizes = append(dimSizes, d)
				dimSizesMutex.Unlock()
			}
		}(i, dimension)
	}
	wg.Wait()

	if getErrorCount != 0 {
		logData := log.Data{"dataset_id": event.DatasetID, "edition": event.Edition, "version": event.Version}
		log.Info(ctx, fmt.Sprintf("SortFilter: GetOptions failed for dataset %d times, sorting by default of 'geography' first", getErrorCount), logData)
		// Frig dimension sizes and if geography is present, make it the largest (because it typically is the largest)
		// and to retain compatibility with what the neptune dp-graph library was doing without access to information
		// from mongo.
		dimSizes = dimSizes[:0]
		for i, dimension := range dbFilter.Dimensions {
			if strings.ToLower(dimension.Name) == "geography" {
				d := dim{dimensionSize: 999999, index: i}
				dimSizes = append(dimSizes, d)
			} else {
				// Set sizes of dimensions as largest first to retain list order, to improve sort speed
				d := dim{dimensionSize: nofDimensions - i, index: i}
				dimSizes = append(dimSizes, d)
			}
		}
	}

	// sort slice by number of options per dimension. Smallest first
	sort.Slice(dimSizes, func(i, j int) bool {
		return dimSizes[i].dimensionSize < dimSizes[j].dimensionSize
	})

	sortedDimensions := make([]observation.Dimension, 0, nofDimensions)

	for i := nofDimensions - 1; i >= 0; i-- { // build required return structure. Largest first
		sortedDimensions = append(sortedDimensions, *dbFilter.Dimensions[dimSizes[i].index])
	}

	// Now copy the sorted dimensions back over the original
	for i, dimension := range sortedDimensions {
		*dbFilter.Dimensions[i] = dimension
	}
}

var CreateFilterForAll = func(ctx context.Context, handler *ExportHandler, event *FilterSubmitted, isPublished bool) (*observation.DimensionFilters, error) {
	// Get the names of the dimensions for the DatasetID
	fmt.Println("GETTING DIMENSIONS")
	dimensions, err := handler.datasetAPICli.GetVersionDimensions(ctx,
		"", // userAuthToken
		handler.serviceAuthToken,
		"", // collectionID
		event.DatasetID, event.Edition, event.Version,
	)
	if err != nil {
		return nil, err
	}

	var dbFilterDimensions []*observation.Dimension

	for _, dim := range dimensions.Items {
		// Get the names of the options for a dimension
		opts, err := handler.datasetAPICli.GetOptionsInBatches(ctx,
			"", // userAuthToken
			handler.serviceAuthToken,
			"", // collectionID
			event.DatasetID, event.Edition, event.Version, dim.Name,
			100,
			10,
		)
		if err != nil {
			return nil, err
		}

		var options []string
		for _, opt := range opts.Items {
			options = append(options, opt.Option)
		}

		// Build info to create filter later
		dbFilterDimensions = append(dbFilterDimensions, &observation.Dimension{
			Name:    dim.Name,
			Options: options,
		})
	}

	// Create filter that has ALL dimensions with ALL options for each dimension
	// for sorting
	dbFilter := &observation.DimensionFilters{
		Dimensions: dbFilterDimensions,
		Published:  &isPublished,
	}

	return dbFilter, nil
}

func (handler *ExportHandler) filterJob(ctx context.Context, event *FilterSubmitted, isPublished bool) (*CSVExported, error) {

	log.Info(ctx, "handling filter job", log.Data{"filter_id": event.FilterID})
	startTime := time.Now()
	filterStruct, err := handler.filterStore.GetFilter(ctx, event.FilterID)
	if err != nil {
		return nil, err
	}

	dbFilter := mapFilter(filterStruct)

	if dbFilter.IsEmpty() {
		dbFilter, err = CreateFilterForAll(ctx, handler, event, isPublished)
		if err != nil {
			return nil, errors.Wrap(err, "error while creating filter for all")
		}
	}

	sortFilterStartTime := time.Now()
	SortFilter(ctx, handler, event, dbFilter)
	sortFilterEndTime := time.Now()

	csvRowReader, err := handler.observationStore.StreamCSVRows(ctx, filterStruct.InstanceID, filterStruct.FilterID, dbFilter, nil)
	if err != nil {
		return nil, err
	}

	exporterReader := observation.NewReader(csvRowReader)
	defer func() {
		closeErr := exporterReader.Close(ctx)
		if closeErr != nil {
			log.Error(ctx, "error closing reader", closeErr)
		}
	}()

	timestamp := strings.ReplaceAll(time.Now().UTC().Format(time.RFC3339), ":", "-") // ISO8601, with colon ':' replaced by hyphen '-'
	s := strings.Split(timestamp, "+")                                               // Strip off timezone data
	name := event.FilterID + "/" + event.DatasetID + "-" + event.Edition + "-v" + event.Version + "-filtered-" + s[0]

	log.Info(ctx, "storing filtered dataset file", log.Data{"name": name})

	filename := handler.filteredDatasetFilePrefix + name + ".csv"

	// When getting the data from the reader, this will call the neo4j driver to start streaming the data
	// into the S3 library. We can only tell if data is present by reading the stream.
	fileURL, err := handler.fileStore.PutFile(ctx, exporterReader, filename, isPublished)
	if err != nil {
		if strings.Contains(err.Error(), observation.ErrNoResultsFound.Error()) {
			log.Info(ctx, "empty results from filter job", log.Data{"instance_id": filterStruct.InstanceID,
				"filter": filterStruct})
			updateErr := handler.filterStore.PutStateAsEmpty(ctx, filterStruct.FilterID)
			if updateErr != nil {
				return nil, updateErr
			}
			return nil, err
		} else if strings.Contains(err.Error(), observation.ErrNoInstanceFound.Error()) {
			log.Error(ctx, "instance not found", err, log.Data{"instance_id": filterStruct.InstanceID, "filter": filterStruct})
			updateErr := handler.filterStore.PutStateAsError(ctx, filterStruct.FilterID)
			if updateErr != nil {
				return nil, updateErr
			}
			return nil, err
		}
		return nil, err
	}

	csv := createCSVDownloadData(handler, event, isPublished, exporterReader, fileURL)

	// write url and file size to filter API
	err = handler.filterStore.PutCSVData(ctx, filterStruct.FilterID, csv)
	if err != nil {
		return nil, errors.Wrap(err, "error while putting CSV in filter store")
	}

	rowCount := exporterReader.ObservationsCount()

	totTime := time.Now()

	sortFilterDiff := sortFilterEndTime.Sub(sortFilterStartTime)
	fd := sortFilterDiff.Microseconds()
	seconds := fd / 1000000
	microsecs := fd % 1000000
	sortFilterTime := fmt.Sprintf("%d.%d seconds", seconds, microsecs)

	totalDiff := totTime.Sub(startTime)
	td := totalDiff.Microseconds()
	seconds = td / 1000000
	microsecs = td % 1000000
	totalTime := fmt.Sprintf("%d.%d seconds", seconds, microsecs)

	log.Info(ctx, "csv export completed",
		log.Data{
			"filter_id":        filterStruct.FilterID,
			"file_url":         fileURL,
			"row_count":        rowCount,
			"sort_filter_time": sortFilterTime,
			"total_time":       totalTime,
		})
	return &CSVExported{FilterID: filterStruct.FilterID, DatasetID: event.DatasetID, Edition: event.Edition, Version: event.Version, FileURL: fileURL, RowCount: rowCount}, nil
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
	log.Info(ctx, "handling pre canned job", log.Data{
		"instance_id": event.InstanceID,
		"dataset_id":  event.DatasetID,
		"edition":     event.Edition,
		"version":     event.Version,
	})

	name := event.DatasetID + "-" + event.Edition + "-v" + event.Version
	log.Info(ctx, "storing pre-publish file", log.Data{"name": name})

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

	log.Info(ctx, "pre publish version csv download file generation completed", log.Data{
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
	dbFilter, err := CreateFilterForAll(ctx, handler, event, isPublished)
	if err != nil {
		return nil, "", "", 0, errors.Wrap(err, "error while creating filter for all")
	}
	SortFilter(ctx, handler, event, dbFilter)

	csvRowReader, err := handler.observationStore.StreamCSVRows(ctx, event.InstanceID, "", dbFilter, nil)
	if err != nil {
		return nil, "", "", 0, err
	}

	rReader := observation.NewReader(csvRowReader)
	defer func() (*dataset.Download, string, string, int32, error) {
		closeErr := rReader.Close(ctx)
		if closeErr != nil {
			log.Error(ctx, "error closing observation reader", closeErr)
		}
		return nil, "", "", 0, closeErr
	}()

	hReader := reader.New(rReader)

	header, err := hReader.PeekBytes('\n')

	if err != nil {
		return nil, "", "", 0, errors.Wrap(err, "could not peek")
	}

	log.Info(ctx, "header extracted from csv", log.Data{"header": header})

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

	log.Info(ctx, "updating dataset api with download link", log.Data{
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
