package main

import (
	"context"
	errs "errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	filterCli "github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	vault "github.com/ONSdigital/dp-vault"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/file"
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
	log.Event(ctx, "Starting dataset exporter", log.INFO)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	exitIfError(ctx, err)

	log.Event(ctx, "loaded config", log.INFO, log.Data{"config": cfg})

	// a channel used to signal when an exit is required
	errorChannel := make(chan error)

	// serviceList keeps track of what dependency services have been initialised
	var serviceList initialise.ExternalServiceList

	// Create kafka Consumer - exit on channel validation error. Non-initialised consumers will not error at creation time.
	kafkaConsumer, err := serviceList.GetConsumer(ctx, cfg)
	exitIfError(ctx, err)

	// Create kafka Producer - exit on channel validation error. Non-initialised producers will not error at creation time.
	kafkaProducer, err := serviceList.GetProducer(
		ctx,
		cfg.KafkaAddr,
		cfg.CSVExportedProducerTopic,
		initialise.CSVExported,
	)
	exitIfError(ctx, err)

	// Create kafka ErrorProducer - exit on channel validation error. Non-initialised producers will not error at creation time.
	kafkaErrorProducer, err := serviceList.GetProducer(
		ctx,
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

	filterAPIClient := filterCli.New(cfg.FilterAPIURL)
	filterStore := filter.NewStore(filterAPIClient, cfg.ServiceAuthToken)

	observationStore, err := serviceList.GetObservationStore(ctx)
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
		log.Event(ctx, "failed to create service version information", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	// Add checkers to healthcheck
	err = registerCheckers(ctx, &hc,
		kafkaProducer, kafkaErrorProducer, kafkaConsumer,
		vaultClient,
		datasetAPICli,
		filterAPIClient,
		health.NewClient("DownloadService", cfg.DownloadServiceURL),
		health.NewClient("Zebedee", cfg.ZebedeeURL),
		fileStore.Uploader, fileStore.CryptoUploader)
	if err != nil {
		os.Exit(1)
	}

	httpServer := startHealthCheck(ctx, &hc, cfg.BindAddr)

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
				if _, err = datasetAPICli.GetDatasets(ctx, "", cfg.ServiceAuthToken, ""); err != nil {
					// GetDatasets failed
					continue
				}
				// Kafka initialised and dataset client got datasets -> Start consuming
				serviceList.EventConsumer = true
				log.Event(ctx, "starting to consume messages", log.INFO)
				eventConsumer.Consume(kafkaConsumer, eventHandler, errorHandler)
				return
			}
		}
	}()

	// kafka error logging go-routines
	kafkaConsumer.Channels().LogErrors(ctx, "kafka consumer")
	kafkaProducer.Channels().LogErrors(ctx, "kafka result producer")
	kafkaErrorProducer.Channels().LogErrors(ctx, "kafka error producer")
	// error logging go-routine for errorChannel and httpServerDoneChannels
	go func() {
		for {
			select {
			case err := <-errorChannel:
				log.Event(ctx, "error channel", log.ERROR, log.Error(err))
			}
		}
	}()

	// block until a fatal error / OS signal occurs
	select {
	case <-signals:
		log.Event(ctx, "os signal received", log.INFO)
	}

	log.Event(ctx, "gracefully shutting down", log.INFO, log.Data{"graceful_shutdown_timeout": cfg.GracefulShutdownTimeout})
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.GracefulShutdownTimeout)

	// Gracefully shutdown the application closing any open resources
	go func() {
		defer cancel()

		// Health Checker should always be closed first as it relies on other
		// services (clients) to exist - prevents DATA RACE
		if serviceList.HealthCheck {
			log.Event(shutdownCtx, "stopping healthchecker", log.INFO)
			hc.Stop()
		}

		log.Event(shutdownCtx, "shutting down http server", log.INFO)
		logIfError(shutdownCtx, httpServer.Shutdown(shutdownCtx))

		if serviceList.Consumer {
			log.Event(shutdownCtx, "stop listening to consumer", log.INFO)
			logIfError(shutdownCtx, kafkaConsumer.StopListeningToConsumer(shutdownCtx))
		}

		if serviceList.CSVExportedProducer {
			log.Event(shutdownCtx, "closing csv exporter producer", log.INFO)
			logIfError(shutdownCtx, kafkaProducer.Close(shutdownCtx))
		}

		if serviceList.ErrorProducer {
			log.Event(shutdownCtx, "closing error producer", log.INFO)
			logIfError(shutdownCtx, kafkaErrorProducer.Close(shutdownCtx))
		}

		if serviceList.EventConsumer {
			log.Event(shutdownCtx, "closing event consumer", log.INFO)
			logIfError(shutdownCtx, eventConsumer.Close(shutdownCtx))
		}

		if serviceList.Consumer {
			log.Event(shutdownCtx, "closing consumer", log.INFO)
			logIfError(shutdownCtx, kafkaConsumer.Close(shutdownCtx))
		}

		if serviceList.ObservationStore {
			log.Event(shutdownCtx, "closing observation store", log.INFO)
			logIfError(shutdownCtx, observationStore.Close(shutdownCtx))
		}

	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-shutdownCtx.Done()

	log.Event(shutdownCtx, "shutdown complete", log.INFO, log.Data{"ctx": shutdownCtx.Err()})
	os.Exit(0)
}

// startHealthCheck sets up the Handler, starts the healthcheck and the http server that serves healthcheck endpoint
func startHealthCheck(ctx context.Context, hc *healthcheck.HealthCheck, bindAddr string) *server.Server {

	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)

	httpServer := server.New(bindAddr, router)
	httpServer.HandleOSSignals = false

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Event(ctx, "httpServer error", log.ERROR, log.Error(err))
		}
	}()
	return httpServer
}

// registerCheckers adds the checkers for the provided clients to the healthcheck object
func registerCheckers(ctx context.Context, hc *healthcheck.HealthCheck,
	kafkaProducer, kafkaErrorProducer *kafka.Producer, kafkaConsumer *kafka.ConsumerGroup,
	vaultClient *vault.Client,
	datasetAPICli *dataset.Client,
	filterAPICli *filterCli.Client,
	downloadServiceCli, zebedeeCli *health.Client,
	publicUploader, privateUploader file.Uploader) (err error) {

	hasErrors := false

	if err = hc.AddCheck("Kafka Producer", kafkaProducer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for Kafka Producer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Kafka Error Producer", kafkaErrorProducer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for Kafka Error Producer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Kafka Consumer", kafkaConsumer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for Kafka Consumer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Vault", vaultClient.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for Vault", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Filter API", filterAPICli.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for Filter API", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Dataset API", datasetAPICli.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for Dataset API", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck(fmt.Sprintf("S3 %s bucket", publicUploader.BucketName()), publicUploader.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for public S3 bucket", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck(fmt.Sprintf("S3 %s private bucket", privateUploader.BucketName()), privateUploader.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for private S3 bucket", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Download Service", downloadServiceCli.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for Download Service", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Zebedee", zebedeeCli.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "Error Adding Check for Zebedee", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errs.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}

func logIfError(ctx context.Context, err error) {
	if err != nil {
		log.Event(ctx, "error", log.ERROR, log.Error(err))
		return
	}
}

func exitIfError(ctx context.Context, err error) {
	if err != nil {
		log.Event(ctx, "fatal error", log.FATAL, log.Error(err))
		os.Exit(1)
	}
}
