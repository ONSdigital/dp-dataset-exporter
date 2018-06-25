package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"

	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/go-ns/clients/dataset"
	filterHealthCheck "github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/neo4j"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/ONSdigital/go-ns/vault"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
)

func main() {
	log.Namespace = "dp-dataset-exporter"
	log.Info("Starting dataset exporter", nil)

	cfg, err := config.Get()
	exitIfError(err)

	log.Info("loaded config", log.Data{"config": cfg})

	// a channel used to signal when an exit is required
	errorChannel := make(chan error)

	kafkaBrokers := cfg.KafkaAddr
	kafkaConsumer, err := kafka.NewSyncConsumer(
		kafkaBrokers,
		cfg.FilterConsumerTopic,
		cfg.FilterConsumerGroup,
		kafka.OffsetNewest,
	)
	exitIfError(err)

	kafkaProducer, err := kafka.NewProducer(kafkaBrokers, cfg.CSVExportedProducerTopic, 0)
	exitIfError(err)

	kafkaErrorProducer, err := kafka.NewProducer(cfg.KafkaAddr, cfg.ErrorProducerTopic, 0)
	exitIfError(err)

	neo4jConnPool, err := bolt.NewClosableDriverPool(cfg.DatabaseAddress, cfg.Neo4jPoolSize)
	exitIfError(err)

	vaultClient, err := vault.CreateVaultClient(cfg.VaultToken, cfg.VaultAddress, 3)
	exitIfError(err)

	// when errors occur - we send a message on an error topic.
	errorHandler := errors.NewKafkaHandler(kafkaErrorProducer)

	httpClient := rchttp.ClientWithServiceToken(
		rchttp.ClientWithTimeout(nil, time.Second*15),
		cfg.ServiceAuthToken,
	)
	filterStore := filter.NewStore(cfg.FilterAPIURL, httpClient)

	observationStore := observation.NewStore(neo4jConnPool)
	fileStore, err := file.NewStore(cfg.AWSRegion, cfg.S3BucketName, cfg.S3PrivateBucketName, cfg.VaultPath, vaultClient)
	exitIfError(err)
	eventProducer := event.NewAvroProducer(kafkaProducer, schema.CSVExportedEvent)

	datasetAPICli := dataset.NewAPIClient(cfg.DatasetAPIURL, cfg.DatasetAPIAuthToken, "")

	eventHandler := event.NewExportHandler(filterStore, observationStore, fileStore, eventProducer, datasetAPICli, cfg.DownloadServiceURL)

	// eventConsumer will Consume when the service is healthy - see goroutine below
	eventConsumer := event.NewConsumer()

	healthAlertChan := make(chan bool, 1)
	healthcheckRequestChan := make(chan bool, 1)
	healthChecker := healthcheck.NewServerWithAlerts(
		cfg.BindAddr,
		cfg.HealthCheckInterval, cfg.HealthCheckRecoveryInterval,
		errorChannel,
		healthAlertChan, healthcheckRequestChan,
		filterHealthCheck.New(cfg.FilterAPIURL, "", ""),
		neo4j.NewHealthCheckClient(neo4jConnPool),
		vaultClient,
		datasetAPICli,
	)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Check that the healthChecker succeeds before testing the service token (GetDatasets)
	// once both succeed, then we can set Consume off
	// else exit program (via errorChannel) if the above fails to happen within the StartupTimeout
	ctxStartup, startupCancel := context.WithTimeout(context.Background(), cfg.StartupTimeout)
	go func() {
		defer startupCancel()
		healthOK := false
		var err error
		log.Info("Checking service token", nil)
		for {
			select {
			case healthOK = <-healthAlertChan:
			case <-ctxStartup.Done():
				errorChannel <- ctxStartup.Err()
				return
			case <-time.After(time.Second * 2):
				if healthOK {
					if _, err = datasetAPICli.GetDatasets(ctxStartup); err == nil {
						eventConsumer.Consume(kafkaConsumer, eventHandler, errorHandler, healthAlertChan)
						return
					}
				}
			}
		}
	}()

	// block until a fatal error occurs
	select {
	case err := <-kafkaConsumer.Errors():
		log.ErrorC("kafka consumer", err, nil)
	case err := <-kafkaProducer.Errors():
		log.ErrorC("kafka result producer", err, nil)
	case err := <-kafkaErrorProducer.Errors():
		log.ErrorC("kafka error producer", err, nil)
	case err := <-errorChannel:
		log.ErrorC("error channel", err, nil)
	case <-signals:
		log.Debug("os signal received", nil)
	}

	// shutdown within timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	go func() {
		defer cancel()

		// gracefully dispose resources
		err = eventConsumer.Close(ctx)
		logIfError(err)

		err = kafkaConsumer.StopListeningToConsumer(ctx)
		logIfError(err)

		err = kafkaConsumer.Close(ctx)
		logIfError(err)

		err = kafkaProducer.Close(ctx)
		logIfError(err)

		err = kafkaErrorProducer.Close(ctx)
		logIfError(err)

		err = healthChecker.Close(ctx)
		logIfError(err)
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	log.Info("shutdown complete", log.Data{"ctx": ctx.Err()})
	os.Exit(1)
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
