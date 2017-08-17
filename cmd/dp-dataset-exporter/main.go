package main

import (
	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/observation"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Namespace = "dp-dataset-exporter"
	log.Debug("Starting dataset exporter", nil)

	config, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	// Avoid logging the neo4j URL as it may contain a password
	log.Debug("loaded config", log.Data{
		"topics":         []string{config.FilterJobConsumerTopic, config.CSVExportedProducerTopic},
		"brokers":        config.KafkaAddr,
		"consumer_group": config.FilterJobConsumerGroup,
		"filter_api_url": config.FilterAPIURL,
		"aws_region":     config.AWSRegion,
		"s3_bucket_name": config.S3BucketName,
		"bind_addr":      config.BindAddr})

	kafkaBrokers := []string{config.KafkaAddr}
	kafkaConsumer, err := kafka.NewConsumerGroup(
		kafkaBrokers,
		config.FilterJobConsumerTopic,
		config.FilterJobConsumerGroup,
		kafka.OffsetNewest)

	if err != nil {
		log.Error(err, log.Data{"message": "failed to create kafka consumer"})
		os.Exit(1)
	}

	kafkaProducer, err := kafka.NewProducer(kafkaBrokers, config.CSVExportedProducerTopic, 0)
	if err != nil {
		log.Error(err, log.Data{"message": "failed to create kafka producer"})
		os.Exit(1)
	}

	dbConnection, err := bolt.NewDriver().OpenNeo(config.DatabaseAddress)

	if err != nil {
		log.Error(err, log.Data{"message": "failed to create connection to Neo4j"})
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	exit := make(chan struct{})

	httpClient := http.Client{Timeout: time.Second * 15}

	filterStore := filter.NewStore(config.FilterAPIURL, config.FilterAPIAuthToken, &httpClient)
	observationStore := observation.NewStore(dbConnection)
	fileStore := file.NewStore(config.AWSRegion, config.S3BucketName)
	eventProducer := event.NewAvroProducer(kafkaProducer)

	eventHandler := event.NewExportHandler(filterStore, observationStore, fileStore, eventProducer)

	go event.Consume(kafkaConsumer, eventHandler)

	<-signals

	close(exit)

	// gracefully dispose resources
	kafkaConsumer.Closer() <- true
	kafkaProducer.Closer() <- true

	err = dbConnection.Close()
	if err != nil {
		log.Error(err, log.Data{"message": "failed to close connection to Neo4j"})
		os.Exit(0)
	}

	log.Debug("graceful shutdown was successful", nil)
	os.Exit(0)
}
