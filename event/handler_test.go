package event_test

import (
	"context"
	"io"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	filterCli "github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-graph/v2/observation"
	"github.com/ONSdigital/dp-graph/v2/observation/observationtest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	filterOutputId  = "345"
	fileHRef        = "s3://some/url/123.csv"
	associatedState = "associated"

	downloadServiceURL        = "http://download-service"
	apiDomainURL              = "http://api-example"
	fullDatasetFilePrefix     = "full-dataset"
	filteredDatasetFilePrefix = "filtered-dataset"
	serviceAuthToken          = "serviceAuthToken"
)

// testing Config struct
var cfg = &config.Config{
	DownloadServiceURL:        downloadServiceURL,
	APIDomainURL:              apiDomainURL,
	FullDatasetFilePrefix:     fullDatasetFilePrefix,
	FilteredDatasetFilePrefix: filteredDatasetFilePrefix,
	ServiceAuthToken:          serviceAuthToken,
}

var filterSubmittedEvent = &event.FilterSubmitted{
	FilterID:  filterOutputId,
	DatasetID: "12345",
	Edition:   "2018",
	Version:   "1",
}

var fullFileDownloadSubmittedEvent = &event.FilterSubmitted{
	DatasetID: "12345",
	Edition:   "2018",
	Version:   "1",
}

var filter = &filterCli.Model{
	FilterID: filterOutputId,
	Dimensions: []filterCli.ModelDimension{
		{Name: "age", Options: []string{"29", "30"}},
		{Name: "sex", Options: []string{"male", "female"}},
	},
}

var associatedDataset = dataset.Version{
	State: associatedState,
}

var metadata = dataset.Metadata{
	Version: dataset.Version{
		Downloads: map[string]dataset.Download{
			"CSV": {
				URL: "/url",
			},
		},
		Dimensions: []dataset.VersionDimension{
			{},
			{},
		},
	},
	DatasetDetails: dataset.DatasetDetails{
		Publisher: &dataset.Publisher{},
	},
}

var csvBytes = []byte(`v4_0, a,b,c,
	1,e,f,g,
	2,h,i,j,
	`)

var csvContent = string(csvBytes)

var ctx = context.Background()

func TestExportHandler_Handle_FilterStoreGetError(t *testing.T) {
	Convey("Given a handler with a mock filter store that returns an error", t, func() {

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string, string, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockError := errors.New("get filters failed")

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(ctx context.Context, filterJobId string) (*filterCli.Model, error) {
				return nil, mockError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, nil, nil, nil, datasetAPIMock, cfg)

		Convey("When handle is called", func() {

			err := handler.Handle(ctx, cfg, filterSubmittedEvent)

			Convey("The error from the filter store is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, mockError.Error())
			})
		})
	})
}

func TestExportHandler_Handle_ObservationStoreError(t *testing.T) {

	Convey("Given a handler with a mocked observation store that returns an error", t, func() {

		expectedError := errors.New("something bad happened in the database")

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string, string, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		var mockFilterStore = &eventtest.FilterStoreMock{
			GetFilterFunc: func(ctx context.Context, filterJobId string) (*filterCli.Model, error) {
				return filter, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
				return nil, expectedError
			},
		}

		originalFunc := event.SortFilter
		defer func() {
			event.SortFilter = originalFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, nil, nil, datasetAPIMock, cfg)

		Convey("When handle is called", func() {

			err := handler.Handle(ctx, cfg, filterSubmittedEvent)

			Convey("The error returned is the error returned from the observation store", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedError)
			})
		})
	})
}

func TestExportHandler_Handle_FileStoreError(t *testing.T) {
	Convey("Given a handler with a mocked file store that returns an error", t, func() {
		expectedError := errors.New("something bad happened in the database")

		datasetAPIMock := &eventtest.DatasetAPIMock{}
		datasetAPIMock.GetVersionFunc = func(context.Context, string, string, string, string, string, string, string) (dataset.Version, error) {
			return associatedDataset, nil
		}

		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(ctx context.Context, filterJobId string) (*filterCli.Model, error) {
				return filter, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(ctx context.Context, reader io.Reader, filename string, isPublished bool) (string, error) {
				return "", expectedError
			},
		}

		originalFunc := event.SortFilter
		defer func() {
			event.SortFilter = originalFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, datasetAPIMock, cfg)

		Convey("When handle is called", func() {

			err := handler.Handle(ctx, cfg, filterSubmittedEvent)

			Convey("The error returned is the error returned from the file store", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedError)
			})
		})
	})
}

