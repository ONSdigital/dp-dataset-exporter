package event_test

import (
	"context"
	"io"
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/dp-filter/observation/observationtest"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	filterOutputId  = "345"
	fileHRef        = "s3://some/url/123.csv"
	publishedState  = "published"
	associatedState = "associated"
	ServiceToken    = "gravy"

	downloadServiceURL        = "http://download-service"
	fullDatasetFilePrefix     = "full-dataset"
	filteredDatasetFilePrefix = "filtered-dataset"
)

var filterSubmittedEvent = &event.FilterSubmitted{
	FilterID: filterOutputId,
}

var fullFileDownloadSubmittedEvent = &event.FilterSubmitted{
	DatasetID: "12345",
	Edition:   "2018",
	Version:   "1",
}

var filter = &observation.Filter{
	FilterID: filterOutputId,
	DimensionFilters: []*observation.DimensionFilter{
		{Name: "age", Options: []string{"29", "30"}},
		{Name: "sex", Options: []string{"male", "female"}},
	},
}

var fullDownloadFilter = &observation.Filter{
	InstanceID: "888",
}

var publishedDataset = dataset.Version{
	State: publishedState,
}

var associatedDataset = dataset.Version{
	State: associatedState,
}

var metadata = dataset.Metadata{
	Version: dataset.Version{
		Downloads: map[string]dataset.Download{
			"CSV": dataset.Download{
				URL: "/url",
			},
		},
		Dimensions: []dataset.Dimension{
			dataset.Dimension{},
			dataset.Dimension{},
		},
	},
	Model: dataset.Model{
		Publisher: &dataset.Publisher{},
	},
}

var csvBytes = []byte(`v4_0, a,b,c,
1,e,f,g,
2,h,i,j,
`)

var csvContent = string(csvBytes)

var cfg = config.Config{}

