package event_test

import (
	"context"
	"github.com/ONSdigital/dp-dataset-exporter/errors/errorstest"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	"github.com/ONSdigital/go-ns/log"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
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

		consumer := event.NewConsumer()

		Convey("When consume messages is called", func() {

			consumer.Consume(mockConsumer, mockEventHandler, nil)
			waitForEventsToBeSentToHandler(mockEventHandler)

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
		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(filterJobSubmittedEvent *event.FilterJobSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage(marshal(*expectedEvent))

		messages <- message

		consumer := event.NewConsumer()

		Convey("When consume is called", func() {

			consumer.Consume(mockConsumer, mockEventHandler, nil)
			waitForEventsToBeSentToHandler(mockEventHandler)

			Convey("A event is sent to the mockEventHandler ", func() {
				So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)

				event := mockEventHandler.HandleCalls()[0].FilterJobSubmittedEvent
				So(event.FilterJobID, ShouldEqual, expectedEvent.FilterJobID)
			})

			Convey("The message is committed", func() {
				So(message.Committed(), ShouldEqual, true)
			})
		})
	})
}

func TestConsume_HandlerError(t *testing.T) {

	Convey("Given an event consumer with a mock event handler that returns an error", t, func() {

		expectedError := errors.New("Something bad happened in the event handler.")

		messages := make(chan kafka.Message, 1)
		mockConsumer := kafkatest.NewMessageConsumer(messages)
		mockEventHandler := &eventtest.HandlerMock{
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

		consumer := event.NewConsumer()

		Convey("When consume is called", func() {

			consumer.Consume(mockConsumer, mockEventHandler, mockErrorHandler)
			waitForEventsToBeSentToHandler(mockEventHandler)

			Convey("The error handler is given the error returned from the event handler", func() {
				So(len(mockErrorHandler.HandleCalls()), ShouldEqual, 1)
				So(mockErrorHandler.HandleCalls()[0].Err, ShouldEqual, expectedError)
				So(mockErrorHandler.HandleCalls()[0].FilterID, ShouldEqual, expectedEvent.FilterJobID)
			})

			Convey("The message is committed", func() {
				So(message.Committed(), ShouldEqual, true)
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Given a consumer", t, func() {

		messages := make(chan kafka.Message, 1)
		mockConsumer := kafkatest.NewMessageConsumer(messages)
		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(filterJobSubmittedEvent *event.FilterJobSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage(marshal(*expectedEvent))

		messages <- message

		consumer := event.NewConsumer()

		consumer.Consume(mockConsumer, mockEventHandler, nil)

		Convey("When close is called", func() {

			err := consumer.Close(context.Background())

			Convey("Then no errors are returned", func() {
				So(err, ShouldBeNil)
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

func waitForEventsToBeSentToHandler(eventHandler *eventtest.HandlerMock) {

	start := time.Now()
	timeout := start.Add(time.Millisecond * 500)
	for {
		if len(eventHandler.HandleCalls()) > 0 {
			log.Debug("events have been sent to the handler", nil)
			break
		}

		if time.Now().After(timeout) {
			log.Debug("timeout hit", nil)
			break
		}

		time.Sleep(time.Millisecond * 10)
	}
}
