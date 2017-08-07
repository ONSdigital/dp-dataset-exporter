package event_test

import (
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/observation"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

func TestExportHandler_Handle_FilterStoreError(t *testing.T) {

	Convey("Given a handler with a mock filter store that returns an error", t, func() {

		filterJobEvent := &event.FilterJobSubmitted{FilterJobID: "345"}

		mockError := errors.New("get filters failed")

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFiltersFunc: func(filterJobId string) (*observation.Filter, error) {
				return nil, mockError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, nil, nil)

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

		filterJobEvent := &event.FilterJobSubmitted{FilterJobID: "345"}

		filter := &observation.Filter{
			DataSetFilterID: "888",
			DimensionFilters: []observation.DimensionFilter{
				{Name: "age", Values: []string{"29", "30"}},
				{Name: "sex", Values: []string{"male", "female"}},
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFiltersFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter) (observation.CSVRowReader, error) {
				return nil, expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, nil)

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

		filterJobEvent := &event.FilterJobSubmitted{FilterJobID: "345"}

		filter := &observation.Filter{
			DataSetFilterID: "888",
			DimensionFilters: []observation.DimensionFilter{
				{Name: "age", Values: []string{"29", "30"}},
				{Name: "sex", Values: []string{"male", "female"}},
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFiltersFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter) (observation.CSVRowReader, error) {
				return nil, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader) error {
				return expectedError
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore)

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

		filterJobEvent := &event.FilterJobSubmitted{FilterJobID: "345"}

		filter := &observation.Filter{
			DataSetFilterID: "888",
			DimensionFilters: []observation.DimensionFilter{
				{Name: "age", Values: []string{"29", "30"}},
				{Name: "sex", Values: []string{"male", "female"}},
			},
		}

		mockFilterStore := &eventtest.FilterStoreMock{
			GetFiltersFunc: func(filterJobId string) (*observation.Filter, error) {
				return filter, nil
			},
		}

		mockObservationStore := &eventtest.ObservationStoreMock{
			GetCSVRowsFunc: func(filter *observation.Filter) (observation.CSVRowReader, error) {
				return nil, nil
			},
		}

		mockedFileStore := &eventtest.FileStoreMock{
			PutFileFunc: func(reader io.Reader) error {
				return nil
			},
		}

		handler := event.NewExportHandler(mockFilterStore, mockObservationStore, mockedFileStore)

		Convey("When handle is called", func() {

			err := handler.Handle(filterJobEvent)

			Convey("The error returned is nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("The filter store is called with the correct filter job ID", func() {

				So(len(mockFilterStore.GetFiltersCalls()), ShouldEqual, 1)

				actualFilterID := mockFilterStore.GetFiltersCalls()[0].FilterJobID
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
		})
	})
}
