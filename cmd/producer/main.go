package main

import (
	"bufio"
	"fmt"
	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"os"
)

func main() {
	log.Namespace = "dp-dataset-exporter"

	config, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	// Avoid logging the neo4j FileURL as it may contain a password
	log.Debug("loaded config", log.Data{"config": config})

	kafkaBrokers := config.KafkaAddr

	kafkaProducer, err := kafka.NewProducer(kafkaBrokers, config.FilterJobConsumerTopic, 0)
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		filterJobID := scanner.Text()
		fmt.Println("Sending filter job event with ID " + filterJobID)

		event := event.FilterJobSubmitted{
			FilterJobID: filterJobID,
		}

		bytes, err := schema.FilterJobSubmittedEvent.Marshal(event)
		if err != nil {
			log.Error(err, nil)
			os.Exit(1)
		}

		kafkaProducer.Output() <- bytes
	}
}
