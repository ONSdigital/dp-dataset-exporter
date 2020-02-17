package event

import (
	"context"
	errs "errors"

	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out eventtest/handler.go -pkg eventtest . Handler

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Channels() *kafka.ConsumerGroupChannels
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
func (consumer *Consumer) Consume(messageConsumer MessageConsumer, handler Handler, errorHandler errors.Handler /*, healthChangeChan chan bool*/) {

	go func() {
		defer close(consumer.closed)

		log.Event(nil, "starting to consume messages")
		for {
			select {
			case message := <-messageConsumer.Channels().Upstream:

				err := processMessage(message, handler, errorHandler)
				if err != nil {
					log.Event(nil, "failed to process message", log.Data{"err": err})
				} else {
					logData := log.Data{"message_offset": message.Offset()}
					log.Event(nil, "event processed - committing message", logData)
					messageConsumer.CommitAndRelease(message)
					log.Event(nil, "message committed", logData)
				}

			case <-consumer.closing:
				log.Event(nil, "closing event consumer loop")
				return

				// TODO stop consuming on error?
				// case healthy := <-healthChangeChan:
				// 	if !healthy {
				// 		log.Event(nil, "poor health - pausing Consume")
				// 		<-healthChangeChan
				// 		log.Event(nil, "good health - resuming Consume")
				// }
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
		log.Event(ctx, "successfully closed event consumer")
		return nil
	case <-ctx.Done():
		log.Event(ctx, "shutdown context time exceeded, skipping graceful shutdown of event consumer")
		return errs.New("Shutdown context timed out")
	}
}

func processMessage(message kafka.Message, handler Handler, errorHandler errors.Handler) error {

	// TODO need better context passing
	ctx := context.Background()

	logData := log.Data{"message_offset": message.Offset()}

	event, err := unmarshal(message)
	if err != nil {
		logData["message_error"] = "failed to unmarshal event"
		log.Event(ctx, "error processing message", logData, log.Error(err))
		// return nil here to commit message because this message will never succeed
		return nil
	}

	logData["event"] = event

	log.Event(ctx, "event received", logData)

	err = handler.Handle(ctx, event)
	if err != nil {
		errorHandler.Handle(event.FilterID, err)
		logData["message_error"] = "failed to handle event"
		log.Event(ctx, "Handle error", logData, log.Error(err))
	}

	return err
}

// unmarshal converts a event instance to []byte.
func unmarshal(message kafka.Message) (*FilterSubmitted, error) {
	var event FilterSubmitted
	err := schema.FilterSubmittedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}
