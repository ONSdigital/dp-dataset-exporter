package main

import (
	"bufio"
	"context"
	"os"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/log.go/v2/log"
)

func main() {
	ctx := context.Background()
	log.Namespace = "dp-dataset-exporter"

	config, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error getting config", err)
		os.Exit(1)
	}

	// Avoid logging the neo4j FileURL as it may contain a password
	log.Info(ctx, "loaded config", log.Data{"config": config})

	// Create Kafka Producer
	pConfig := &kafka.ProducerConfig{
		KafkaVersion: &config.KafkaVersion,
		Topic:        config.FilterConsumerTopic,
		BrokerAddrs:  config.KafkaAddr,
	}

	kafkaProducer, err := kafka.NewProducer(ctx, pConfig)
	if err != nil {
		log.Fatal(ctx, "fatal error trying to create kafka producer", err, log.Data{"topic": config.FilterConsumerTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.LogErrors(ctx)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		filterID := scanner.Text()

		log.Info(ctx, "sending filter output event", log.Data{"filter_ouput_id": filterID})

		event := event.FilterSubmitted{
			FilterID: filterID,
		}

		bytes, err := schema.FilterSubmittedEvent.Marshal(event)
		if err != nil {
			log.Fatal(ctx, "filter submitted event error", err)
			os.Exit(1)
		}

		// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
		kafkaProducer.Initialise(ctx)
		kafkaProducer.Channels().Output <- kafka.BytesMessage{Value: bytes, Context: ctx} 
	}
}
