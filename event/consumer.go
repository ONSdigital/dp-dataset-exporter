package event

import (
	"context"
	errs "errors"
	"sync"

	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out eventtest/handler.go -pkg eventtest . Handler

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Channels() *kafka.ConsumerGroupChannels
}

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(ctx context.Context, filterSubmittedEvent *FilterSubmitted) error
}

type closeEvent struct {
	ctx context.Context
}

// Consumer consumes event messages.
type Consumer struct {
	closing    chan closeEvent
	closed     chan bool
	numWorkers int
}

// NewConsumer returns a new consumer instance.
func NewConsumer(numWorkers int) *Consumer {
	if numWorkers < 1 {
		numWorkers = 1
	}
	return &Consumer{
		closing:    make(chan closeEvent),
		closed:     make(chan bool),
		numWorkers: numWorkers,
	}
}

// Consume converts messages to event instances, and pass the event to the provided handler.
func (consumer *Consumer) Consume(messageConsumer MessageConsumer, handler Handler, errorHandler errors.Handler) {

	// waitGroup for workers
	wg := &sync.WaitGroup{}

	// func to be executed by each worker in a goroutine
	workerConsume := func(workerNum int) {
		defer wg.Done()
		for {
			select {
			case message := <-messageConsumer.Channels().Upstream:
				// This context will be obtained from the kafka message in the future
				ctx := context.Background()
				logData := log.Data{"message_offset": message.Offset(), "worker": workerNum}
				err := processMessage(ctx, message, handler, errorHandler)
				if err != nil {
					log.Event(ctx, "failed to process message", log.ERROR, log.Error(err), logData)
				} else {
					log.Event(ctx, "event processed - committing message", log.INFO, logData)
				}

				message.CommitAndRelease()
				log.Event(ctx, "message committed and released", log.INFO, logData)

			case event, ok := <-consumer.closing:
				if !ok {
					return // 'closing' channel already closed
				}
				log.Event(event.ctx, "closing event consumer loop", log.INFO)
				close(consumer.closing)
				return
			}
		}
	}

	// Create the required workers to consume messages in parallel
	go func() {
		defer close(consumer.closed)
		for w := 1; w <= consumer.numWorkers; w++ {
			wg.Add(1)
			go workerConsume(w)
		}
		wg.Wait()
	}()
}

// Close safely closes the consumer and releases all resources
func (consumer *Consumer) Close(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	consumer.closing <- closeEvent{ctx: ctx}

	select {
	case <-consumer.closed:
		log.Event(ctx, "successfully closed event consumer", log.INFO)
		return nil
	case <-ctx.Done():
		log.Event(ctx, "shutdown context time exceeded, skipping graceful shutdown of event consumer", log.INFO)
		return errs.New("Shutdown context timed out")
	}
}

func processMessage(ctx context.Context, message kafka.Message, handler Handler, errorHandler errors.Handler) error {

	logData := log.Data{"message_offset": message.Offset()}

	event, err := unmarshal(message)
	if err != nil {
		logData["message_error"] = "failed to unmarshal event"
		log.Event(ctx, "error processing message", log.ERROR, logData, log.Error(err))
		// return nil here to commit message because this message will never succeed
		return nil
	}

	logData["event"] = event

	log.Event(ctx, "event received", log.INFO, logData)

	err = handler.Handle(ctx, event)
	if err != nil {
		errorHandler.Handle(ctx, event.FilterID, err)
		logData["message_error"] = "failed to handle event"
		log.Event(ctx, "handle error", log.ERROR, logData, log.Error(err))
	}

	return err
}

// unmarshal converts a event instance to []byte.
func unmarshal(message kafka.Message) (*FilterSubmitted, error) {
	var event FilterSubmitted
	err := schema.FilterSubmittedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}
