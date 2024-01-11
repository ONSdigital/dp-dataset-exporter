package event_test

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/event/eventtest"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/log.go/v2/log"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAvroProducer_CSVExported(t *testing.T) {

	Convey("Given an a mock message producer", t, func() {

		// channel to capture messages sent.
		outputChannel := make(chan kafka.BytesMessage, 1)

		avroBytes := []byte("hello world")

		marshallerMock := &eventtest.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return avroBytes, nil
			},
		}

		eventProducer := event.NewAvroProducer(outputChannel, marshallerMock)

		Convey("when CSVExported is called with a nil event", func() {
			err := eventProducer.CSVExported(context.Background(), nil)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, "event required but was nil")
			})

			Convey("and marshaller is never called", func() {
				So(len(marshallerMock.MarshalCalls()), ShouldEqual, 0)
			})
		})

		Convey("When CSVExported is called on the event producer", func() {
			csvEvent := &event.CSVExported{
				DatasetID:  "",
				InstanceID: "",
				Edition:    "",
				Version:    "",
				FileURL:    "",
				Filename:   "",
				RowCount:   0,
			}
			err := eventProducer.CSVExported(context.Background(), csvEvent)

			Convey("The expected event is available on the output channel", func() {
				log.Info(ctx, "error is:", log.Data{"error": err})
				So(err, ShouldBeNil)

				messageBytes := <-outputChannel
				close(outputChannel)
				So(messageBytes.Value, ShouldResemble, avroBytes)
			})
		})
	})
}

// Unmarshal converts observation events to []byte.
func unmarshal(bytes []byte) *event.CSVExported {
	csvEvent := &event.CSVExported{}
	err := schema.CSVExportedEvent.Unmarshal(bytes, csvEvent)
	So(err, ShouldBeNil)
	return csvEvent
}
