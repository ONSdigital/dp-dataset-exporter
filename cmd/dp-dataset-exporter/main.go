package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	kafka "github.com/ONSdigital/dp-kafka"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/log.go/log"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/initialise"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
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
	kafkaConsumer, err := serviceList.GetConsumer(cfg.KafkaAddr, cfg)
	handleError(ctx, err)

	// Create kafka Producer
	kafkaProducer, err := serviceList.GetProducer(
		cfg.KafkaAddr,
		cfg.CSVExportedProducerTopic,
		initialise.CSVExported,
	)
	handleError(ctx, err)

	// Create kafka ErrorProducer
	kafkaErrorProducer, err := serviceList.GetProducer(
		cfg.KafkaAddr,
		cfg.ErrorProducerTopic,
		initialise.Error,
	)
	handleError(ctx, err)

	// Create vault client
	vaultClient, err := serviceList.GetVault(cfg, 3)
	handleError(ctx, err)

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
	handleError(ctx, err)

	fileStore, err := serviceList.GetFileStore(cfg, vaultClient)
	handleError(ctx, err)

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
				if err = kafkaConsumer.InitialiseSarama(ctx); err != nil {
					// Kafka client cannot be initialised
					continue
				}
				userAuthToken := "" // TODO do we have userAuthToken?
				collectionID := ""  // TODO where do we get collectionID?
				if _, err = datasetAPICli.GetDatasets(ctx, userAuthToken, cfg.ServiceAuthToken, collectionID); err != nil {
					// GetDatasets failed
					continue
				}
				// Success
				serviceList.EventConsumer = true
				eventConsumer.Consume(kafkaConsumer, eventHandler, errorHandler)
				return
			}
		}
	}()

	// kafka error handling go-routines
	logErrorsFromChan(ctx, kafkaConsumer.Channels().Errors, kafkaConsumer.Channels().Closer, "kafka consumer")
	logErrorsFromChan(ctx, kafkaProducer.Channels().Errors, kafkaProducer.Channels().Closer, "kafka result producer")
	logErrorsFromChan(ctx, kafkaErrorProducer.Channels().Errors, kafkaErrorProducer.Channels().Closer, "kafka error producer")
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

		// Health Ticker/Checker should always be closed first as it relies on other
		// services (clients) to exist - prevents DATA RACE
		// if serviceList.HealthTicker {
		// 	log.Event(ctx, "closing healthchecker")
		// 	handleError(ctx, healthChecker.Close(ctx))
		// }

		if serviceList.EventConsumer {
			log.Event(ctx, "closing event consumer")
			handleError(ctx, eventConsumer.Close(ctx))
		}

		if serviceList.Consumer {
			log.Event(ctx, "stop listening to consumer")
			handleError(ctx, kafkaConsumer.StopListeningToConsumer(ctx))

			log.Event(ctx, "closing consumer")
			handleError(ctx, kafkaConsumer.Close(ctx))
		}

		if serviceList.CSVExportedProducer {
			log.Event(ctx, "closing csv exporter producer")
			handleError(ctx, kafkaProducer.Close(ctx))
		}

		if serviceList.ErrorProducer {
			log.Event(ctx, "closing error producer")
			handleError(ctx, kafkaErrorProducer.Close(ctx))
		}

		if serviceList.ObservationStore {
			log.Event(ctx, "closing observation store")
			handleError(ctx, observationStore.Close(ctx))
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	log.Event(ctx, "shutdown complete", log.Data{"ctx": ctx.Err()})
	os.Exit(1)
}

// logErrorsFromChan creates a go-routine that waits on chErrors channel and logs any error received. It exists on chCloser channel event
func logErrorsFromChan(ctx context.Context, chErrors chan error, chCloser chan struct{}, errMsg string) {
	go func() {
		for true {
			select {
			case err := <-chErrors:
				log.Event(ctx, errMsg, log.Error(err))
			case <-chCloser:
				return
			}
		}
	}()
}

// handleError exists on fatal errors and logs other errors without exiting
func handleError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	switch err.(type) {
	case *kafka.ErrNoChannel:
		log.Event(ctx, "fatal error", log.Error(err))
		os.Exit(1)
	default:
		log.Event(ctx, "error", log.Error(err))
	}
}

// exitIfError logs and exits if there is an error
func exitIfError(ctx context.Context, err error) {
	if err != nil {
		log.Event(ctx, "fatal error", log.Error(err))
		os.Exit(1)
	}
}
