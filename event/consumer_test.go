package event_test

import (
	"context"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/dp-kafka/v4/kafkatest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var testCtx = context.Background()

var errHandler = errors.New("handler Error")

var testEvent = event.FilterSubmitted{
	FilterID: "World",
}

// TODO: remove or replace hello called logic with app specific
func TestConsume(t *testing.T) {
	Convey("Given kafka consumer and event handler mocks", t, func() {
		cgChannels := &kafka.ConsumerGroupChannels{Upstream: make(chan kafka.Message, 2)}
		mockConsumer := &kafkatest.IConsumerGroupMock{
			ChannelsFunc: func() *kafka.ConsumerGroupChannels { return cgChannels },
		}

		handlerWg := &sync.WaitGroup{}
		mockEventHandler := &eventtest.HandlerMock{
			HandleFunc: func(ctx context.Context, config *config.Config, event *event.FilterSubmitted) error {
				defer handlerWg.Done()
				return nil
			},
		}

		Convey("And a kafka message with the valid schema being sent to the Upstream channel", func() {
			message, err := kafkatest.NewMessage(marshal(testEvent), 0)
			if err != nil {
				t.Errorf("unable to create new message")
			}

			mockConsumer.Channels().Upstream <- message

			Convey("When consume message is called", func() {
				handlerWg.Add(1)
				event.Consume(testCtx, mockConsumer, mockEventHandler, &config.Config{KafkaConfig: config.KafkaConfig{NumWorkers: 1}})
				handlerWg.Wait()

				Convey("An event is sent to the mockEventHandler ", func() {
					So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
					So(*mockEventHandler.HandleCalls()[0].FilterSubmitted, ShouldResemble, testEvent)
				})

				Convey("The message is committed and the consumer is released", func() {
					<-message.UpstreamDone()
					So(len(message.CommitCalls()), ShouldEqual, 1)
					So(len(message.ReleaseCalls()), ShouldEqual, 1)
				})
			})
		})

		Convey("And two kafka messages, one with a valid schema and one with an invalid schema", func() {
			validMessage, err := kafkatest.NewMessage(marshal(testEvent), 1)
			if err != nil {
				t.Errorf("unable to create new message")
			}

			invalidMessage, err := kafkatest.NewMessage([]byte("invalid schema"), 0)
			if err != nil {
				t.Errorf("unable to create new message")
			}

			mockConsumer.Channels().Upstream <- invalidMessage
			mockConsumer.Channels().Upstream <- validMessage

			Convey("When consume messages is called", func() {
				handlerWg.Add(1)
				event.Consume(testCtx, mockConsumer, mockEventHandler, &config.Config{KafkaConfig: config.KafkaConfig{NumWorkers: 1}})
				handlerWg.Wait()

				Convey("Only the valid event is sent to the mockEventHandler ", func() {
					So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
					So(*mockEventHandler.HandleCalls()[0].FilterSubmitted, ShouldResemble, testEvent)
				})

				Convey("Only the valid message is committed, but the consumer is released for both messages", func() {
					<-validMessage.UpstreamDone()
					<-invalidMessage.UpstreamDone()
					So(len(validMessage.CommitCalls()), ShouldEqual, 1)
					So(len(invalidMessage.CommitCalls()), ShouldEqual, 1)
					So(len(validMessage.ReleaseCalls()), ShouldEqual, 1)
					So(len(invalidMessage.ReleaseCalls()), ShouldEqual, 1)
				})
			})
		})

		Convey("With a failing handler and a kafka message with the valid schema being sent to the Upstream channel", func() {
			mockEventHandler.HandleFunc = func(ctx context.Context, config *config.Config, event *event.FilterSubmitted) error {
				defer handlerWg.Done()
				return errHandler
			}

			message, err := kafkatest.NewMessage(marshal(testEvent), 0)
			if err != nil {
				t.Errorf("unable to create new message")
			}

			mockConsumer.Channels().Upstream <- message

			Convey("When consume message is called", func() {
				handlerWg.Add(1)
				event.Consume(testCtx, mockConsumer, mockEventHandler, &config.Config{KafkaConfig: config.KafkaConfig{NumWorkers: 1}})
				handlerWg.Wait()

				Convey("An event is sent to the mockEventHandler ", func() {
					So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)
					So(*mockEventHandler.HandleCalls()[0].FilterSubmitted, ShouldResemble, testEvent)
				})

				Convey("The message is committed and the consumer is released", func() {
					<-message.UpstreamDone()
					So(len(message.CommitCalls()), ShouldEqual, 1)
					So(len(message.ReleaseCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}

// marshal helper method to marshal a event into a []byte
func marshal(e event.FilterSubmitted) []byte {
	bytes, err := schema.FilterSubmittedEvent.Marshal(e)
	So(err, ShouldBeNil)
	return bytes
}

// var ctx = context.Background()

// func TestConsume_UnmarshallError(t *testing.T) {
// 	Convey("Given an event consumer with an invalid schema and a valid schema", t, func() {

// 		// Create mock kafka consumer with upstream channel with 2 buffered messages
// 		channels := kafka.CreateConsumerGroupChannels(2, 1)
// 		mockConsumer := kafkatest.IConsumerGroupMock{
// 			ChannelsFunc: func() *kafka.ConsumerGroupChannels {
// 				return channels
// 			},
// 		}

// 		cfg, err := config.Get()
// 		So(err, ShouldBeNil)

// 		mockEventHandler := &eventtest.HandlerMock{
// 			HandleFunc: func(ctx context.Context, cfg *config.Config, filterJobSubmittedEvent *event.FilterSubmitted) error {
// 				return nil
// 			},
// 		}

// 		expectedEvent := getExampleEvent()

// 		unexpectedMessage, err := kafkatest.NewMessage([]byte("invalid schema"), 1)
// 		if err != nil {
// 			t.Errorf("unable to create new message")
// 		}
// 		mockConsumer.Channels().Upstream <- unexpectedMessage

// 		expectedMessage, err := kafkatest.NewMessage(marshal(*expectedEvent), 1)
// 		if err != nil {
// 			t.Errorf("unable to create new message")
// 		}
// 		mockConsumer.Channels().Upstream <- expectedMessage

// 		consumer := event.NewConsumer(cfg.KafkaConsumerWorkers)

// 		Convey("When consume messages is called", func() {
// 			consumer.Consume(&mockConsumer, mockEventHandler, nil)
// 			waitForMessageToBeCommitted(expectedMessage)

// 			Convey("Only the valid event is sent to the mockEventHandler ", func() {
// 				So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)

// 				eventStruct := mockEventHandler.HandleCalls()[0].FilterSubmittedEvent
// 				So(eventStruct.FilterID, ShouldEqual, expectedEvent.FilterID)
// 			})
// 		})
// 	})
// }

// func TestConsume(t *testing.T) {

// 	Convey("Given an event consumer with a valid schema", t, func() {

// 		mockConsumer, err := kafkatest.NewConsumer(
// 			context.Background(),
// 			&kafka.ConsumerGroupConfig{
// 				BrokerAddrs: []string{"localhost"},
// 				Topic:       "test-topic",
// 				GroupName:   "test-group",
// 			},
// 			nil,
// 		)
// 		if err != nil {
// 			t.Errorf("unable to create consumer")
// 		}

// 		mockEventHandler := &eventtest.HandlerMock{
// 			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
// 				return nil
// 			},
// 		}

// 		expectedEvent := getExampleEvent()
// 		message, err := kafkatest.NewMessage(marshal(*expectedEvent), 1)
// 		if err != nil {
// 			t.Errorf("unable to create new message")
// 		}

// 		mockConsumer.Mock.Channels().Upstream <- message

// 		consumer := event.NewConsumer(1)

// 		Convey("When consume is called", func() {
// 			consumer.Consume(mockConsumer.Mock, mockEventHandler, nil)
// 			waitForMessageToBeCommitted(message)

// 			Convey("A event is sent to the mockEventHandler ", func() {
// 				So(len(mockEventHandler.HandleCalls()), ShouldEqual, 1)

// 				eventStruct := mockEventHandler.HandleCalls()[0].FilterSubmittedEvent
// 				So(eventStruct.FilterID, ShouldEqual, expectedEvent.FilterID)
// 			})

// 			Convey("The message is committed", func() {
// 				So(message.IsMarked(), ShouldBeTrue)
// 				So(message.IsCommitted(), ShouldBeTrue)
// 				So(len(message.CommitAndReleaseCalls()), ShouldEqual, 1)
// 			})
// 		})
// 	})
// }

// func TestConsume_HandlerError(t *testing.T) {

// 	Convey("Given an event consumer with a mock event handler that returns an error", t, func() {

// 		expectedError := errors.New("something bad happened in the event handler")

// 		mockConsumer, err := kafkatest.NewConsumer(
// 			context.Background(),
// 			&kafka.ConsumerGroupConfig{
// 				BrokerAddrs: []string{"localhost"},
// 				Topic:       "test-topic",
// 				GroupName:   "test-group",
// 			},
// 			nil,
// 		)
// 		if err != nil {
// 			t.Errorf("unable to create consumer")
// 		}

// 		mockEventHandler := &eventtest.HandlerMock{
// 			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
// 				return expectedError
// 			},
// 		}

// 		mockErrorHandler := &errorstest.HandlerMock{
// 			HandleFunc: func(ctx context.Context, instanceID string, err error) {
// 				// do nothing, just going to inspect the call.
// 			},
// 		}

// 		expectedEvent := getExampleEvent()

// 		message, err := kafkatest.NewMessage(marshal(*expectedEvent), 1)
// 		if err != nil {
// 			t.Errorf("unable to create new message")
// 		}
// 		mockConsumer.Mock.Channels().Upstream <- message

// 		consumer := event.NewConsumer(1)

// 		Convey("When consume is called", func() {
// 			consumer.Consume(mockConsumer.Mock, mockEventHandler, mockErrorHandler)
// 			waitForMessageToBeCommitted(message)

// 			Convey("Then the error handler is given the error returned from the event handler", func() {
// 				So(len(mockErrorHandler.HandleCalls()), ShouldEqual, 1)
// 				So(mockErrorHandler.HandleCalls()[0].Err, ShouldEqual, expectedError)
// 				So(mockErrorHandler.HandleCalls()[0].FilterID, ShouldEqual, expectedEvent.FilterID)
// 			})

// 			Convey("and the message is committed - we assume any retry logic has occurred within the consumer", func() {
// 				So(message.IsMarked(), ShouldBeTrue)
// 				So(message.IsCommitted(), ShouldBeTrue)
// 				So(len(message.CommitAndReleaseCalls()), ShouldEqual, 1)
// 			})
// 		})
// 	})
// }

// func TestClose(t *testing.T) {
// 	Convey("Given a consumer", t, func() {

// 		mockConsumer, err := kafkatest.NewConsumer(
// 			context.Background(),
// 			&kafka.ConsumerGroupConfig{
// 				BrokerAddrs: []string{"localhost"},
// 				Topic:       "test-topic",
// 				GroupName:   "test-group",
// 			},
// 			nil,
// 		)
// 		if err != nil {
// 			t.Errorf("unable to create consumer")
// 		}

// 		mockEventHandler := &eventtest.HandlerMock{
// 			HandleFunc: func(ctx context.Context, filterJobSubmittedEvent *event.FilterSubmitted) error {
// 				return nil
// 			},
// 		}

// 		expectedEvent := getExampleEvent()
// 		message, err := kafkatest.NewMessage(marshal(*expectedEvent), 1)
// 		if err != nil {
// 			t.Errorf("unable to create new message")
// 		}

// 		mockConsumer.Mock.Channels().Upstream <- message

// 		consumer := event.NewConsumer(1)
// 		consumer.Consume(mockConsumer.Mock, mockEventHandler, nil)
// 		Convey("When close is called", func() {

// 			err := consumer.Close(ctx)

// 			Convey("Then no errors are returned", func() {
// 				So(err, ShouldBeNil)
// 			})
// 		})
// 	})
// }

// // marshal helper method to marshal a event into a []byte
// func marshal(event event.FilterSubmitted) []byte {
// 	bytes, err := schema.FilterSubmittedEvent.Marshal(event)
// 	So(err, ShouldBeNil)
// 	return bytes
// }

// func getExampleEvent() *event.FilterSubmitted {
// 	expectedEvent := &event.FilterSubmitted{
// 		FilterID: "123321",
// 	}
// 	return expectedEvent
// }

// func waitForMessageToBeCommitted(message *kafkatest.Message) {
// 	start := time.Now()
// 	timeout := start.Add(time.Millisecond * 500)
// 	for {
// 		if len(message.CommitAndReleaseCalls()) > 0 {
// 			log.Info(ctx, "message has been committed")
// 			break
// 		}

// 		if time.Now().After(timeout) {
// 			log.Info(ctx, "timeout hit")
// 			break
// 		}

// 		time.Sleep(time.Millisecond * 10)
// 	}
// }
