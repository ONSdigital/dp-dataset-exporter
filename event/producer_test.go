package event_test

import (
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAvroProducer_CSVExported(t *testing.T) {

	Convey("Given an a mock message producer", t, func() {

		// mock schema producer contains the output channel to capture messages sent.
		outputChannel := make(chan []byte, 1)
		mockMessageProducer := kafkatest.NewMessageProducer(outputChannel, nil, nil)

		eventProducer := event.NewAvroProducer(mockMessageProducer)

		Convey("When CSVExported is called on the event producer", func() {

			eventProducer.CSVExported(filterJobId)

			Convey("The expected event is available on the output channel", func() {

				messageBytes := <-outputChannel
				close(outputChannel)
				observationEvent := Unmarshal(messageBytes)
				So(observationEvent.FilterJobID, ShouldEqual, filterJobId)
			})
		})
	})
}

func TestMarshal(t *testing.T) {

	Convey("Given an example CSV exported event", t, func() {

		expectedEvent := event.CSVExported{
			FilterJobID: filterJobId,
		}

		Convey("When Marshal is called", func() {

			bytes, err := event.Marshal(expectedEvent)
			So(err, ShouldBeNil)

			Convey("The event can be unmarshalled and has the expected values", func() {

				actualEvent := Unmarshal(bytes)
				So(actualEvent.FilterJobID, ShouldEqual, expectedEvent.FilterJobID)
			})
		})
	})
}

// Unmarshal converts observation events to []byte.
func Unmarshal(bytes []byte) *event.CSVExported {
	event := &event.CSVExported{}
	err := schema.CSVExportedEvent.Unmarshal(bytes, event)
	So(err, ShouldBeNil)
	return event
}
