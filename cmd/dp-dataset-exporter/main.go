package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-graph/graph"
	"github.com/ONSdigital/go-ns/clients/dataset"
	filterHealthCheck "github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/ONSdigital/go-ns/vault"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
)

const (
	serviceConsumer            = "kafka-consumer"
	serviceCsvExportedProducer = "kafka-csv-exported-producer"
	serviceErrorProducer       = "kafka-error-producer"
	serviceFileStore           = "file-store"
	serviceGraph               = "graph"
	serviceHealthTicker        = "health-ticker"
	serviceVault               = "vault"
)

var services = map[string]bool{
	serviceConsumer:            false,
	serviceCsvExportedProducer: false,
	serviceErrorProducer:       false,
	serviceFileStore:           false,
	serviceGraph:               false,
	serviceHealthTicker:        false,
	serviceVault:               false,
}

func main() {
	log.Namespace = "dp-dataset-exporter"
	log.Info("Starting dataset exporter", nil)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

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
	updateInitialisation(err, services, serviceConsumer)

	kafkaProducer, err := kafka.NewProducer(kafkaBrokers, cfg.CSVExportedProducerTopic, 0)
	updateInitialisation(err, services, serviceCsvExportedProducer)

	kafkaErrorProducer, err := kafka.NewProducer(cfg.KafkaAddr, cfg.ErrorProducerTopic, 0)
	updateInitialisation(err, services, serviceErrorProducer)

	vaultClient, err := vault.CreateVaultClient(cfg.VaultToken, cfg.VaultAddress, 3)
	updateInitialisation(err, services, serviceVault)

	// when errors occur - we send a message on an error topic.
	errorHandler := errors.NewKafkaHandler(kafkaErrorProducer)

	httpClient := rchttp.ClientWithServiceToken(
		rchttp.ClientWithTimeout(nil, time.Second*15),
		cfg.ServiceAuthToken,
	)
	filterStore := filter.NewStore(cfg.FilterAPIURL, httpClient)

	observationStore, err := graph.New(context.Background(), graph.Subsets{Observation: true})
	updateInitialisation(err, services, serviceGraph)

	fileStore, err := file.NewStore(
		cfg.AWSRegion,
		cfg.S3BucketName,
		cfg.S3PrivateBucketName,
		cfg.VaultPath,
		vaultClient,
	)
	updateInitialisation(err, services, serviceFileStore)

	eventProducer := event.NewAvroProducer(kafkaProducer, schema.CSVExportedEvent)

	datasetAPICli := dataset.NewAPIClient(cfg.DatasetAPIURL, cfg.ServiceAuthToken, "")

	eventHandler := event.NewExportHandler(
		filterStore,
		observationStore,
		fileStore,
		eventProducer,
		datasetAPICli,
		cfg.DownloadServiceURL,
		cfg.APIDomainURL,
		cfg.FullDatasetFilePrefix,
		cfg.FilteredDatasetFilePrefix,
	)

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
		observationStore,
		vaultClient,
		datasetAPICli,
	)
	services[serviceHealthTicker] = true

	// Gracefully shutdown the application closing any open resources
	gracefulShutdown := func() {
		log.Info("gracefully shutting down", log.Data{"graceful_shutdown_timeout": cfg.GracefulShutdownTimeout})
		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

		if services[serviceConsumer] {
			err = eventConsumer.Close(ctx)
			logIfError(err)

			err = kafkaConsumer.StopListeningToConsumer(ctx)
			logIfError(err)

			err = kafkaConsumer.Close(ctx)
			logIfError(err)
		}

		if services[serviceCsvExportedProducer] {
			err = kafkaProducer.Close(ctx)
			logIfError(err)
		}

		if services[serviceErrorProducer] {
			err = kafkaErrorProducer.Close(ctx)
			logIfError(err)
		}

		if services[serviceGraph] {
			err = observationStore.Close(ctx)
			logIfError(err)
		}

		if services[serviceHealthTicker] {
			err = healthChecker.Close(ctx)
			logIfError(err)
		}

		log.Info("shutdown complete", log.Data{"ctx": ctx.Err()})

		cancel()
		os.Exit(1)
	}

	// Check that the healthChecker succeeds before testing the service token (GetDatasets)
	// once both succeed, then we can set Consume off
	go func() {
		healthOK := false
		var err error
		log.Info("Checking service token", nil)
		for {
			select {
			case healthOK = <-healthAlertChan:
			case <-time.After(time.Second * 2):
				// FIXME Once HealthCheck has been added to kafka consumer groups, consumers
				// and producers this extra check can then be removed `services[serviceConsumer]`
				if healthOK && services[serviceConsumer] {
					if _, err = datasetAPICli.GetDatasets(context.Background()); err == nil {
						eventConsumer.Consume(kafkaConsumer, eventHandler, errorHandler, healthAlertChan)
						return
					}
				}
			}
		}
	}()

	for {
		select {
		case <-signals:
			log.Debug("os signal received", nil)
			gracefulShutdown()
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

func updateInitialisation(err error, service map[string]bool, name string) {
	if err != nil {
		log.Error(err, nil)
		return
	}

	service[name] = true
}
