package event_test

import (
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/observation"
	"github.com/ONSdigital/dp-dataset-exporter/observation/observationtest"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

const filterJobId = "345"
const fileUrl = "s3://some/url/123.csv"

var filterJobEvent = &event.FilterJobSubmitted{FilterJobID: filterJobId}

var filter = &observation.Filter{
	JobID:           filterJobId,
	DataSetFilterID: "888",
	DimensionFilters: []*observation.DimensionFilter{
		{Name: "age", Options: []*observation.DimensionOption{
			{Option: "29"},
			{Option: "30"}}},
		{
			Name: "sex", Options: []*observation.DimensionOption{
				{Option: "male"},
				{Option: "female"}},
		},
	},
}

func TestExportHandler_Handle_FilterStoreGetError(t *testing.T) {

	Convey("Given a handler with a mock filter store that returns an error", t, func() {

		mockError := errors.New("get filters failed")

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFilterFunc: func(filterJobId string) (*observation.Filter, error) {
				return nil, mockError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, nil, nil, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(filterJobEvent)

			Convey("The error from the filter store is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, mockError)
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
			GetCSVRowsFunc: func(filter *observation.Filter) (observation.CSVRowReader, error) {
				return nil, expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, nil, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(filterJobEvent)

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
			GetCSVRowsFunc: func(filter *observation.Filter) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, filter *observation.Filter) (string, error) {
				return "", expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(filterJobEvent)

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
			GetCSVRowsFunc: func(filter *observation.Filter) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, filter *observation.Filter) (string, error) {
				return "", expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, nil)

		Convey("When handle is called", func() {

			err := handler.Handle(filterJobEvent)

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
			GetCSVRowsFunc: func(filter *observation.Filter) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, filter *observation.Filter) (string, error) {
				return fileUrl, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(filterJobID string) error {
				return expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer)

		Convey("When handle is called", func() {

			err := handler.Handle(filterJobEvent)

			Convey("The error returned is the error returned from the file store", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, expectedError)
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
			GetCSVRowsFunc: func(filter *observation.Filter) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader, filter *observation.Filter) (string, error) {
				return fileUrl, nil
			},
		}

		mockedEventProducer := &eventtest.ProducerMock{
			CSVExportedFunc: func(filterJobID string) error {
				return nil
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore, mockedEventProducer)

		Convey("When handle is called", func() {

			err := handler.Handle(filterJobEvent)

			Convey("The error returned is nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("The filter store is called with the correct filter job ID", func() {

				So(len(mockFilterStore.GetFilterCalls()), ShouldEqual, 1)

				actualFilterID := mockFilterStore.GetFilterCalls()[0].FilterJobID
				So(actualFilterID, ShouldEqual, filterJobEvent.FilterJobID)
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

			Convey("The filter store is called with file URL returned from the file store.", func() {

				So(len(mockFilterStore.PutCSVDataCalls()), ShouldEqual, 1)

				So(mockFilterStore.PutCSVDataCalls()[0].FilterJobID, ShouldEqual, filterJobEvent.FilterJobID)
				So(mockFilterStore.PutCSVDataCalls()[0].CsvURL, ShouldEqual, fileUrl)
			})

			Convey("The event producer is called with the filter ID.", func() {

				So(len(mockedEventProducer.CSVExportedCalls()), ShouldEqual, 1)
				So(mockedEventProducer.CSVExportedCalls()[0].FilterJobID, ShouldEqual, filterJobEvent.FilterJobID)
			})

			Convey("The CSV row reader returned from the DB is closed.", func() {
				So(len(mockRowReader.CloseCalls()), ShouldEqual, 1)
			})
		})
	})
}
