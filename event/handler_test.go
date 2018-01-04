package event_test

import (
	"io"
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/dp-filter/observation/observationtest"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const filterOutputId = "345"
const fileUrl = "s3://some/url/123.csv"

var filterOutputEvent = &event.FilterSubmitted{FilterID: filterOutputId}

var filter = &observation.Filter{
	FilterID:   filterOutputId,
	InstanceID: "888",
	DimensionFilters: []*observation.DimensionFilter{
		{Name: "age", Options: []string{"29", "30"}},
		{Name: "sex", Options: []string{"male", "female"}},
	},
}

func TestExportHandler_Handle_FilterStoreGetError(t *testing.T) {
	Convey("Given a handler with a mock filter store that returns an error", t, func() {

		datasetAPIMock := &eventtest.DatasetAPIMock{}

		mockError := errors.New("get filters failed")

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return nil, mockError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, nil, nil, nil, datasetAPIMock)

		Convey("When handle is called", func() {

			err := handler.Handle(filterOutputEvent)

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

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, nil, nil, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(filterOutputEvent)

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

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return "wut", nil
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
			PutFileFunc: func(reader io.Reader, fileID string) (string, error) {
				return "", expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(filterOutputEvent)

			Convey("The error returned is the error returned from the file store", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedError)
			})
		})
	})
}

func TestExportHandler_Handle_FilterStorePutError(t *testing.T) {

	Convey("Given a handler with a mocked file store that returns an error", t, func() {

		expectedError := errors.New("something bad happened in put CSV data")

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return "wut", nil
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
			PutFileFunc: func(reader io.Reader, fileID string) (string, error) {
				return "", expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(filterOutputEvent)

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

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return "wut", nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
			PutCSVDataFunc: func(filterJobID string, csvURL string, csvSize int64) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, fileID string) (string, error) {
				return fileUrl, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(e *event.CSVExported) error {
				return expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(filterOutputEvent)

			Convey("The error returned is the error returned from the file store", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errors.Wrap(expectedError, "error while attempting to send csv exported event to producer").Error())
			})
		})
	})
}

