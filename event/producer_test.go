package event_test

import (
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAvroProducer_CSVExported(t *testing.T) {

	Convey("Given an a mock message producer", t, func() {

		// mock schema producer contains the output channel to capture messages sent.
		outputChannel := make(chan []byte, 1)
		mockMessageProducer := kafkatest.NewMessageProducer(outputChannel, nil, nil)

		avroBytes := []byte("hello world")

		marshallerMock := &eventtest.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return avroBytes, nil
			},
		}

		eventProducer := event.NewAvroProducer(mockMessageProducer, marshallerMock)

		Convey("when CSVExported is called with a nil event", func() {
			err := eventProducer.CSVExported(nil)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, "event required but was nil")
			})

			Convey("and marshaller is never called", func() {
				So(len(marshallerMock.MarshalCalls()), ShouldEqual, 0)
			})
		})

		Convey("When CSVExported is called on the event producer", func() {
			event := &event.CSVExported{
				DatasetID:  "",
				InstanceID: "",
				Edition:    "",
				Version:    "",
				FileURL:    "",
				Filename:   "",
				RowCount:   0,
			}
			err := eventProducer.CSVExported(event)

			Convey("The expected event is available on the output channel", func() {
				log.Debug("error is:", log.Data{"error": err})
				So(err, ShouldBeNil)

				messageBytes := <-outputChannel
				close(outputChannel)
				So(messageBytes, ShouldResemble, avroBytes)
			})
		})
	})
}

// Unmarshal converts observation events to []byte.
func unmarshal(bytes []byte) *event.CSVExported {
	event := &event.CSVExported{}
	err := schema.CSVExportedEvent.Unmarshal(bytes, event)
	So(err, ShouldBeNil)
	return event
}
