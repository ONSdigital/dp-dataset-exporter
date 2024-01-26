package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dp-dataset-exporter/errors/errorstest"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/dp-kafka/v4/kafkatest"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var ctx = context.Background()

func TestConsume_UnmarshallError(t *testing.T) {
	Convey("Given an event consumer with an invalid schema and a valid schema", t, func() {

		// Create mock kafka consumer with upstream channel with 2 buffered messages
		channels := kafka.CreateConsumerGroupChannels(2, 1)
		mockConsumer := kafkatest.IConsumerGroupMock{
			ChannelsFunc: func() *kafka.ConsumerGroupChannels {
				return channels
			},
		}

		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()

		unexpectedMessage, err := kafkatest.NewMessage([]byte("invalid schema"), 1)
		if err != nil {
			t.Errorf("unable to create new message")
		}
		mockConsumer.Channels().Upstream <- unexpectedMessage

		expectedMessage, err := kafkatest.NewMessage(marshal(*expectedEvent), 1)
		if err != nil {
			t.Errorf("unable to create new message")
		}
		mockConsumer.Channels().Upstream <- expectedMessage

		consumer := event.NewConsumer(cfg.KafkaConsumerWorkers)

		Convey("When consume messages is called", func() {
			consumer.Consume(&mockConsumer, mockEventHandler, nil)
			waitForMessageToBeCommitted(expectedMessage)

			Convey("Only the valid event is sent to the mockEventHandler ", func() {
				So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)

				eventStruct := mockEventHandler.HandleCalls()[0].FilterSubmittedEvent
				So(eventStruct.FilterID, ShouldEqual, expectedEvent.FilterID)
			})
		})
	})
}

func TestConsume(t *testing.T) {

	Convey("Given an event consumer with a valid schema", t, func() {

		mockConsumer, err := kafkatest.NewConsumer(
			context.Background(),
			&kafka.ConsumerGroupConfig{
				BrokerAddrs:  []string{"localhost"},
				Topic:        "test-topic",
				GroupName:    "test-group",
			}, 
			nil,
		)
		if err != nil {
			t.Errorf("unable to create consumer")
		}

		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()
		message, err := kafkatest.NewMessage(marshal(*expectedEvent), 1)
		if err != nil {
			t.Errorf("unable to create new message")
		}

		mockConsumer.Mock.Channels().Upstream <- message

		consumer := event.NewConsumer(1)

		Convey("When consume is called", func() {
			consumer.Consume(mockConsumer.Mock, mockEventHandler, nil)
			waitForMessageToBeCommitted(message)

			Convey("A event is sent to the mockEventHandler ", func() {
				So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)

				eventStruct := mockEventHandler.HandleCalls()[0].FilterSubmittedEvent
				So(eventStruct.FilterID, ShouldEqual, expectedEvent.FilterID)
			})

			Convey("The message is committed", func() {
				So(message.IsMarked(), ShouldBeTrue)
				So(message.IsCommitted(), ShouldBeTrue)
				So(len(message.CommitAndReleaseCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestConsume_HandlerError(t *testing.T) {

	Convey("Given an event consumer with a mock event handler that returns an error", t, func() {

		expectedError := errors.New("something bad happened in the event handler")

		mockConsumer, err := kafkatest.NewConsumer(
			context.Background(),
			&kafka.ConsumerGroupConfig{
				BrokerAddrs:  []string{"localhost"},
				Topic:        "test-topic",
				GroupName:    "test-group",
			}, 
			nil,
		)
		if err != nil {
			t.Errorf("unable to create consumer")
		}

		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
				return expectedError
			},
		}

		mockErrorHandler := &errorstest.HandlerMock{
			HandleFunc: func(ctx context.Context, instanceID string, err error) {
				// do nothing, just going to inspect the call.
			},
		}

		expectedEvent := getExampleEvent()

		message, err := kafkatest.NewMessage(marshal(*expectedEvent), 1)
		if err != nil {
			t.Errorf("unable to create new message")
		}
		mockConsumer.Mock.Channels().Upstream <- message

		consumer := event.NewConsumer(1)

		Convey("When consume is called", func() {
			consumer.Consume(mockConsumer.Mock, mockEventHandler, mockErrorHandler)
			waitForMessageToBeCommitted(message)

			Convey("Then the error handler is given the error returned from the event handler", func() {
				So(len(mockErrorHandler.HandleCalls()), ShouldEqual, 1)
				So(mockErrorHandler.HandleCalls()[0].Err, ShouldEqual, expectedError)
				So(mockErrorHandler.HandleCalls()[0].FilterID, ShouldEqual, expectedEvent.FilterID)
			})

			Convey("and the message is committed - we assume any retry logic has occurred within the consumer", func() {
				So(message.IsMarked(), ShouldBeTrue)
				So(message.IsCommitted(), ShouldBeTrue)
				So(len(message.CommitAndReleaseCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Given a consumer", t, func() {

		mockConsumer, err := kafkatest.NewConsumer(
			context.Background(),
			&kafka.ConsumerGroupConfig{
				BrokerAddrs:  []string{"localhost"},
				Topic:        "test-topic",
				GroupName:    "test-group",
			}, 
			nil,
		)
		if err != nil {
			t.Errorf("unable to create consumer")
		}

		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()
		message, err := kafkatest.NewMessage(marshal(*expectedEvent), 1)
		if err != nil {
			t.Errorf("unable to create new message")
		}

		mockConsumer.Mock.Channels().Upstream <- message

		consumer := event.NewConsumer(1)
		consumer.Consume(mockConsumer.Mock, mockEventHandler, nil)
		Convey("When close is called", func() {

			err := consumer.Close(ctx)

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
	start := time.Now()
	timeout := start.Add(time.Millisecond * 500)
	for {
		if len(message.CommitAndReleaseCalls()) > 0 {
			log.Info(ctx, "message has been committed")
			break
		}

		if time.Now().After(timeout) {
			log.Info(ctx, "timeout hit")
			break
		}

		time.Sleep(time.Millisecond * 10)
	}
}