func TestExportHandler_Handle(t *testing.T) {

	Convey("Given a handler with a mocked dependencies", t, func() {

		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				return "wut", nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
			PutCSVDataFunc: func(filterJobID string, csvURL string, csvSize int64) error {
				return nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, fileID string) (string, error) {
				return fileUrl, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(e *event.CSVExported) error {
				return nil
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(filterOutputEvent)

			Convey("The error returned is nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("The filter store is called with the correct filter job ID", func() {

				So(len(mockFilterStore.GetFilterCalls()), ShouldEqual, 1)

				actualFilterID := mockFilterStore.GetFilterCalls()[0].FilterID
				So(actualFilterID, ShouldEqual, filterOutputEvent.FilterID)
			})

			Convey("The observation store is called with the filter returned from the filter store", func() {

				So(len(mockObservationStore.GetCSVRowsCalls()), ShouldEqual, 1)

				actualFilter := mockObservationStore.GetCSVRowsCalls()[0].Filter
				So(actualFilter, ShouldEqual, filter)
			})

			Convey("The file store is called with the reader returned from the observation store.", func() {

				So(len(mockedFileStore.PutFileCalls()), ShouldEqual, 1)

				actual := mockedFileStore.PutFileCalls()[0].Reader
				So(actual, ShouldNotBeNil)
			})

			Convey("The filter store is called with file FileURL returned from the file store.", func() {

				So(len(mockFilterStore.PutCSVDataCalls()), ShouldEqual, 1)

				So(mockFilterStore.PutCSVDataCalls()[0].FilterID, ShouldEqual, filterOutputEvent.FilterID)
				So(mockFilterStore.PutCSVDataCalls()[0].CsvURL, ShouldEqual, fileUrl)
			})

			Convey("The event producer is called with the filter ID.", func() {

				So(len(mockedEventProducer.CSVExportedCalls()), ShouldEqual, 1)
				So(mockedEventProducer.CSVExportedCalls()[0].E.FilterID, ShouldEqual, filterOutputEvent.FilterID)
			})

			Convey("The CSV row reader returned from the DB is closed.", func() {
				So(len(mockRowReader.CloseCalls()), ShouldEqual, 1)
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

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock)

		Convey("when handle is called", func() {
			err := handler.Handle(e)

			Convey("then the expected error is returned", func() {
				So(err, ShouldResemble, mockErr)
			})

			Convey("and only the expected calls are made", func() {
				So(len(observationStoreMock.GetCSVRowsCalls()), ShouldEqual, 1)
				So(len(fileStockMock.PutFileCalls()), ShouldEqual, 0)
				So(len(datasetApiMock.PutVersionCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("given filestore put file returns an error", t, func() {
		observationStoreMock, filterStoreMock, fileStockMock, producerMock, datasetApiMock := mocks()

		csvRowReaderMock := &observationtest.CSVRowReaderMock{
			CloseFunc: func() error {
				return nil
			},
		}

		observationStoreMock.GetCSVRowsFunc = func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
			return csvRowReaderMock, nil
		}

		fileStockMock.PutFileFunc = func(reader io.Reader, fileID string) (string, error) {
			return "", mockErr
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock)

		Convey("when handle is called", func() {
			err := handler.Handle(e)

			Convey("then the expected error is returned", func() {
				So(err, ShouldResemble, mockErr)
			})

			Convey("and only the expected calls are made", func() {
				So(len(observationStoreMock.GetCSVRowsCalls()), ShouldEqual, 1)
				So(observationStoreMock.GetCSVRowsCalls()[0].Filter.InstanceID, ShouldEqual, instanceID)

				So(len(fileStockMock.PutFileCalls()), ShouldEqual, 1)
				So(fileStockMock.PutFileCalls()[0].Reader, ShouldNotBeNil)

				So(len(datasetApiMock.PutVersionCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("given there are no errors", t, func() {
		observationStoreMock, filterStoreMock, fileStockMock, producerMock, datasetApiMock := mocks()

		csvRowReaderMock := &observationtest.CSVRowReaderMock{
			CloseFunc: func() error {
				return nil
			},
		}

		observationStoreMock.GetCSVRowsFunc = func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
			return csvRowReaderMock, nil
		}

		fileStockMock.PutFileFunc = func(reader io.Reader, fileID string) (string, error) {
			return "/url", nil
		}

		datasetApiMock.PutVersionFunc = func(id string, edition string, version string, m dataset.Version) error {
			return nil
		}

		producerMock.CSVExportedFunc = func(e *event.CSVExported) error {
			return nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock)

		Convey("when handle is called", func() {
			err := handler.Handle(e)

			Convey("then no is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and the expected calls are made with the expected ", func() {
				So(len(observationStoreMock.GetCSVRowsCalls()), ShouldEqual, 1)
				So(observationStoreMock.GetCSVRowsCalls()[0].Filter.InstanceID, ShouldEqual, instanceID)

				So(len(fileStockMock.PutFileCalls()), ShouldEqual, 1)
				So(fileStockMock.PutFileCalls()[0].Reader, ShouldNotBeNil)

				So(len(datasetApiMock.PutVersionCalls()), ShouldEqual, 1)
				So(datasetApiMock.PutVersionCalls()[0].M.Downloads, ShouldResemble, map[string]dataset.Download{
					"CSV": {Size: "0", URL: "/url"},
				})
			})
		})
	})

	Convey("given datasetapi.putversion returns an error", t, func() {
		observationStoreMock, filterStoreMock, fileStockMock, producerMock, datasetApiMock := mocks()

		csvRowReaderMock := &observationtest.CSVRowReaderMock{
			CloseFunc: func() error {
				return nil
			},
		}

		observationStoreMock.GetCSVRowsFunc = func(filter *observation.Filter, limit *int) (observation.CSVRowReader, error) {
			return csvRowReaderMock, nil
		}

		fileStockMock.PutFileFunc = func(reader io.Reader, fileID string) (string, error) {
			return "/url", nil
		}

		datasetApiMock.PutVersionFunc = func(id string, edition string, version string, m dataset.Version) error {
			return mockErr
		}

		producerMock.CSVExportedFunc = func(e *event.CSVExported) error {
			return nil
		}

		handler := event.NewExportHandler(filterStoreMock, observationStoreMock, fileStockMock, producerMock, datasetApiMock)

		Convey("when handle is called", func() {
			err := handler.Handle(e)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "error while attempting update version downloads").Error())
			})

			Convey("and the expected calls are made with the expected ", func() {
				So(len(observationStoreMock.GetCSVRowsCalls()), ShouldEqual, 1)
				So(observationStoreMock.GetCSVRowsCalls()[0].Filter.InstanceID, ShouldEqual, instanceID)

				So(len(fileStockMock.PutFileCalls()), ShouldEqual, 1)
				So(fileStockMock.PutFileCalls()[0].Reader, ShouldNotBeNil)

				So(len(datasetApiMock.PutVersionCalls()), ShouldEqual, 1)
				So(datasetApiMock.PutVersionCalls()[0].M.Downloads, ShouldResemble, map[string]dataset.Download{
					"CSV": {Size: "0", URL: "/url"},
				})
			})
		})
	})

}
