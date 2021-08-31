package errors

import (
	"context"

	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out errorstest/handler.go -pkg errorstest . Handler

var _ Handler = (*KafkaHandler)(nil)

// Handler is a generic interface for handling errors
type Handler interface {
	Handle(ctx context.Context, filterID string, err error)
}

// KafkaHandler provides an error handler that writes to the kafka error topic
type KafkaHandler struct {
	messageProducer chan []byte
}

// NewKafkaHandler returns a new KafkaHandler that sends error messages
func NewKafkaHandler(messageProducer chan []byte) *KafkaHandler {
	return &KafkaHandler{
		messageProducer: messageProducer,
	}
}

// Handle logs the error to the error handler via a kafka message
func (handler *KafkaHandler) Handle(ctx context.Context, filterID string, err error) {
	data := log.Data{"filter_id": filterID, "error": err.Error()}
	log.Info(ctx, "an error occurred while processing a filter job", data)

	errorStr := Event{
		FilterID:    filterID,
		EventType:   "error",
		EventMsg:    err.Error(),
		ServiceName: "dp-dataset-exporter",
	}

	errMsg, err := EventSchema.Marshal(errorStr)
	if err != nil {
		log.Error(ctx, "failed to marshall error to event-reporter", err, data)
		return
	}
	handler.messageProducer <- errMsg
}
