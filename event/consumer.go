package event

import (
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

//go:generate moq -out eventtest/handler.go -pkg eventtest . Handler

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Incoming() chan kafka.Message
	Closer() chan bool
}

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(filterJobSubmittedEvent FilterJobSubmitted) error
}

// Consume converts messages to event instances, and pass the event to the provided handler.
func Consume(messageConsumer MessageConsumer, handler Handler) {
	for message := range messageConsumer.Incoming() {

		event, err := Unmarshal(message)
		if err != nil {
			log.Error(err, log.Data{"message": "failed to unmarshal event"})
			continue
		}

		log.Debug("event received", log.Data{"event": event})

		err = handler.Handle(*event)
		if err != nil {
			log.Error(err, log.Data{"message": "failed to handle event"})
			continue
		}

		log.Debug("event processed - committing message", log.Data{"event": event})
		message.Commit()
		log.Debug("message committed", log.Data{"event": event})
	}
}

// Unmarshal converts a event instance to []byte.
func Unmarshal(message kafka.Message) (*FilterJobSubmitted, error) {
	var event FilterJobSubmitted
	err := schema.FilterJobSubmittedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}
