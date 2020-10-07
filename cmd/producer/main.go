package main

import (
	"bufio"
	"context"
	"os"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/log"
)

func main() {
	ctx := context.Background()
	log.Namespace = "dp-dataset-exporter"

	config, err := config.Get()
	if err != nil {
		log.Event(ctx, "error getting config", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	// Avoid logging the neo4j FileURL as it may contain a password
	log.Event(ctx, "loaded config", log.INFO, log.Data{"config": config})

	// Create Kafka Producer
	pChannels := kafka.CreateProducerChannels()
	pConfig := &kafka.ProducerConfig{KafkaVersion: &config.KafkaVersion}
	kafkaProducer, err := kafka.NewProducer(ctx, config.KafkaAddr, config.FilterConsumerTopic, pChannels, pConfig)
	if err != nil {
		log.Event(ctx, "fatal error trying to create kafka producer", log.FATAL, log.Error(err), log.Data{"topic": config.FilterConsumerTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.Channels().LogErrors(ctx, "kafka producer")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		filterID := scanner.Text()

		log.Event(ctx, "sending filter output event", log.INFO, log.Data{"filter_ouput_id": filterID})

		event := event.FilterSubmitted{
			FilterID: filterID,
		}

		bytes, err := schema.FilterSubmittedEvent.Marshal(event)
		if err != nil {
			log.Event(ctx, "filter submitted event error", log.FATAL, log.Error(err))
			os.Exit(1)
		}

		// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
		kafkaProducer.Initialise(ctx)
		kafkaProducer.Channels().Output <- bytes
	}
}
