package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/initialise"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/go-ns/server"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	ctx := context.Background()
	log.Namespace = "dp-dataset-exporter"
	log.Event(ctx, "Starting dataset exporter")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	exitIfError(ctx, err)

	log.Event(ctx, "loaded config", log.Data{"config": cfg})

	// a channel used to signal when an exit is required
	errorChannel := make(chan error)

	// serviceList keeps track of what dependency services have been initialised
	var serviceList initialise.ExternalServiceList

	// Create kafka Consumer
	kafkaConsumer, err := serviceList.GetConsumer(cfg)
	exitIfError(ctx, err)

	// Create kafka Producer
	kafkaProducer, err := serviceList.GetProducer(
		cfg.KafkaAddr,
		cfg.CSVExportedProducerTopic,
		initialise.CSVExported,
	)
	exitIfError(ctx, err)

	// Create kafka ErrorProducer
	kafkaErrorProducer, err := serviceList.GetProducer(
		cfg.KafkaAddr,
		cfg.ErrorProducerTopic,
		initialise.Error,
	)
	exitIfError(ctx, err)

	// Create vault client
	vaultClient, err := serviceList.GetVault(cfg, 3)
	logIfError(ctx, err)

	// when errors occur - we send a message on an error topic.
	errorHandler := errors.NewKafkaHandler(kafkaErrorProducer.Channels().Output)

	// TODO move this to dp-rchttp?
	// context.WithValue(ctx)
	// httpClient := rchttp.ClientWithServiceToken(
	// 	rchttp.ClientWithTimeout(nil, time.Second*15),
	// 	cfg.ServiceAuthToken,
	// )
	httpClient := rchttp.ClientWithTimeout(nil, time.Second*15)
	// TODO Add service auth token?
	filterStore := filter.NewStore(cfg.FilterAPIURL, httpClient)

	observationStore, err := serviceList.GetObservationStore()
	logIfError(ctx, err)

	fileStore, err := serviceList.GetFileStore(cfg, vaultClient)
	logIfError(ctx, err)

	eventProducer := event.NewAvroProducer(kafkaProducer.Channels().Output, schema.CSVExportedEvent)

	datasetAPICli := dataset.NewAPIClient(cfg.DatasetAPIURL)

	eventHandler := event.NewExportHandler(
		filterStore,
		observationStore,
		fileStore,
		eventProducer,
		datasetAPICli,
		cfg,
	)

	// eventConsumer will Consume when the service is healthy - see goroutine below
	eventConsumer := event.NewConsumer()

	// Create healthcheck object with versionInfo
	hc, err := serviceList.GetHealthCheck(cfg, BuildTime, GitCommit, Version)
	if err != nil {
		log.Event(ctx, "failed to create service version information", log.Error(err))
		os.Exit(1)
	}

	// TODO add Checkers to hc
	r := mux.NewRouter()
	r.HandleFunc("/health", hc.Handler)

	// Start healthcheck
	hc.Start(ctx)

	// Create and start http server for healthcheck
	httpServer := server.New(cfg.BindAddr, r)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Event(ctx, "failed to start healthcheck HTTP server", log.Data{"config": cfg}, log.Error(err))
			os.Exit(2)
		}
	}()

	// healthAlertChan := make(chan bool, 1)
	// healthcheckRequestChan := make(chan bool, 1)
	// healthChecker := healthcheck.NewServerWithAlerts(
	// 	cfg.BindAddr,
	// 	cfg.HealthCheckInterval, cfg.HealthCheckRecoveryInterval,
	// 	errorChannel,
	// 	healthAlertChan, healthcheckRequestChan,
	// 	filterHealthCheck.New(cfg.FilterAPIURL, "", ""),
	// 	observationStore,
	// 	vaultClient,
	// 	datasetAPICli,
	// )
	// serviceList.HealthTicker = true

	// // Check that the healthChecker succeeds before testing the service token (GetDatasets)
	// // once both succeed, then we can set Consume off
	// go func() {
	// 	healthOK := false
	// 	var err error
	// 	log.Event(ctx, "Checking service token", nil)
	// 	for {
	// 		select {
	// 		case healthOK = <-healthAlertChan:
	// 		case <-time.After(time.Second * 2):
	// 			// FIXME Once HealthCheck has been added to kafka consumer groups, consumers
	// 			// and producers this extra check can then be removed `services[serviceConsumer]`
	// 			if healthOK && serviceList.Consumer {
	// 				// func (c *Client) GetDatasets(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string) (m List, err error) {
	// 				userAuthToken := "" // TODO do we have userAuthToken?
	// 				collectionID := ""  // TODO where do we get collectionID?
	// 				if _, err = datasetAPICli.GetDatasets(context.Background(), userAuthToken, cfg.ServiceAuthToken, collectionID); err == nil {
	// 					serviceList.EventConsumer = true
	// 					eventConsumer.Consume(kafkaConsumer, eventHandler, errorHandler, healthAlertChan)
	// 					return
	// 				}
	// 			}
	// 		}
	// 	}
	// }()

	// Check that the healthChecker succeeds before testing the service token (GetDatasets)
	// once both succeed, then we can set Consume off
	go func() {
		for {
			select {
			case <-time.After(time.Second * 2):
				if serviceList.Consumer == false {
					// Consumer not created yet
					continue
				}
				if err = kafkaConsumer.Initialise(ctx); err != nil {
					// Kafka client cannot be initialised
					continue
				}
				userAuthToken := "" // TODO do we have userAuthToken?
				collectionID := ""  // TODO where do we get collectionID?
				if _, err = datasetAPICli.GetDatasets(ctx, userAuthToken, cfg.ServiceAuthToken, collectionID); err != nil {
					// GetDatasets failed
					continue
				}
				// Kafka initialised and dataset client got datasets -> Start consuming
				serviceList.EventConsumer = true
				eventConsumer.Consume(kafkaConsumer, eventHandler, errorHandler)
				return
			}
		}
	}()

	// kafka error logging go-routines
	kafkaConsumer.Channels().LogErrors(ctx, "kafka consumer")
	kafkaProducer.Channels().LogErrors(ctx, "kafka result producer")
	kafkaErrorProducer.Channels().LogErrors(ctx, "kafka error producer")
	go func() {
		select {
		case err := <-errorChannel:
			log.Event(ctx, "error channel", log.Error(err))
		}
	}()

	// block until a fatal error / OS signal occurs
	select {
	case <-signals:
		log.Event(ctx, "os signal received")
	}

	log.Event(ctx, "gracefully shutting down", log.Data{"graceful_shutdown_timeout": cfg.GracefulShutdownTimeout})
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	// Gracefully shutdown the application closing any open resources
	go func() {
		defer cancel()

		// Health Checker should always be closed first as it relies on other
		// services (clients) to exist - prevents DATA RACE
		if serviceList.HealthCheck {
			log.Event(ctx, "stopping healthchecker")
			hc.Stop()
		}

		// TODO stop HTTP server
		// log.Event(ctx, "stopping healthcheck HTTP server")
		// err := httpServer.Server.Shutdown(ctx)
		// logIfError(ctx, err)

		if serviceList.EventConsumer {
			log.Event(ctx, "closing event consumer")
			logIfError(ctx, eventConsumer.Close(ctx))
		}

		if serviceList.Consumer {
			log.Event(ctx, "stop listening to consumer")
			logIfError(ctx, kafkaConsumer.StopListeningToConsumer(ctx))

			log.Event(ctx, "closing consumer")
			logIfError(ctx, kafkaConsumer.Close(ctx))
		}

		if serviceList.CSVExportedProducer {
			log.Event(ctx, "closing csv exporter producer")
			logIfError(ctx, kafkaProducer.Close(ctx))
		}

		if serviceList.ErrorProducer {
			log.Event(ctx, "closing error producer")
			logIfError(ctx, kafkaErrorProducer.Close(ctx))
		}

		if serviceList.ObservationStore {
			log.Event(ctx, "closing observation store")
			logIfError(ctx, observationStore.Close(ctx))
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	log.Event(ctx, "shutdown complete", log.Data{"ctx": ctx.Err()})
	os.Exit(1)
}

func logIfError(ctx context.Context, err error) {
	if err != nil {
		log.Event(ctx, "error", log.Error(err))
		return
	}
}

func exitIfError(ctx context.Context, err error) {
	if err != nil {
		log.Event(ctx, "fatal error", log.Error(err))
		os.Exit(1)
	}
}