func TestExportHandler_Handle_Empty_Results(t *testing.T) {
	Convey("Given a handler with a filter query job returns no results", t, func() {

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string, string, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(ctx context.Context, filterJobId string) (*filterCli.Model, error) {
				return filter, nil
			},
			PutStateAsEmptyFunc: func(ctx context.Context, filterJobID string) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(ctx context.Context, reader io.Reader, filter string, isPublished bool) (string, error) {
				return "", observation.ErrNoResultsFound
			},
		}

		originalFunc := event.SortFilter
		defer func() {
			event.SortFilter = originalFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, datasetAPIMock, cfg)

		Convey("When handle is called", func() {
			err := handler.Handle(ctx, cfg, filterSubmittedEvent)

			Convey("The filter state is updated to empty", func() {
				So(err, ShouldNotBeNil)
				So(mockFilterStore.PutStateAsEmptyCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestExportHandler_Handle_Instance_Not_Found(t *testing.T) {
	Convey("Given a handler with a filter query job returns no results", t, func() {

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string, string, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(ctx context.Context, filterJobId string) (*filterCli.Model, error) {
				return filter, nil
			},
			PutStateAsErrorFunc: func(ctx context.Context, filterJobID string) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(ctx context.Context, reader io.Reader, filter string, isPublished bool) (string, error) {
				return "", observation.ErrNoInstanceFound
			},
		}

		originalFunc := event.SortFilter
		defer func() {
			event.SortFilter = originalFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, datasetAPIMock, cfg)

		Convey("When handle is called", func() {
			err := handler.Handle(ctx, cfg, filterSubmittedEvent)

			Convey("The filter state is updated to empty", func() {
				So(err, ShouldNotBeNil)
				So(1, ShouldEqual, len(mockFilterStore.PutStateAsErrorCalls()))
			})
		})
	})
}

func TestExportHandler_Handle_FilterStorePutError(t *testing.T) {
	Convey("Given a handler with a mocked file store that returns an error", t, func() {

		expectedError := errors.New("something bad happened in put CSV data")

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string, string, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(ctx context.Context, filterJobId string) (*filterCli.Model, error) {
				return filter, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(ctx context.Context, reader io.Reader, filename string, isPublished bool) (string, error) {
				return "", expectedError
			},
		}

		originalFunc := event.SortFilter
		defer func() {
			event.SortFilter = originalFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, datasetAPIMock, cfg)

		Convey("When handle is called", func() {
			err := handler.Handle(ctx, cfg, filterSubmittedEvent)

			Convey("The error returned is the error returned from the file store", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedError)
			})
		})
	})
}

func TestExportHandler_Handle_EventProducerError(t *testing.T) {
	Convey("Given a handler with a mocked event producer that returns an error", t, func() {

		expectedError := errors.New("something bad happened in event producer")

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string, string, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(ctx context.Context, filterJobId string) (*filterCli.Model, error) {
				return filter, nil
			},
			PutCSVDataFunc: func(ctx context.Context, filterJobID string, csv filterCli.Download) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(ctx context.Context, reader io.Reader, filename string, isPublished bool) (string, error) {
				return fileHRef, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(ctx context.Context, e *event.CSVExported) error {
				return expectedError
			},
		}

		originalFunc := event.SortFilter
		defer func() {
			event.SortFilter = originalFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer, datasetAPIMock, cfg)

		Convey("When handle is called", func() {
			err := handler.Handle(ctx, cfg, filterSubmittedEvent)

			Convey("The error returned is the error returned from the file store", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(expectedError, "error while attempting to send csv exported event to producer").Error())
			})
		})
	})
}

func TestExportHandler_Handle_Filter(t *testing.T) {
	Convey("Given a handler with a mocked dependencies", t, func() {

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string, string, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(ctx context.Context, filterJobId string) (*filterCli.Model, error) {
				return filter, nil
			},
			PutCSVDataFunc: func(ctx context.Context, filterJobID string, csv filterCli.Download) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(ctx context.Context, reader io.Reader, filename string, isPublished bool) (string, error) {
				return fileHRef, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(ctx context.Context, e *event.CSVExported) error {
				return nil
			},
		}

		originalFunc := event.SortFilter
		defer func() {
			event.SortFilter = originalFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer, datasetAPIMock, cfg)

		Convey("When handle is called", func() {

			err := handler.Handle(ctx, cfg, filterSubmittedEvent)

			Convey("The error returned is nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("The filter store is called with the correct filter job ID", func() {

				So(mockFilterStore.GetFilterCalls(), ShouldHaveLength, 2)

				actualFilterID := mockFilterStore.GetFilterCalls()[0].FilterID
				So(actualFilterID, ShouldEqual, filterSubmittedEvent.FilterID)
			})

			Convey("The observation store is called with the filter returned from the filter store", func() {

				So(mockObservationStore.StreamCSVRowsCalls(), ShouldHaveLength, 1)

				actualFilter := mockObservationStore.StreamCSVRowsCalls()[0]
				So(actualFilter.FilterID, ShouldEqual, filter.FilterID)
				So(actualFilter.Filters.Dimensions[0].Name, ShouldEqual, filter.Dimensions[0].Name)
				So(actualFilter.Filters.Dimensions[0].Options[0], ShouldEqual, filter.Dimensions[0].Options[0])
				So(actualFilter.Filters.Dimensions[0].Options[1], ShouldEqual, filter.Dimensions[0].Options[1])
				So(actualFilter.Filters.Dimensions[1].Name, ShouldEqual, filter.Dimensions[1].Name)
				So(actualFilter.Filters.Dimensions[1].Options[0], ShouldEqual, filter.Dimensions[1].Options[0])
				So(actualFilter.Filters.Dimensions[1].Options[1], ShouldEqual, filter.Dimensions[1].Options[1])
			})

			Convey("The file store is called with the reader returned from the observation store.", func() {

				So(mockedFileStore.PutFileCalls(), ShouldHaveLength, 1)

				actual := mockedFileStore.PutFileCalls()[0].Reader
				So(actual, ShouldNotBeNil)
			})

			Convey("The file store is called with a filename containing the configured prefix.", func() {

				So(mockedFileStore.PutFileCalls()[0].Filename, ShouldStartWith, filteredDatasetFilePrefix)
			})

			Convey("The filter store is called with file FileURL returned from the file store.", func() {

				So(mockFilterStore.PutCSVDataCalls(), ShouldHaveLength, 1)

				So(mockFilterStore.PutCSVDataCalls()[0].FilterID, ShouldEqual, filterSubmittedEvent.FilterID)
				So(mockFilterStore.PutCSVDataCalls()[0].DownloadItem.URL, ShouldEqual, downloadServiceURL+"/downloads/filter-outputs/"+filterOutputId+".csv")
				So(mockFilterStore.PutCSVDataCalls()[0].DownloadItem.Private, ShouldEqual, fileHRef)
				So(mockFilterStore.PutCSVDataCalls()[0].DownloadItem.Public, ShouldBeEmpty)
			})

			Convey("The event producer is called with the filter ID.", func() {

				So(mockedEventProducer.CSVExportedCalls(), ShouldHaveLength, 1)
				So(mockedEventProducer.CSVExportedCalls()[0].E.FilterID, ShouldEqual, filterSubmittedEvent.FilterID)
			})

			Convey("The CSV row reader returned from the DB is closed.", func() {
				So(len(mockRowReader.CloseCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestExportHandler_Handle_FullFileDownload(t *testing.T) {
	Convey("Given a handler with a mocked dependencies", t, func() {
		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string, string, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetMetadataURLFunc: func(string, string, string) string {
				return "/metadata"
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
			PutVersionFunc: func(context.Context, string, string, string, string, string, string, dataset.Version) error {
				return nil
			},
		}

		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{}

		mockObservationStore := &eventtest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(ctx context.Context, reader io.Reader, filename string, isPublished bool) (string, error) {
				return fileHRef, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(ctx context.Context, e *event.CSVExported) error {
				return nil
			},
		}

		originalSortFunc := event.SortFilter
		originalFilterFunc := event.CreateFilterForAll
		defer func() {
			event.SortFilter = originalSortFunc
			event.CreateFilterForAll = originalFilterFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}
		event.CreateFilterForAll = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, isPublished bool) (*observation.DimensionFilters, error) {
			return &observation.DimensionFilters{}, nil
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer, datasetAPIMock, cfg)

		Convey("When handle is called", func() {
			err := handler.Handle(ctx, cfg, fullFileDownloadSubmittedEvent)

			Convey("The error returned is nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("The dataset API is called with the correct version parameters", func() {
				So(datasetAPIMock.GetVersionCalls(), ShouldHaveLength, 1)

				getVersionCall := datasetAPIMock.GetVersionCalls()[0]

				So(getVersionCall.DatasetID, ShouldEqual, fullFileDownloadSubmittedEvent.DatasetID)
				So(getVersionCall.Edition, ShouldEqual, fullFileDownloadSubmittedEvent.Edition)
				So(getVersionCall.Version, ShouldEqual, fullFileDownloadSubmittedEvent.Version)

				So(datasetAPIMock.GetVersionMetadataCalls(), ShouldHaveLength, 1)
			})

			Convey("The observation store is called", func() {
				So(mockObservationStore.StreamCSVRowsCalls(), ShouldHaveLength, 1)
			})

			Convey("The file store is called with the reader returned from the observation store.", func() {
				So(mockedFileStore.PutFileCalls(), ShouldHaveLength, 2)

				actual := mockedFileStore.PutFileCalls()[0].Reader
				So(actual, ShouldNotBeNil)
			})

			Convey("The file store is called with a filename containing the configured prefix.", func() {
				So(mockedFileStore.PutFileCalls()[0].Filename, ShouldStartWith, fullDatasetFilePrefix)
			})

			Convey("The filter store is called with file FileURL returned from the file store.", func() {
				So(datasetAPIMock.PutVersionCalls(), ShouldHaveLength, 1)

				So(datasetAPIMock.PutVersionCalls()[0].DatasetID, ShouldEqual, fullFileDownloadSubmittedEvent.DatasetID)
				So(datasetAPIMock.PutVersionCalls()[0].Edition, ShouldEqual, fullFileDownloadSubmittedEvent.Edition)
				So(datasetAPIMock.PutVersionCalls()[0].Version, ShouldEqual, fullFileDownloadSubmittedEvent.Version)
			})

			Convey("The event producer is called with the filter ID.", func() {
				So(mockedEventProducer.CSVExportedCalls(), ShouldHaveLength, 1)
				So(mockedEventProducer.CSVExportedCalls()[0].E.DatasetID, ShouldEqual, fullFileDownloadSubmittedEvent.DatasetID)
				So(mockedEventProducer.CSVExportedCalls()[0].E.Edition, ShouldEqual, fullFileDownloadSubmittedEvent.Edition)
				So(mockedEventProducer.CSVExportedCalls()[0].E.Version, ShouldEqual, fullFileDownloadSubmittedEvent.Version)
			})

			Convey("The CSV row reader returned from the DB is closed.", func() {
				So(mockRowReader.CloseCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestExportHandler_HandlePrePublish(t *testing.T) {
	mocks := func() (*eventtest.ObservationStoreMock, *eventtest.FilterStoreMock, *eventtest.FileStoreMock, *eventtest.ProducerMock, *eventtest.DatasetAPIMock) {
		return &eventtest.ObservationStoreMock{},
			&eventtest.FilterStoreMock{},
			&eventtest.FileStoreMock{},
			&eventtest.ProducerMock{},
			&eventtest.DatasetAPIMock{}
	}

	datasetID := "111"
	instanceID := "222"
	edition := "333"
	version := "444"
	mockErr := errors.New("i am error")

	e := &event.FilterSubmitted{
		FilterID:   "",
		DatasetID:  datasetID,
		InstanceID: instanceID,
		Edition:    edition,
		Version:    version,
	}

	Convey("given observation store get csv rows returns an error", t, func() {
		observationStoreMock, filterStoreMock, fileStockMock, producerMock, datasetApiMock := mocks()
		observationStoreMock.StreamCSVRowsFunc = func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
			return nil, mockErr
		}
		datasetApiMock.GetInstanceFunc = func(context.Context, string, string, string, string, string) (dataset.Instance, string, error) {
			return dataset.Instance{Version: associatedDataset}, "", nil
		}
		datasetApiMock.GetMetadataURLFunc = func(string, string, string) string {
			return "/metadata"
		}

		datasetApiMock.GetVersionMetadataFunc = func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
			return metadata, nil
		}

		originalSortFunc := event.SortFilter
		originalFilterFunc := event.CreateFilterForAll
		defer func() {
			event.SortFilter = originalSortFunc
			event.CreateFilterForAll = originalFilterFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}
		event.CreateFilterForAll = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, isPublished bool) (*observation.DimensionFilters, error) {
			return &observation.DimensionFilters{}, nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, cfg)

		Convey("when handle is called", func() {
			err := handler.Handle(ctx, cfg, e)

			Convey("then the expected error is returned", func() {
				So(err, ShouldResemble, mockErr)
			})

			Convey("and only the expected calls are made", func() {
				So(observationStoreMock.StreamCSVRowsCalls(), ShouldHaveLength, 1)
				So(fileStockMock.PutFileCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetVersionCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetVersionMetadataCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetInstanceCalls(), ShouldHaveLength, 1)
				So(datasetApiMock.PutVersionCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("given filestore put file returns an error", t, func() {
		observationStoreMock, filterStoreMock, fileStockMock, producerMock, datasetApiMock := mocks()

		datasetApiMock.GetInstanceFunc = func(context.Context, string, string, string, string, string) (dataset.Instance, string, error) {
			return dataset.Instance{Version: associatedDataset}, "", nil
		}

		csvRowReaderMock := &observationtest.StreamRowReaderMock{
			CloseFunc: func(context.Context) error {
				return nil
			},
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
		}

		observationStoreMock.StreamCSVRowsFunc = func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
			return csvRowReaderMock, nil
		}

		fileStockMock.PutFileFunc = func(ctx context.Context, reader io.Reader, filename string, isPublished bool) (string, error) {
			return "", mockErr
		}

		originalSortFunc := event.SortFilter
		originalFilterFunc := event.CreateFilterForAll
		defer func() {
			event.SortFilter = originalSortFunc
			event.CreateFilterForAll = originalFilterFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}
		event.CreateFilterForAll = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, isPublished bool) (*observation.DimensionFilters, error) {
			return &observation.DimensionFilters{}, nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, cfg)

		Convey("when handle is called", func() {
			err := handler.Handle(ctx, cfg, e)

			Convey("then the expected error is returned", func() {
				So(err, ShouldResemble, mockErr)
			})

			Convey("and only the expected calls are made", func() {
				So(observationStoreMock.StreamCSVRowsCalls(), ShouldHaveLength, 1)
				So(observationStoreMock.StreamCSVRowsCalls()[0].InstanceID, ShouldEqual, instanceID)

				So(fileStockMock.PutFileCalls(), ShouldHaveLength, 1)
				So(fileStockMock.PutFileCalls()[0].Reader, ShouldNotBeNil)

				So(datasetApiMock.GetVersionCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetInstanceCalls(), ShouldHaveLength, 1)
				So(datasetApiMock.GetVersionMetadataCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.PutVersionCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("given there are no errors", t, func() {
		observationStoreMock, filterStoreMock, fileStockMock, producerMock, datasetApiMock := mocks()

		datasetApiMock.GetInstanceFunc = func(context.Context, string, string, string, string, string) (dataset.Instance, string, error) {
			return dataset.Instance{Version: associatedDataset}, "", nil
		}

		datasetApiMock.GetMetadataURLFunc = func(string, string, string) string {
			return "/metadata"
		}

		datasetApiMock.GetVersionMetadataFunc = func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
			return metadata, nil
		}

		csvRowReaderMock := &observationtest.StreamRowReaderMock{
			CloseFunc: func(context.Context) error {
				return nil
			},
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
		}

		observationStoreMock.StreamCSVRowsFunc = func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
			return csvRowReaderMock, nil
		}

		fileStockMock.PutFileFunc = func(ctx context.Context, reader io.Reader, filename string, isPublished bool) (string, error) {
			return "/url", nil
		}

		datasetApiMock.PutVersionFunc = func(
			ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version string, m dataset.Version) error {
			return nil
		}

		producerMock.CSVExportedFunc = func(ctx context.Context, e *event.CSVExported) error {
			return nil
		}

		originalSortFunc := event.SortFilter
		originalFilterFunc := event.CreateFilterForAll
		defer func() {
			event.SortFilter = originalSortFunc
			event.CreateFilterForAll = originalFilterFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}
		event.CreateFilterForAll = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, isPublished bool) (*observation.DimensionFilters, error) {
			return &observation.DimensionFilters{}, nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, cfg)

		Convey("when handle is called", func() {
			err := handler.Handle(ctx, cfg, e)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and the expected calls are made with the expected ", func() {
				So(observationStoreMock.StreamCSVRowsCalls(), ShouldHaveLength, 1)
				So(observationStoreMock.StreamCSVRowsCalls()[0].InstanceID, ShouldEqual, instanceID)

				So(fileStockMock.PutFileCalls(), ShouldHaveLength, 2)
				So(fileStockMock.PutFileCalls()[0].Reader, ShouldNotBeNil)

				So(datasetApiMock.PutVersionCalls(), ShouldHaveLength, 1)
				So(datasetApiMock.GetVersionCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetVersionMetadataCalls(), ShouldHaveLength, 1)
				So(datasetApiMock.GetInstanceCalls(), ShouldHaveLength, 1)

				So(datasetApiMock.PutVersionCalls(), ShouldHaveLength, 1)
				So(datasetApiMock.PutVersionCalls()[0].M.Downloads["CSV"], ShouldResemble, dataset.Download{
					Size:    "0",
					URL:     downloadServiceURL + "/downloads/datasets/111/editions/333/versions/444.csv",
					Private: "/url",
				})
			})
		})
	})

	Convey("given datasetapi.getinstance returns an error", t, func() {
		observationStoreMock, filterStoreMock, fileStockMock, producerMock, datasetApiMock := mocks()

		datasetApiMock.GetInstanceFunc = func(context.Context, string, string, string, string, string) (dataset.Instance, string, error) {
			return dataset.Instance{}, "", errors.New("dataset instances error")
		}

		csvRowReaderMock := &observationtest.StreamRowReaderMock{
			CloseFunc: func(context.Context) error {
				return nil
			},
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
		}

		observationStoreMock.StreamCSVRowsFunc = func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
			return csvRowReaderMock, nil
		}

		originalSortFunc := event.SortFilter
		originalFilterFunc := event.CreateFilterForAll
		defer func() {
			event.SortFilter = originalSortFunc
			event.CreateFilterForAll = originalFilterFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}
		event.CreateFilterForAll = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, isPublished bool) (*observation.DimensionFilters, error) {
			return &observation.DimensionFilters{}, nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, cfg)

		Convey("when handle is called", func() {
			err := handler.Handle(ctx, cfg, e)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, "dataset instances error")
			})

			Convey("and the expected calls are made", func() {
				So(observationStoreMock.StreamCSVRowsCalls(), ShouldHaveLength, 0)

				So(fileStockMock.PutFileCalls(), ShouldHaveLength, 0)

				So(datasetApiMock.PutVersionCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetVersionCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetVersionMetadataCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetInstanceCalls(), ShouldHaveLength, 1)
				So(datasetApiMock.PutVersionCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("given datasetapi.putversion returns an error", t, func() {
		observationStoreMock, filterStoreMock, fileStockMock, producerMock, datasetApiMock := mocks()

		datasetApiMock.GetInstanceFunc = func(context.Context, string, string, string, string, string) (dataset.Instance, string, error) {
			return dataset.Instance{Version: associatedDataset}, "", nil
		}

		datasetApiMock.GetMetadataURLFunc = func(string, string, string) string {
			return "/metadata"
		}

		datasetApiMock.GetVersionMetadataFunc = func(context.Context, string, string, string, string, string, string) (dataset.Metadata, error) {
			return metadata, nil
		}

		csvRowReaderMock := &observationtest.StreamRowReaderMock{
			CloseFunc: func(context.Context) error {
				return nil
			},
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
		}

		observationStoreMock.StreamCSVRowsFunc = func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (reader observation.StreamRowReader, err error) {
			return csvRowReaderMock, nil
		}

		fileStockMock.PutFileFunc = func(ctx context.Context, reader io.Reader, filename string, isPublished bool) (string, error) {
			return "/url", nil
		}

		datasetApiMock.PutVersionFunc = func(
			ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version string, m dataset.Version) error {
			return mockErr
		}

		producerMock.CSVExportedFunc = func(ctx context.Context, e *event.CSVExported) error {
			return nil
		}

		originalSortFunc := event.SortFilter
		originalFilterFunc := event.CreateFilterForAll
		defer func() {
			event.SortFilter = originalSortFunc
			event.CreateFilterForAll = originalFilterFunc
		}()
		event.SortFilter = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}
		event.CreateFilterForAll = func(ctx context.Context, handler *event.ExportHandler, event *event.FilterSubmitted, isPublished bool) (*observation.DimensionFilters, error) {
			return &observation.DimensionFilters{}, nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, cfg)

		Convey("when handle is called", func() {
			err := handler.Handle(ctx, cfg, e)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "error while attempting update version downloads").Error())
			})

			Convey("and the expected calls are made with the expected ", func() {
				So(observationStoreMock.StreamCSVRowsCalls(), ShouldHaveLength, 1)
				So(observationStoreMock.StreamCSVRowsCalls()[0].InstanceID, ShouldEqual, instanceID)

				So(fileStockMock.PutFileCalls(), ShouldHaveLength, 2)
				So(fileStockMock.PutFileCalls()[0].Reader, ShouldNotBeNil)

				So(datasetApiMock.PutVersionCalls(), ShouldHaveLength, 1)
				So(datasetApiMock.GetVersionCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetInstanceCalls(), ShouldHaveLength, 1)

				So(datasetApiMock.PutVersionCalls(), ShouldHaveLength, 1)
				So(datasetApiMock.PutVersionCalls()[0].M.Downloads["CSV"], ShouldResemble, dataset.Download{
					Size:    "0",
					URL:     downloadServiceURL + "/downloads/datasets/111/editions/333/versions/444.csv",
					Private: "/url",
				})
			})
		})
	})
}

func TestSortFilter(t *testing.T) {
	eventFilterSubmitted := event.FilterSubmitted{
		FilterID:   "whatever",
		InstanceID: "460b5039-bb09-4038-b8eb-9091713f4497",
		DatasetID:  "older-people-economic-activity",
		Edition:    "time-series",
		Version:    "1",
	}

	var pub = false

	// The following test is to code cover the first return in SortFilter
	Convey("Given a dimension of one, with a mock GetOptions", t, func() {
		var dbFilter = observation.DimensionFilters{
			Dimensions: []*observation.Dimension{
				{
					Name:    "economicactivity",
					Options: []string{"economic-activity", "employment-rate"},
				},
			},
			Published: &pub,
		}

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetOptionsFunc: func(context.Context, string, string, string, string, string, string, string, *dataset.QueryParams) (dataset.Options, error) {
				return dataset.Options{}, nil
			},
		}

		handler := event.NewExportHandler(nil, nil, nil, nil, datasetAPIMock, cfg)

		Convey("When SortFilter is called", func() {
			event.SortFilter(ctx, handler, &eventFilterSubmitted, &dbFilter)

			Convey("The dimension sees no change", func() {
				So(len(datasetAPIMock.GetOptionsCalls()), ShouldEqual, 0)
				So(dbFilter.Dimensions[0].Name, ShouldEqual, "economicactivity")
			})
		})
	})

	Convey("Given a dimension of three, with a mock GetOptions that returns nil simulating error getting record from mongo", t, func() {
		var dbFilter = observation.DimensionFilters{
			Dimensions: []*observation.Dimension{
				{
					Name:    "economicactivity",
					Options: []string{"economic-activity", "employment-rate"},
				},
				{
					Name:    "geography",
					Options: []string{"W92000004"},
				},
				{
					Name:    "sex",
					Options: []string{"people", "men"},
				},
			},
			Published: &pub,
		}

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetOptionsFunc: func(context.Context, string, string, string, string, string, string, string, *dataset.QueryParams) (dataset.Options, error) {
				return dataset.Options{}, errors.New("can't find record")
			},
		}

		handler := event.NewExportHandler(nil, nil, nil, nil, datasetAPIMock, cfg)

		Convey("When SortFilter is called", func() {
			event.SortFilter(ctx, handler, &eventFilterSubmitted, &dbFilter)

			Convey("The dimension order puts 'geogrphy' first and the rest retain their order", func() {
				// NOTE: As we are simulating mongo errors, depending on how fast the loop in SortFilter
				// manages to run all 3 go routines for the 3 dimensions, sometimes the 1st go routine
				// launched may return the expected error from simulated mongo and exit the loop before
				// the 3rd go routine runs ...
				// which means we can not check the value of GetOptionCalls() as elsewhere as it might
				// return 2, or it might return 3
				// So(len(datasetAPIMock.GetOptionsCalls()), ShouldEqual, 3)
				So(dbFilter.Dimensions[0].Name, ShouldEqual, "geography")
				So(dbFilter.Dimensions[1].Name, ShouldEqual, "economicactivity")
				So(dbFilter.Dimensions[2].Name, ShouldEqual, "sex")
			})
		})
	})

	Convey("Given a dimension of three, with a mock GetOptions that returns size of a Dimension", t, func() {
		var dbFilter = observation.DimensionFilters{
			Dimensions: []*observation.Dimension{
				{
					Name:    "economicactivity",
					Options: []string{"economic-activity", "employment-rate"},
				},
				{
					Name:    "geography",
					Options: []string{"W92000004"},
				},
				{
					Name:    "sex",
					Options: []string{"people", "men"},
				},
			},
			Published: &pub,
		}

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetOptionsFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, q *dataset.QueryParams) (dataset.Options, error) {
				switch dimension {
				case "economicactivity":
					return dataset.Options{TotalCount: 2}, nil // smallest
				case "geography":
					return dataset.Options{TotalCount: 383}, nil // largest
				case "sex":
					return dataset.Options{TotalCount: 3}, nil // in the middle
				}
				return dataset.Options{}, errors.New("can't find record")
			},
		}

		handler := event.NewExportHandler(nil, nil, nil, nil, datasetAPIMock, cfg)

		Convey("When SortFilter is called", func() {
			event.SortFilter(ctx, handler, &eventFilterSubmitted, &dbFilter)

			Convey("The dimension order is returned by largest dimension first to smallest last order", func() {
				So(len(datasetAPIMock.GetOptionsCalls()), ShouldEqual, 3)
				So(dbFilter.Dimensions[0].Name, ShouldEqual, "geography")        // largest first
				So(dbFilter.Dimensions[1].Name, ShouldEqual, "sex")              // in the middle
				So(dbFilter.Dimensions[2].Name, ShouldEqual, "economicactivity") // smallest last
			})
		})
	})
}

func TestCreateFilterForAll(t *testing.T) {
	eventFilterSubmitted := event.FilterSubmitted{
		FilterID:   "whatever",
		InstanceID: "460b5039-bb09-4038-b8eb-9091713f4497",
		DatasetID:  "older-people-economic-activity",
		Edition:    "time-series",
		Version:    "1",
	}

	// The following test is to code cover the first error return in CreateFilterForAll
	Convey("With a mock GetVersionDimensions that returns nil simulating error getting record from mongo", t, func() {
		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionDimensionsFunc: func(context.Context, string, string, string, string, string, string) (dataset.VersionDimensions, error) {
				return dataset.VersionDimensions{}, errors.New("can't find record")
			},
		}

		handler := event.NewExportHandler(nil, nil, nil, nil, datasetAPIMock, cfg)

		Convey("When CreateFilterForAll is called", func() {
			_, err := event.CreateFilterForAll(ctx, handler, &eventFilterSubmitted, false)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, "can't find record")
			})
		})
	})

	// The following test is to code cover the second error return in CreateFilterForAll
	Convey("Given a dimension of one, with a mock GetOptionsInBatches that returns nil simulating error getting record from mongo (GetVersionDimensions returns a good value)", t, func() {
		var versionDimensions = dataset.VersionDimensions{
			Items: []dataset.VersionDimension{
				{
					Name: "economicactivity",
				},
			},
		}

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionDimensionsFunc: func(context.Context, string, string, string, string, string, string) (dataset.VersionDimensions, error) {
				return versionDimensions, nil
			},
			GetOptionsInBatchesFunc: func(context.Context, string, string, string, string, string, string, string, int, int) (dataset.Options, error) {
				return dataset.Options{}, errors.New("can't find record")
			},
		}

		handler := event.NewExportHandler(nil, nil, nil, nil, datasetAPIMock, cfg)

		Convey("When CreateFilterForAll is called", func() {
			_, err := event.CreateFilterForAll(ctx, handler, &eventFilterSubmitted, false)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, "can't find record")
			})
		})
	})

	// The following test is to code cover the happy path in CreateFilterForAll
	Convey("Given a dimension of one, with a good GetOptionsInBatches and good GetVersionDimensions", t, func() {
		var versionDimensions = dataset.VersionDimensions{
			Items: []dataset.VersionDimension{
				{Name: "economicactivity"}, {Name: "sex"},
			},
		}

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionDimensionsFunc: func(context.Context, string, string, string, string, string, string) (dataset.VersionDimensions, error) {
				return versionDimensions, nil
			},
			GetOptionsInBatchesFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, batchSize int, maxWorkers int) (dataset.Options, error) {
				switch dimension {
				case "economicactivity":
					return dataset.Options{
						TotalCount: 2,
						Items: []dataset.Option{
							{Option: "economic-activity"}, {Option: "employment-rate"},
						},
					}, nil
				case "sex":
					return dataset.Options{
						TotalCount: 3,
						Items: []dataset.Option{
							{Option: "people"}, {Option: "men"}, {Option: "women"},
						},
					}, nil
				}
				return dataset.Options{}, errors.New("can't find record")
			},
		}

		handler := event.NewExportHandler(nil, nil, nil, nil, datasetAPIMock, cfg)

		Convey("When CreateFilterForAll is called", func() {
			dbFilter, err := event.CreateFilterForAll(ctx, handler, &eventFilterSubmitted, false)

			Convey("The expected dimension filter is returned", func() {
				So(err, ShouldBeNil)
				So(dbFilter.Dimensions[0].Name, ShouldEqual, "economicactivity")
				So(len(dbFilter.Dimensions), ShouldEqual, 2)
				So(len(dbFilter.Dimensions[0].Options), ShouldEqual, 2)
				So(dbFilter.Dimensions[0].Options[0], ShouldEqual, "economic-activity")
				So(dbFilter.Dimensions[0].Options[1], ShouldEqual, "employment-rate")
				So(len(dbFilter.Dimensions[1].Options), ShouldEqual, 3)
				So(dbFilter.Dimensions[1].Options[0], ShouldEqual, "people")
				So(dbFilter.Dimensions[1].Options[1], ShouldEqual, "men")
				So(dbFilter.Dimensions[1].Options[2], ShouldEqual, "women")
			})
		})
	})
}
