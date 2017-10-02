package event_test

import (
	"github.com/ONSdigital/dp-dataset-exporter/errors/errorstest"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestConsume_UnmarshallError(t *testing.T) {
	Convey("Given an event consumer with an invalid schema and a valid schema", t, func() {

		messages := make(chan kafka.Message, 2)
		mockConsumer := kafkatest.NewMessageConsumer(messages)

		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(filterJobSubmittedEvent *event.FilterJobSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()

		messages <- kafkatest.NewMessage([]byte("invalid schema"))
		messages <- kafkatest.NewMessage(marshal(*expectedEvent))
		close(messages)

		Convey("When consume messages is called", func() {

			event.Consume(mockConsumer, mockEventHandler, nil)

			Convey("Only the valid event is sent to the mockEventHandler ", func() {
				So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)

				event := mockEventHandler.HandleCalls()[0].FilterJobSubmittedEvent
				So(event.FilterJobID, ShouldEqual, expectedEvent.FilterJobID)
			})
		})
	})
}

func TestConsume(t *testing.T) {

	Convey("Given an event consumer with a valid schema", t, func() {

		messages := make(chan kafka.Message, 1)
		mockConsumer := kafkatest.NewMessageConsumer(messages)
		handlerMock := &eventtest.HandlerMock{
			HandleFunc: func(filterJobSubmittedEvent *event.FilterJobSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()

		message := kafkatest.NewMessage(marshal(*expectedEvent))

		messages <- message
		close(messages)

		Convey("When consume is called", func() {

			event.Consume(mockConsumer, handlerMock, nil)

			Convey("A event is sent to the handlerMock ", func() {
				So(len(handlerMock.HandleCalls()), ShouldEqual, 1)

				event := handlerMock.HandleCalls()[0].FilterJobSubmittedEvent
				So(event.FilterJobID, ShouldEqual, expectedEvent.FilterJobID)
			})

			Convey("The message is committed", func() {
				So(message.Committed(), ShouldEqual, true)
			})
		})
	})
}

func TestConsume_HandlerError(t *testing.T) {

	Convey("Given an event consumer with a valid schema", t, func() {

		expectedError := errors.New("Something bad happened in the event handler.")

		messages := make(chan kafka.Message, 1)
		mockConsumer := kafkatest.NewMessageConsumer(messages)
		handlerMock := &eventtest.HandlerMock{
			HandleFunc: func(filterJobSubmittedEvent *event.FilterJobSubmitted) error {
				return expectedError
			},
		}

		mockErrorHandler := &errorstest.HandlerMock{
			HandleFunc: func(instanceID string, err error) {
				// do nothing, just going to inspect the call.
			},
		}

		expectedEvent := getExampleEvent()

		message := kafkatest.NewMessage(marshal(*expectedEvent))

		messages <- message
		close(messages)

		Convey("When consume is called", func() {

			event.Consume(mockConsumer, handlerMock, mockErrorHandler)

			Convey("A event is sent to the handlerMock", func() {
				So(len(handlerMock.HandleCalls()), ShouldEqual, 1)

				event := handlerMock.HandleCalls()[0].FilterJobSubmittedEvent
				So(event.FilterJobID, ShouldEqual, expectedEvent.FilterJobID)
			})

			Convey("The error handler is given the error returned from the event handler", func() {
				So(len(mockErrorHandler.HandleCalls()), ShouldEqual, 1)
				So(mockErrorHandler.HandleCalls()[0].Err, ShouldEqual, expectedError)
				So(mockErrorHandler.HandleCalls()[0].InstanceID, ShouldEqual, expectedEvent.FilterJobID)
			})

			Convey("The message is committed", func() {
				So(message.Committed(), ShouldEqual, true)
			})
		})
	})
}

// marshal helper method to marshal a event into a []byte
func marshal(event event.FilterJobSubmitted) []byte {
	bytes, err := schema.FilterJobSubmittedEvent.Marshal(event)
	So(err, ShouldBeNil)
	return bytes
}

func getExampleEvent() *event.FilterJobSubmitted {
	expectedEvent := &event.FilterJobSubmitted{
		FilterJobID: "123321",
	}
	return expectedEvent
}
