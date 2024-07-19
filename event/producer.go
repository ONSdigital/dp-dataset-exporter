package event

import (
	"context"
	"errors"

	kafka "github.com/ONSdigital/dp-kafka/v4"
)

// AvroProducer of output events.
type AvroProducer struct {
	messageProducer chan kafka.BytesMessage
	marshaller      Marshaller
}

// Marshaller marshals events into messages.
type Marshaller interface {
	Marshal(s interface{}) ([]byte, error)
}

// NewAvroProducer returns a new instance of AvroProducer.
func NewAvroProducer(messageProducer chan kafka.BytesMessage, marshaller Marshaller) *AvroProducer {
	return &AvroProducer{
		messageProducer: messageProducer,
		marshaller:      marshaller,
	}
}

// CSVExported produces a new CSV exported event for the given filter output ID.
func (producer *AvroProducer) CSVExported(ctx context.Context, event *CSVExported) error {
	if event == nil {
		return errors.New("event required but was nil")
	}
	bytes, err := producer.marshaller.Marshal(event)
	if err != nil {
		return err
	}

	producer.messageProducer <- kafka.BytesMessage{Value: bytes, Context: ctx}

	return nil
}
