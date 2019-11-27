package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/go-ns/clients/dataset"
	filterHealthCheck "github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/initialise"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
)

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

	var serviceList initialise.ExternalServiceList

	kafkaBrokers := cfg.KafkaAddr
	kafkaConsumer, err := serviceList.GetConsumer(kafkaBrokers, cfg)
	logIfError(err)

	kafkaProducer, err := serviceList.GetProducer(
		kafkaBrokers,
		cfg.CSVExportedProducerTopic,
		initialise.CSVExportedProducer,
	)
	logIfError(err)

	kafkaErrorProducer, err := serviceList.GetProducer(
		kafkaBrokers,
		cfg.ErrorProducerTopic,
		initialise.ErrorProducer,
	)
	logIfError(err)

	vaultClient, err := serviceList.GetVault(cfg, 3)
	logIfError(err)

	// when errors occur - we send a message on an error topic.
	errorHandler := errors.NewKafkaHandler(kafkaErrorProducer)

	httpClient := rchttp.ClientWithServiceToken(
		rchttp.ClientWithTimeout(nil, time.Second*15),
		cfg.ServiceAuthToken,
	)
	filterStore := filter.NewStore(cfg.FilterAPIURL, httpClient)

	observationStore, err := serviceList.GetObservationStore()
	logIfError(err)

	fileStore, err := serviceList.GetFileStore(cfg, vaultClient)
	logIfError(err)

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
	serviceList.HealthTicker = true

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
				if healthOK && serviceList.Consumer {
					if _, err = datasetAPICli.GetDatasets(context.Background()); err == nil {
						eventConsumer.Consume(kafkaConsumer, eventHandler, errorHandler, healthAlertChan)
						return
					}
				}
			}
		}
	}()

	// block until a fatal error occurs
	select {
	case <-signals:
		log.Debug("os signal received", nil)
	}

	log.Info("gracefully shutting down", log.Data{"graceful_shutdown_timeout": cfg.GracefulShutdownTimeout})
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	// Gracefully shutdown the application closing any open resources
	go func() {
		defer cancel()

		if serviceList.Consumer {
			log.Info("closing event consumer", nil)
			logIfError(eventConsumer.Close(ctx))

			log.Info("stop listening to consumer", nil)
			logIfError(kafkaConsumer.StopListeningToConsumer(ctx))

			log.Info("closing consumer", nil)
			logIfError(kafkaConsumer.Close(ctx))
		}

		if serviceList.CSVExportedProducer {
			log.Info("closing csv exporter producer", nil)
			logIfError(kafkaProducer.Close(ctx))
		}

		if serviceList.ErrorProducer {
			log.Info("closing error producer", nil)
			logIfError(kafkaErrorProducer.Close(ctx))
		}

		if serviceList.ObservationStore {
			log.Info("closing observation store", nil)
			logIfError(observationStore.Close(ctx))
		}

		if serviceList.HealthTicker {
			log.Info("closing healthchecker", nil)
			logIfError(healthChecker.Close(ctx))
		}
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