func TestExportHandler_Handle_FilterStoreGetError(t *testing.T) {
	Convey("Given a handler with a mock filter store that returns an error", t, func() {

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockError := errors.New("get filters failed")

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return nil, mockError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, nil, nil, nil, datasetAPIMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("When handle is called", func() {

			err := handler.Handle(filterSubmittedEvent)

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
			GetVersionFunc: func(context.Context, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		var mockFilterStore = &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return nil, expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, nil, nil, datasetAPIMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("When handle is called", func() {

			err := handler.Handle(filterSubmittedEvent)

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

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, fileID string, isPublished bool) (string, error) {
				return "", expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, datasetAPIMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("When handle is called", func() {

			err := handler.Handle(filterSubmittedEvent)

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
			GetVersionFunc: func(context.Context, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
			PutStateAsEmptyFunc: func(filterJobID string) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, filter string, isPublished bool) (string, error) {
				return "", observation.ErrNoResultsFound
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, datasetAPIMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("When handle is called", func() {

			err := handler.Handle(filterSubmittedEvent)

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
			GetVersionFunc: func(context.Context, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
			PutStateAsErrorFunc: func(filterJobID string) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, filter string, isPublished bool) (string, error) {
				return "", observation.ErrNoInstanceFound
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, datasetAPIMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("When handle is called", func() {

			err := handler.Handle(filterSubmittedEvent)

			Convey("The filter state is updated to empty", func() {
				So(err, ShouldNotBeNil)
				So(mockFilterStore.PutStateAsErrorCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestExportHandler_Handle_FilterStorePutError(t *testing.T) {

	Convey("Given a handler with a mocked file store that returns an error", t, func() {

		expectedError := errors.New("something bad happened in put CSV data")

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, fileID string, isPublished bool) (string, error) {
				return "", expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, datasetAPIMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("When handle is called", func() {

			err := handler.Handle(filterSubmittedEvent)

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
			GetVersionFunc: func(context.Context, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
			PutCSVDataFunc: func(filterJobID string, csv observation.DownloadItem) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, fileID string, isPublished bool) (string, error) {
				return fileHRef, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(e *event.CSVExported) error {
				return expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer, datasetAPIMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("When handle is called", func() {

			err := handler.Handle(filterSubmittedEvent)

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
			GetVersionFunc: func(context.Context, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
		}

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
			PutCSVDataFunc: func(filterJobID string, csv observation.DownloadItem) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, fileID string, isPublished bool) (string, error) {
				return fileHRef, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(e *event.CSVExported) error {
				return nil
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer, datasetAPIMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("When handle is called", func() {

			err := handler.Handle(filterSubmittedEvent)

			Convey("The error returned is nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("The filter store is called with the correct filter job ID", func() {

				So(mockFilterStore.GetFilterCalls(), ShouldHaveLength, 2)

				actualFilterID := mockFilterStore.GetFilterCalls()[0].FilterID
				So(actualFilterID, ShouldEqual, filterSubmittedEvent.FilterID)
			})

			Convey("The observation store is called with the filter returned from the filter store", func() {

				So(mockObservationStore.GetCSVRowsCalls(), ShouldHaveLength, 1)

				actualFilter := mockObservationStore.GetCSVRowsCalls()[0].Filter
				So(actualFilter, ShouldEqual, filter)
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

				So(len(mockFilterStore.PutCSVDataCalls()), ShouldEqual, 1)

				So(mockFilterStore.PutCSVDataCalls()[0].FilterID, ShouldEqual, filterSubmittedEvent.FilterID)
				So(mockFilterStore.PutCSVDataCalls()[0].DownloadItem.HRef, ShouldEqual, downloadServiceURL+"/downloads/filter-outputs/"+filterOutputId+".csv")
				So(mockFilterStore.PutCSVDataCalls()[0].DownloadItem.Private, ShouldEqual, fileHRef)
				So(mockFilterStore.PutCSVDataCalls()[0].DownloadItem.Public, ShouldBeEmpty)
			})

			Convey("The event producer is called with the filter ID.", func() {

				So(mockedEventProducer.CSVExportedCalls(), ShouldHaveLength, 1)
				So(mockedEventProducer.CSVExportedCalls()[0].E.FilterID, ShouldEqual, filterSubmittedEvent.FilterID)
			})

			Convey("The CSV row reader returned from the DB is closed.", func() {
				So(mockRowReader.CloseCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestExportHandler_Handle_FullFileDownload(t *testing.T) {

	Convey("Given a handler with a mocked dependencies", t, func() {

		datasetAPIMock := &eventtest.DatasetAPIMock{
			GetVersionFunc: func(context.Context, string, string, string) (dataset.Version, error) {
				return associatedDataset, nil
			},
			GetMetadataURLFunc: func(context.Context, string, string, string) string {
				return "/metadata"
			},
			GetVersionMetadataFunc: func(context.Context, string, string, string) (dataset.Metadata, error) {
				return metadata, nil
			},
			PutVersionFunc: func(context.Context, string, string, string, dataset.Version) error {
				return nil
			},
		}

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, fileID string, isPublished bool) (string, error) {
				return fileHRef, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(e *event.CSVExported) error {
				return nil
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer, datasetAPIMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("When handle is called", func() {

			err := handler.Handle(fullFileDownloadSubmittedEvent)

			Convey("The error returned is nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("The dataset API is called with the correct version parameters", func() {

				So(datasetAPIMock.GetVersionCalls(), ShouldHaveLength, 1)

				getVersionCall := datasetAPIMock.GetVersionCalls()[0]

				So(getVersionCall.ID, ShouldEqual, fullFileDownloadSubmittedEvent.DatasetID)
				So(getVersionCall.Edition, ShouldEqual, fullFileDownloadSubmittedEvent.Edition)
				So(getVersionCall.Version, ShouldEqual, fullFileDownloadSubmittedEvent.Version)

				So(datasetAPIMock.GetVersionMetadataCalls(), ShouldHaveLength, 1)

			})

			Convey("The observation store is called", func() {
				So(mockObservationStore.GetCSVRowsCalls(), ShouldHaveLength, 1)
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

				So(datasetAPIMock.PutVersionCalls()[0].ID, ShouldEqual, fullFileDownloadSubmittedEvent.DatasetID)
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
		observationStoreMock.GetCSVRowsFunc = func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
			return nil, mockErr
		}
		datasetApiMock.GetInstanceFunc = func(context.Context, string) (dataset.Instance, error) {
			return dataset.Instance{Version: associatedDataset}, nil
		}

		datasetApiMock.GetMetadataURLFunc = func(context.Context, string, string, string) string {
			return "/metadata"
		}

		datasetApiMock.GetVersionMetadataFunc = func(context.Context, string, string, string) (dataset.Metadata, error) {
			return metadata, nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("when handle is called", func() {
			err := handler.Handle(e)

			Convey("then the expected error is returned", func() {
				So(err, ShouldResemble, mockErr)
			})

			Convey("and only the expected calls are made", func() {
				So(observationStoreMock.GetCSVRowsCalls(), ShouldHaveLength, 1)
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

		datasetApiMock.GetInstanceFunc = func(context.Context, string) (dataset.Instance, error) {
			return dataset.Instance{Version: associatedDataset}, nil
		}

		csvRowReaderMock := &observationtest.CSVRowReaderMock{
			CloseFunc: func() error {
				return nil
			},
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
		}

		observationStoreMock.GetCSVRowsFunc = func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
			return csvRowReaderMock, nil
		}

		fileStockMock.PutFileFunc = func(reader io.Reader, fileID string, isPublished bool) (string, error) {
			return "", mockErr
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("when handle is called", func() {
			err := handler.Handle(e)

			Convey("then the expected error is returned", func() {
				So(err, ShouldResemble, mockErr)
			})

			Convey("and only the expected calls are made", func() {
				So(observationStoreMock.GetCSVRowsCalls(), ShouldHaveLength, 1)
				So(observationStoreMock.GetCSVRowsCalls()[0].Filter.InstanceID, ShouldEqual, instanceID)

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

		datasetApiMock.GetInstanceFunc = func(context.Context, string) (dataset.Instance, error) {
			return dataset.Instance{Version: associatedDataset}, nil
		}

		datasetApiMock.GetMetadataURLFunc = func(context.Context, string, string, string) string {
			return "/metadata"
		}

		datasetApiMock.GetVersionMetadataFunc = func(context.Context, string, string, string) (dataset.Metadata, error) {
			return metadata, nil
		}

		csvRowReaderMock := &observationtest.CSVRowReaderMock{
			CloseFunc: func() error {
				return nil
			},
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
		}

		observationStoreMock.GetCSVRowsFunc = func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
			return csvRowReaderMock, nil
		}

		fileStockMock.PutFileFunc = func(reader io.Reader, fileID string, isPublished bool) (string, error) {
			return "/url", nil
		}

		datasetApiMock.PutVersionFunc = func(ctx context.Context, id string, edition string, version string, m dataset.Version) error {
			return nil
		}

		producerMock.CSVExportedFunc = func(e *event.CSVExported) error {
			return nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("when handle is called", func() {
			err := handler.Handle(e)

			Convey("then no is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and the expected calls are made with the expected ", func() {
				So(observationStoreMock.GetCSVRowsCalls(), ShouldHaveLength, 1)
				So(observationStoreMock.GetCSVRowsCalls()[0].Filter.InstanceID, ShouldEqual, instanceID)

				So(fileStockMock.PutFileCalls(), ShouldHaveLength, 2)
				So(fileStockMock.PutFileCalls()[0].Reader, ShouldNotBeNil)

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

		datasetApiMock.GetInstanceFunc = func(context.Context, string) (dataset.Instance, error) {
			return dataset.Instance{}, errors.New("dataset instances error")
		}

		csvRowReaderMock := &observationtest.CSVRowReaderMock{
			CloseFunc: func() error {
				return nil
			},
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
		}

		observationStoreMock.GetCSVRowsFunc = func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
			return csvRowReaderMock, nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("when handle is called", func() {
			err := handler.Handle(e)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, "dataset instances error")
			})

			Convey("and the expected calls are made", func() {
				So(observationStoreMock.GetCSVRowsCalls(), ShouldHaveLength, 0)

				So(fileStockMock.PutFileCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetVersionCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetVersionMetadataCalls(), ShouldHaveLength, 0)
				So(datasetApiMock.GetInstanceCalls(), ShouldHaveLength, 1)
				So(datasetApiMock.PutVersionCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("given datasetapi.putversion returns an error", t, func() {
		observationStoreMock, filterStoreMock, fileStockMock, producerMock, datasetApiMock := mocks()

		datasetApiMock.GetInstanceFunc = func(context.Context, string) (dataset.Instance, error) {
			return dataset.Instance{Version: associatedDataset}, nil
		}

		datasetApiMock.GetMetadataURLFunc = func(context.Context, string, string, string) string {
			return "/metadata"
		}

		datasetApiMock.GetVersionMetadataFunc = func(context.Context, string, string, string) (dataset.Metadata, error) {
			return metadata, nil
		}

		csvRowReaderMock := &observationtest.CSVRowReaderMock{
			CloseFunc: func() error {
				return nil
			},
			ReadFunc: func() (string, error) {
				return csvContent, nil
			},
		}

		observationStoreMock.GetCSVRowsFunc = func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
			return csvRowReaderMock, nil
		}

		fileStockMock.PutFileFunc = func(reader io.Reader, fileID string, isPublished bool) (string, error) {
			return "/url", nil
		}

		datasetApiMock.PutVersionFunc = func(ctx context.Context, id string, edition string, version string, m dataset.Version) error {
			return mockErr
		}

		producerMock.CSVExportedFunc = func(e *event.CSVExported) error {
			return nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock, downloadServiceURL, fullDatasetFilePrefix, filteredDatasetFilePrefix)

		Convey("when handle is called", func() {
			err := handler.Handle(e)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "error while attempting update version downloads").Error())
			})

			Convey("and the expected calls are made with the expected ", func() {
				So(observationStoreMock.GetCSVRowsCalls(), ShouldHaveLength, 1)
				So(observationStoreMock.GetCSVRowsCalls()[0].Filter.InstanceID, ShouldEqual, instanceID)

				So(fileStockMock.PutFileCalls(), ShouldHaveLength, 2)
				So(fileStockMock.PutFileCalls()[0].Reader, ShouldNotBeNil)

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

}
