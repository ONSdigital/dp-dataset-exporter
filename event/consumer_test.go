package event_test

import (
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestConsume_UnmarshallError(t *testing.T) {
	Convey("Given an event consumer with an invalid schema and a valid schema", t, func() {

		messages := make(chan kafka.Message, 2)
		mockConsumer := kafkatest.NewMessageConsumer(messages)

		handlerMock := &eventtest.HandlerMock{
			HandleFunc: func(filterJobSubmittedEvent event.FilterJobSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()

		messages <- kafkatest.NewMessage([]byte("invalid schema"))
		messages <- kafkatest.NewMessage(Marshal(*expectedEvent))
		close(messages)

		Convey("When consume messages is called", func() {

			event.Consume(mockConsumer, handlerMock)

			Convey("Only the valid event is sent to the handlerMock ", func() {
				So(len(handlerMock.HandleCalls()), ShouldEqual, 1)

				event := handlerMock.HandleCalls()[0].FilterJobSubmittedEvent
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
			HandleFunc: func(filterJobSubmittedEvent event.FilterJobSubmitted) error {
				return nil
			},
		}

		expectedEvent := getExampleEvent()

		message := kafkatest.NewMessage(Marshal(*expectedEvent))

		messages <- message
		close(messages)

		Convey("When consume is called", func() {

			event.Consume(mockConsumer, handlerMock)

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

func TestToEvent(t *testing.T) {

	Convey("Given a event schema encoded using avro", t, func() {

		expectedEvent := getExampleEvent()
		message := kafkatest.NewMessage(Marshal(*expectedEvent))

		Convey("When the expectedEvent is unmarshalled", func() {

			event, err := event.Unmarshal(message)

			Convey("The expectedEvent has the expected values", func() {
				So(err, ShouldBeNil)
				So(event.FilterJobID, ShouldEqual, expectedEvent.FilterJobID)
			})
		})
	})
}

// Marshal helper method to marshal a event into a []byte
func Marshal(event event.FilterJobSubmitted) []byte {
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
