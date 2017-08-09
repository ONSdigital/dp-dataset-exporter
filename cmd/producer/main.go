package main

import (
	"bufio"
	"fmt"
	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/avro"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"os"
)

func main() {
	log.Namespace = "dp-observation-importer"

	config, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	// Avoid logging the neo4j URL as it may contain a password
	log.Debug("loaded config", log.Data{"config": config})

	kafkaBrokers := []string{config.KafkaAddr}

	kafkaProducer, err := kafka.NewProducer(kafkaBrokers, config.FilterJobConsumerTopic, 0)
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	avroProducer := avro.NewProducer(kafkaProducer, schema.FilterJobSubmitted)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		filterJobID := scanner.Text()
		fmt.Println("Sending filter job event with ID " + filterJobID)

		event := event.CSVExported{
			FilterJobID: filterJobID,
		}

		avroProducer.SendMessage(event)
	}
}
