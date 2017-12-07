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
