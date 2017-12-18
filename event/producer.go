package event

import (
	"github.com/pkg/errors"
)

//go:generate moq -out eventtest/marshaller.go -pkg eventtest . Marshaller

// AvroProducer of output events.
type AvroProducer struct {
	messageProducer MessageProducer
	marshaller      Marshaller
}

// MessageProducer dependency that writes messages.
type MessageProducer interface {
	Output() chan []byte
}

// Marshaller marshals events into messages.
type Marshaller interface {
	Marshal(s interface{}) ([]byte, error)
}

// NewAvroProducer returns a new instance of AvroProducer.
func NewAvroProducer(messageProducer MessageProducer, marshaller Marshaller) *AvroProducer {
	return &AvroProducer{
		messageProducer: messageProducer,
		marshaller:      marshaller,
	}
}

// CSVExported produces a new CSV exported event for the given filter output ID.
func (producer *AvroProducer) CSVExported(event *CSVExported) error {
	if event == nil {
		return errors.New("event required but was nil")
	}
	bytes, err := producer.marshaller.Marshal(event)
	if err != nil {
		return err
	}

	producer.messageProducer.Output() <- bytes

	return nil
}
