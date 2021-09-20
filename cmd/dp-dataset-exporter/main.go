package main

import (
	"context"
	errs "errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-graph/v2/graph"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	filterCli "github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	vault "github.com/ONSdigital/dp-vault"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/initialise"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	dphttp "github.com/ONSdigital/dp-net/http"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string

	/* NOTE: replace the above with the below to run code with for example vscode debugger.
	BuildTime string = "1601119818"
	GitCommit string = "6584b786caac36b6214ffe04bf62f058d4021538"
	Version   string = "v0.1.0"

	*/
)

func main() {
	ctx := context.Background()
	log.Namespace = "dp-dataset-exporter"
	log.Info(ctx, "starting dataset exporter")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	exitIfError(ctx, err)

	log.Info(ctx, "loaded config", log.Data{"config": cfg})

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
		cfg.KafkaVersion,
		cfg.KafkaSecProtocol, cfg.KafkaSecCACerts, cfg.KafkaSecClientCert, cfg.KafkaSecClientKey, cfg.KafkaSecSkipVerify,
		initialise.CSVExported,
	)
	exitIfError(ctx, err)

	// Create kafka ErrorProducer - exit on channel validation error. Non-initialised producers will not error at creation time.
	kafkaErrorProducer, err := serviceList.GetProducer(
		ctx,
		cfg.KafkaAddr,
		cfg.ErrorProducerTopic,
		cfg.KafkaVersion,
		cfg.KafkaSecProtocol, cfg.KafkaSecCACerts, cfg.KafkaSecClientCert, cfg.KafkaSecClientKey, cfg.KafkaSecSkipVerify,
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

	var graphErrorConsumer *graph.ErrorConsumer
	if serviceList.ObservationStore {
		graphErrorConsumer = graph.NewLoggingErrorConsumer(ctx, observationStore.Errors)
	}

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
	eventConsumer := event.NewConsumer(cfg.KafkaConsumerWorkers)

	// Create healthcheck object with versionInfo
	hc, err := serviceList.GetHealthCheck(cfg, BuildTime, GitCommit, Version)
	if err != nil {
		log.Fatal(ctx, "failed to create service version information", err)
		os.Exit(1)
	}

	// Add checkers to healthcheck
	err = registerCheckers(ctx, &hc,
		kafkaProducer, kafkaErrorProducer, kafkaConsumer,
		vaultClient,
		datasetAPICli,
		filterAPIClient,
		health.NewClient("Zebedee", cfg.ZebedeeURL),
		fileStore.Uploader,
		fileStore.CryptoUploader,
		observationStore)
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
				if _, err = datasetAPICli.GetDatasets(ctx, "", cfg.ServiceAuthToken, "", nil); err != nil {
					// GetDatasets failed
					continue
				}
				// Kafka initialised and dataset client got datasets -> Start consuming
				serviceList.EventConsumer = true
				log.Info(ctx, "starting to consume messages")
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
				log.Error(ctx, "error channel", err)
			}
		}
	}()

	// block until a fatal error / OS signal occurs
	select {
	case <-signals:
		log.Info(ctx, "os signal received")
	}

	log.Info(ctx, "gracefully shutting down", log.Data{"graceful_shutdown_timeout": cfg.GracefulShutdownTimeout})
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.GracefulShutdownTimeout)

	// Gracefully shutdown the application closing any open resources
	go func() {
		defer cancel()

		// Health Checker should always be closed first as it relies on other
		// services (clients) to exist - prevents DATA RACE
		if serviceList.HealthCheck {
			log.Info(shutdownCtx, "stopping health checker")
			hc.Stop()
		}

		log.Info(shutdownCtx, "shutting down http server")
		logIfError(shutdownCtx, httpServer.Shutdown(shutdownCtx))

		if serviceList.Consumer {
			log.Info(shutdownCtx, "stop listening to consumer")
			logIfError(shutdownCtx, kafkaConsumer.StopListeningToConsumer(shutdownCtx))
		}

		if serviceList.CSVExportedProducer {
			log.Info(shutdownCtx, "closing csv exporter producer")
			logIfError(shutdownCtx, kafkaProducer.Close(shutdownCtx))
		}

		if serviceList.ErrorProducer {
			log.Info(shutdownCtx, "closing error producer")
			logIfError(shutdownCtx, kafkaErrorProducer.Close(shutdownCtx))
		}

		if serviceList.EventConsumer {
			log.Info(shutdownCtx, "closing event consumer")
			logIfError(shutdownCtx, eventConsumer.Close(shutdownCtx))
		}

		if serviceList.Consumer {
			log.Info(shutdownCtx, "closing consumer")
			logIfError(shutdownCtx, kafkaConsumer.Close(shutdownCtx))
		}

		if serviceList.ObservationStore {
			log.Info(shutdownCtx, "closing observation store")
			logIfError(shutdownCtx, observationStore.Close(shutdownCtx))

			log.Info(shutdownCtx, "closing graph db error consumer")
			logIfError(shutdownCtx, graphErrorConsumer.Close(shutdownCtx))
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-shutdownCtx.Done()

	log.Info(shutdownCtx, "shutdown complete", log.Data{"ctx": shutdownCtx.Err()})
	os.Exit(0)
}

// startHealthCheck sets up the Handler, starts the healthcheck and the http server that serves healthcheck endpoint
func startHealthCheck(ctx context.Context, hc *healthcheck.HealthCheck, bindAddr string) *dphttp.Server {

	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)

	httpServer := dphttp.NewServer(bindAddr, router)
	httpServer.HandleOSSignals = false

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error(ctx, "httpServer error", err)
		}
	}()
	return httpServer
}

// registerCheckers adds the checkers for the provided clients to the healthcheck object
func registerCheckers(ctx context.Context, hc *healthcheck.HealthCheck, kafkaProducer, kafkaErrorProducer *kafka.Producer, kafkaConsumer *kafka.ConsumerGroup, vaultClient *vault.Client, datasetAPICli *dataset.Client, filterAPICli *filterCli.Client, zebedeeCli *health.Client, publicUploader, privateUploader file.Uploader, graphDB *graph.DB) (err error) {

	hasErrors := false

	if err = hc.AddCheck("Kafka Producer", kafkaProducer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for kafka producer", err)
	}

	if err = hc.AddCheck("Kafka Error Producer", kafkaErrorProducer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for kafka error producer", err)
	}

	if err = hc.AddCheck("Kafka Consumer", kafkaConsumer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for kafka consumer", err)
	}

	if err = hc.AddCheck("Vault", vaultClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for vault", err)
	}

	if err = hc.AddCheck("Filter API", filterAPICli.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for filter api", err)
	}

	if err = hc.AddCheck("Dataset API", datasetAPICli.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for dataset api", err)
	}

	if err = hc.AddCheck(fmt.Sprintf("S3 %s bucket", publicUploader.BucketName()), publicUploader.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for public s3 bucket", err)
	}

	if err = hc.AddCheck(fmt.Sprintf("S3 %s private bucket", privateUploader.BucketName()), privateUploader.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for private s3 bucket", err)
	}

	if err = hc.AddCheck("Zebedee", zebedeeCli.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for zebedee", err)
	}

	if err = hc.AddCheck("Graph DB", graphDB.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for graph db", err)
	}

	if hasErrors {
		return errs.New("error(s) registering checkers for healthcheck")
	}
	return nil
}

func logIfError(ctx context.Context, err error) {
	if err != nil {
		log.Error(ctx, "error", err)
		return
	}
}

func exitIfError(ctx context.Context, err error) {
	if err != nil {
		log.Fatal(ctx, "fatal error", err)
		os.Exit(1)
	}
}
