package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/neo4j"

	"github.com/ONSdigital/dp-dataset-exporter/auth"
	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/go-ns/clients/dataset"
	filterHealthCheck "github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	bolt "github.com/ONSdigital/golang-neo4j-bolt-driver"
)

func main() {
	log.Namespace = "dp-dataset-exporter"
	log.Info("Starting dataset exporter", nil)

	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	log.Info("loaded config", log.Data{"config": cfg})

	// a channel used to signal a graceful exit is required.
	errorChannel := make(chan error)

	kafkaBrokers := cfg.KafkaAddr
	kafkaConsumer, err := kafka.NewConsumerGroup(
		kafkaBrokers,
		cfg.FilterConsumerTopic,
		cfg.FilterConsumerGroup,
		kafka.OffsetNewest)
	exitIfError(err)

	kafkaProducer, err := kafka.NewProducer(kafkaBrokers, cfg.CSVExportedProducerTopic, 0)
	exitIfError(err)

	kafkaErrorProducer, err := kafka.NewProducer(cfg.KafkaAddr, cfg.ErrorProducerTopic, 0)
	exitIfError(err)

	neo4jConnPool, err := bolt.NewClosableDriverPool(cfg.DatabaseAddress, cfg.Neo4jPoolSize)
	exitIfError(err)

	// when errors occur - we send a message on an error topic.
	errorHandler := errors.NewKafkaHandler(kafkaErrorProducer)

	httpClient := http.Client{Timeout: time.Second * 15}

	filterStore := filter.NewStore(cfg.FilterAPIURL, cfg.FilterAPIAuthToken, cfg.ServiceAuthToken, &httpClient)
	observationStore := observation.NewStore(neo4jConnPool)
	fileStore := file.NewStore(cfg.AWSRegion, cfg.S3BucketName)
	eventProducer := event.NewAvroProducer(kafkaProducer, schema.CSVExportedEvent)

	datasetAPICli := dataset.New(cfg.DatasetAPIURL)
	datasetAPICli.SetInternalToken(cfg.DatasetAPIAuthToken)

	eventHandler := event.NewExportHandler(filterStore, observationStore, fileStore, eventProducer, datasetAPICli, cfg.ServiceAuthToken)

	isReady := make(chan bool)

	eventConsumer := event.NewConsumer()
	eventConsumer.Consume(kafkaConsumer, eventHandler, errorHandler, isReady)

	healthChecker := healthcheck.NewServer(
		cfg.BindAddr,
		cfg.HealthCheckInterval,
		errorChannel,
		filterHealthCheck.New(cfg.FilterAPIURL),
		neo4j.NewHealthCheckClient(neo4jConnPool))

	shutdownGracefully := func() {

		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

		// gracefully dispose resources
		err = eventConsumer.Close(ctx)
		logIfError(err)

		err = kafkaConsumer.Close(ctx)
		logIfError(err)

		err = kafkaProducer.Close(ctx)
		logIfError(err)

		err = kafkaErrorProducer.Close(ctx)
		logIfError(err)

		err = healthChecker.Close(ctx)
		logIfError(err)

		// cancel the timer in the shutdown context.
		cancel()

		log.Info("graceful shutdown was successful", nil)
		os.Exit(0)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Check that a valid service token was provided on startup
	go func() {
		log.Info("Checking service token", nil)
		err = auth.CheckServiceIdentity(context.Background(), cfg.ZebedeeURL, cfg.ServiceAuthToken)
		if err != nil {
			errorChannel <- err
			isReady <- false
		} else {
			isReady <- true
		}
	}()

	for {
		select {
		case err := <-kafkaConsumer.Errors():
			log.ErrorC("kafka consumer", err, nil)
			shutdownGracefully()
		case err := <-kafkaProducer.Errors():
			log.ErrorC("kafka result producer", err, nil)
			shutdownGracefully()
		case err := <-kafkaErrorProducer.Errors():
			log.ErrorC("kafka error producer", err, nil)
			shutdownGracefully()
		case err := <-errorChannel:
			log.ErrorC("error channel", err, nil)
			shutdownGracefully()
		case <-signals:
			log.Debug("os signal received", nil)
			shutdownGracefully()
		}
	}
}

func exitIfError(err error) {
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
}

func logIfError(err error) {
	if err != nil {
		log.Error(err, nil)
	}
}
