package event

import (
	"github.com/ONSdigital/dp-dataset-exporter/schema"
)

// AvroProducer of output events.
type AvroProducer struct {
	messageProducer MessageProducer
}

// MessageProducer dependency that writes messages.
type MessageProducer interface {
	Output() chan []byte
	Closer() chan bool
}

// NewAvroProducer returns a new instance of AvroProducer.
func NewAvroProducer(messageProducer MessageProducer) *AvroProducer {
	return &AvroProducer{
		messageProducer: messageProducer,
	}
}

// CSVExported produces a new CSV exported event for the given filter job ID.
func (producer *AvroProducer) CSVExported(filterJobID string) error {

	csvExported := CSVExported{
		FilterJobID: filterJobID,
	}

	bytes, err := Marshal(csvExported)
	if err != nil {
		return err
	}

	producer.messageProducer.Output() <- bytes

	return nil
}

// Marshal converts the given ObservationsInsertedEvent to a []byte.
func Marshal(event CSVExported) ([]byte, error) {
	bytes, err := schema.CSVExportedEvent.Marshal(event)
	return bytes, err
}
