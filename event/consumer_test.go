package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dp-dataset-exporter/errors/errorstest"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-kafka/kafkatest"
	"github.com/ONSdigital/log.go/log"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConsume_UnmarshallError(t *testing.T) {
	Convey("Given an event consumer with an invalid schema and a valid schema", t, func() {

		// Create mock kafka consumer with upstream channel with 2 buffered messages
		mockConsumer := kafkatest.NewMessageConsumerWithChannels(
			&kafka.ConsumerGroupChannels{
				Upstream:     make(chan kafka.Message, 2),
				Errors:       make(chan error),
				Init:         make(chan struct{}),
				Closer:       make(chan struct{}),
				Closed:       make(chan struct{}),
				UpstreamDone: make(chan bool, 1),
			}, true)

		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()

		mockConsumer.Channels().Upstream <- kafkatest.NewMessage([]byte("invalid schema"), 1)
		message := kafkatest.NewMessage(marshal(*expectedEvent), 1)
		mockConsumer.Channels().Upstream <- message

		consumer := event.NewConsumer()

		Convey("When consume messages is called", func() {
			consumer.Consume(mockConsumer, mockEventHandler, nil)
			waitForMessageToBeCommitted(message)

			Convey("Only the valid event is sent to the mockEventHandler ", func() {
				So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)

				event := mockEventHandler.HandleCalls()[0].FilterSubmittedEvent
				So(event.FilterID, ShouldEqual, expectedEvent.FilterID)
			})
		})
	})
}

func TestConsume(t *testing.T) {

	Convey("Given an event consumer with a valid schema", t, func() {

		mockConsumer := kafkatest.NewMessageConsumer(true)

		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage(marshal(*expectedEvent), 1)

		mockConsumer.Channels().Upstream <- message

		consumer := event.NewConsumer()

		Convey("When consume is called", func() {
			consumer.Consume(mockConsumer, mockEventHandler, nil)
			waitForMessageToBeCommitted(message)

			Convey("A event is sent to the mockEventHandler ", func() {
				So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)

				event := mockEventHandler.HandleCalls()[0].FilterSubmittedEvent
				So(event.FilterID, ShouldEqual, expectedEvent.FilterID)
			})

			Convey("The message is committed", func() {
				So(len(message.CommitCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestConsume_HandlerError(t *testing.T) {

	Convey("Given an event consumer with a mock event handler that returns an error", t, func() {

		expectedError := errors.New("Something bad happened in the event handler.")

		mockConsumer := kafkatest.NewMessageConsumer(true)

		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
				return expectedError
			},
		}

		mockErrorHandler := &errorstest.HandlerMock{
			HandleFunc: func(instanceID string, err error) {
				// do nothing, just going to inspect the call.
			},
		}

		expectedEvent := getExampleEvent()

		message := kafkatest.NewMessage(marshal(*expectedEvent), 1)
		mockConsumer.Channels().Upstream <- message

		consumer := event.NewConsumer()

		Convey("When consume is called", func() {
			consumer.Consume(mockConsumer, mockEventHandler, mockErrorHandler)
			waitForMessageToBeCommitted(message)

			Convey("Then the error handler is given the error returned from the event handler", func() {
				So(len(mockErrorHandler.HandleCalls()), ShouldEqual, 1)
				So(mockErrorHandler.HandleCalls()[0].Err, ShouldEqual, expectedError)
				So(mockErrorHandler.HandleCalls()[0].FilterID, ShouldEqual, expectedEvent.FilterID)
			})

			Convey("and the message is not committed - to be retried", func() {
				So(len(message.CommitCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Given a consumer", t, func() {

		mockConsumer := kafkatest.NewMessageConsumer(true)

		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage(marshal(*expectedEvent), 1)

		mockConsumer.Channels().Upstream <- message

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
func marshal(event event.FilterSubmitted) []byte {
	bytes, err := schema.FilterSubmittedEvent.Marshal(event)
	So(err, ShouldBeNil)
	return bytes
}

func getExampleEvent() *event.FilterSubmitted {
	expectedEvent := &event.FilterSubmitted{
		FilterID: "123321",
	}
	return expectedEvent
}

func waitForMessageToBeCommitted(message *kafkatest.Message) {
	ctx := context.Background()
	start := time.Now()
	timeout := start.Add(time.Millisecond * 500)
	for {
		if len(message.CommitCalls()) > 0 {
			log.Event(ctx, "message has been committed", log.INFO)
			break
		}

		if time.Now().After(timeout) {
			log.Event(ctx, "timeout hit", log.INFO)
			break
		}

		time.Sleep(time.Millisecond * 10)
	}
}
