package event

import (
	"context"
	errs "errors"

	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

//go:generate moq -out eventtest/handler.go -pkg eventtest . Handler

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Incoming() chan kafka.Message
	CommitAndRelease(kafka.Message)
}

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(ctx context.Context, filterSubmittedEvent *FilterSubmitted) error
}

// Consumer consumes event messages.
type Consumer struct {
	closing chan bool
	closed  chan bool
}

// NewConsumer returns a new consumer instance.
func NewConsumer() *Consumer {
	return &Consumer{
		closing: make(chan bool),
		closed:  make(chan bool),
	}
}

// Consume converts messages to event instances, and pass the event to the provided handler.
func (consumer *Consumer) Consume(messageConsumer MessageConsumer, handler Handler, errorHandler errors.Handler, healthChangeChan chan bool) {

	go func() {
		defer close(consumer.closed)

		log.Info("starting to consume messages", nil)
		for {
			select {
			case message := <-messageConsumer.Incoming():

				err := processMessage(message, handler, errorHandler)
				if err == nil {
					logData := log.Data{"message_offset": message.Offset()}
					log.Debug("event processed - committing message", logData)
					messageConsumer.CommitAndRelease(message)
					log.Debug("message committed", logData)
				}

			case <-consumer.closing:
				log.Info("closing event consumer loop", nil)
				return

			case healthy := <-healthChangeChan:
				if !healthy {
					log.Info("poor health - pausing Consume", nil)
					<-healthChangeChan
					log.Info("good health - resuming Consume", nil)
				}
			}
		}
	}()

}

// Close safely closes the consumer and releases all resources
func (consumer *Consumer) Close(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	close(consumer.closing)

	select {
	case <-consumer.closed:
		log.Info("successfully closed event consumer", nil)
		return nil
	case <-ctx.Done():
		log.Info("shutdown context time exceeded, skipping graceful shutdown of event consumer", nil)
		return errs.New("Shutdown context timed out")
	}
}

func processMessage(message kafka.Message, handler Handler, errorHandler errors.Handler) error {
	logData := log.Data{"message_offset": message.Offset()}

	event, err := unmarshal(message)
	if err != nil {
		logData["message_error"] = "failed to unmarshal event"
		log.Error(err, logData)
		// return nil here to commit message because this message will never succeed
		return nil
	}

	logData["event"] = event

	// TODO need better context passing
	ctx := context.Background()

	log.Debug("event received", logData)

	err = handler.Handle(ctx, event)
	if err != nil {
		errorHandler.Handle(event.FilterID, err)
		logData["message_error"] = "failed to handle event"
		log.Error(err, logData)
		return err
	}

	return nil
}

// unmarshal converts a event instance to []byte.
func unmarshal(message kafka.Message) (*FilterSubmitted, error) {
	var event FilterSubmitted
	err := schema.FilterSubmittedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}
