package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	bolt "github.com/ONSdigital/golang-neo4j-bolt-driver"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)

func main() {
	log.Namespace = "dp-dataset-exporter"
	log.Debug("Starting dataset exporter", nil)

	config, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	// Avoid logging the neo4j FileURL as it may contain a password
	log.Debug("loaded config", log.Data{
		"topics":         []string{config.FilterConsumerTopic, config.CSVExportedProducerTopic},
		"brokers":        config.KafkaAddr,
		"consumer_group": config.FilterConsumerGroup,
		"filter_api_url": config.FilterAPIURL,
		"aws_region":     config.AWSRegion,
		"s3_bucket_name": config.S3BucketName,
		"bind_addr":      config.BindAddr})

	// a channel used to signal a graceful exit is required.
	errorChannel := make(chan error)

	router := mux.NewRouter()
	router.Path("/healthcheck").HandlerFunc(healthcheck.Handler)
	httpServer := server.New(config.BindAddr, router)

	// Disable auto handling of os signals by the HTTP server. This is handled
	// in the service so we can gracefully shutdown resources other than just
	// the HTTP server.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("starting http server", log.Data{"bind_addr": config.BindAddr})
		if err := httpServer.ListenAndServe(); err != nil {
			errorChannel <- err
		}
	}()

	kafkaBrokers := config.KafkaAddr
	kafkaConsumer, err := kafka.NewConsumerGroup(
		kafkaBrokers,
		config.FilterConsumerTopic,
		config.FilterConsumerGroup,
		kafka.OffsetNewest)
	exitIfError(err)

	kafkaProducer, err := kafka.NewProducer(kafkaBrokers, config.CSVExportedProducerTopic, 0)
	exitIfError(err)

	kafkaErrorProducer, err := kafka.NewProducer(config.KafkaAddr, config.ErrorProducerTopic, 0)
	exitIfError(err)

	pool, err := bolt.NewClosableDriverPool(config.DatabaseAddress, config.Neo4jPoolSize)
	exitIfError(err)

	// when errors occur - we send a message on an error topic.
	errorHandler := errors.NewKafkaHandler(kafkaErrorProducer)

	httpClient := http.Client{Timeout: time.Second * 15}

	filterStore := filter.NewStore(config.FilterAPIURL, config.FilterAPIAuthToken, &httpClient)
	observationStore := observation.NewStore(pool)
	fileStore := file.NewStore(config.AWSRegion, config.S3BucketName)
	eventProducer := event.NewAvroProducer(kafkaProducer, schema.CSVExportedEvent)

	datasetAPICli := dataset.New(config.DatasetAPIURL)
	datasetAPICli.SetInternalToken(config.DatasetAPIAuthToken)

	generateFileID := uuid.NewV4().String()

	eventHandler := event.NewExportHandler(filterStore, observationStore, fileStore, eventProducer, datasetAPICli, generateFileID)

	eventConsumer := event.NewConsumer()
	eventConsumer.Consume(kafkaConsumer, eventHandler, errorHandler)

	shutdownGracefully := func() {

		ctx, cancel := context.WithTimeout(context.Background(), config.GracefulShutdownTimeout)

		// gracefully dispose resources
		err = eventConsumer.Close(ctx)
		logIfError(err)

		err = kafkaConsumer.Close(ctx)
		logIfError(err)

		err = kafkaProducer.Close(ctx)
		logIfError(err)

		err = kafkaErrorProducer.Close(ctx)
		logIfError(err)

		err = httpServer.Shutdown(ctx)
		logIfError(err)

		// cancel the timer in the shutdown context.
		cancel()

		log.Debug("graceful shutdown was successful", nil)
		os.Exit(0)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

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
