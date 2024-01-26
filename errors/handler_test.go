package errors_test

import (
	"context"
	errs "errors"
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/errors"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	. "github.com/smartystreets/goconvey/convey"
)

var ctx = context.Background()

func TestSpec(t *testing.T) {
	Convey("Given an event handler with a mock message producer", t, func() {

		mockOutputMessageChan := make(chan kafka.BytesMessage, 1)
		errorHandler := errors.NewKafkaHandler(mockOutputMessageChan)

		Convey("When handle is called", func() {

			filterID := "234"
			expectedError := errs.New("the error text")

			errorHandler.Handle(ctx, filterID, expectedError)

			Convey("The error is sent to the message producer with the expected data", func() {

				actualMessage := <-mockOutputMessageChan
				var actualEvent errors.Event
				err := errors.EventSchema.Unmarshal(actualMessage.Value, &actualEvent)

				So(err, ShouldBeNil)
				So(actualEvent.ServiceName, ShouldEqual, "dp-dataset-exporter")
				So(actualEvent.FilterID, ShouldEqual, filterID)
				So(actualEvent.EventMsg, ShouldEqual, expectedError.Error())
				So(actualEvent.EventType, ShouldEqual, "error")
			})
		})
	})
}
