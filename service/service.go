package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	"github.com/ONSdigital/dp-graph/v2/graph"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Service contains all the configs, server and clients to run the event handler service
type Service struct {
	server          HTTPServer
	router          *mux.Router
	serviceList     *ExternalServiceList
	healthCheck     HealthChecker
	consumer        kafka.IConsumerGroup
	producer        kafka.IProducer
	shutdownTimeout time.Duration
}

// Run the service
func Run(ctx context.Context, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {
	log.Info(ctx, "running service")

	// Read config
	cfg, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve service configuration")
	}
	log.Info(ctx, "got service configuration", log.Data{"config": cfg})

	// Get HTTP Server with collectionID checkHeader middlew are
	r := mux.NewRouter()

	var s HTTPServer
	if cfg.OtelEnabled {
		otelHandler := otelhttp.NewHandler(r, "/")
		r.Use(otelmux.Middleware(cfg.OTServiceName))
		s = serviceList.GetHTTPServer(cfg.BindAddr, otelHandler)
	} else {
		s = serviceList.GetHTTPServer(cfg.BindAddr, r)
	}

	// Get Kafka consumer
	consumer, err := serviceList.GetKafkaConsumer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise kafka consumer", err)
		return nil, err
	}

	// Get Kafka consumer
	producer, err := serviceList.GetKafkaProducer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise kafka producer", err)
		return nil, err
	}

	eventProducer := event.NewAvroProducer(producer.Channels().Output, schema.CSVExportedEvent)

	filterStore, err := serviceList.GetFilterStore(cfg, cfg.ServiceAuthToken)
	if err != nil {
		log.Fatal(ctx, "failed to initialise filter store", err)
		return nil, err
	}

	// get dataset api
	datasetAPICli, err := serviceList.GetDatasetAPIClient(cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise dataset api", err)
		return nil, err
	}

	observationStore, err := serviceList.GetObservationStore(ctx)
	if err != nil {
		log.Fatal(ctx, "failed to initialise observation store", err)
		return nil, err
	}

	fileStore, err := serviceList.GetFileStore(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise dataset api", err)
		return nil, err
	}

	eventHandler := event.NewExportHandler(
		filterStore,
		observationStore,
		fileStore,
		eventProducer,
		datasetAPICli,
		cfg,
	)

	// Event Handler for Kafka Consumer
	event.Consume(ctx, consumer, eventHandler, cfg)

	if consumerStartErr := consumer.Start(); consumerStartErr != nil {
		log.Fatal(ctx, "error starting the consumer", consumerStartErr)
		return nil, consumerStartErr
	}

	// Kafka error logging go-routine
	consumer.LogErrors(ctx)

	// Get HealthCheck
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}

	if err := registerCheckers(ctx, hc, consumer, producer, *observationStore, datasetAPICli, filterStore, fileStore.Uploader, fileStore.PrivateUploader, health.NewClient("Zebedee", cfg.ZebedeeURL)); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	r.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return &Service{
		server:          s,
		router:          r,
		serviceList:     serviceList,
		healthCheck:     hc,
		consumer:        consumer,
		producer:        producer,
		shutdownTimeout: cfg.GracefulShutdownTimeout,
	}, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.shutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)

	// track shutdown gracefully closes up
	var gracefulShutdown bool

	go func() {
		defer cancel()
		var hasShutdownError bool

		// stop healthcheck, as it depends on everything else
		if svc.serviceList.HealthCheck {
			svc.healthCheck.Stop()
		}

		// If kafka consumer exists, stop listening to it.
		// This will automatically stop the event consumer loops and no more messages will be processed.
		// The kafka consumer will be closed after the service shuts down.
		if svc.serviceList.KafkaConsumer {
			log.Info(ctx, "stopping kafka consumer listener")
			if err := svc.consumer.Stop(); err != nil {
				log.Error(ctx, "error stopping kafka consumer listener", err)
				hasShutdownError = true
			}
			log.Info(ctx, "stopped kafka consumer listener")
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// If kafka consumer exists, close it.
		if svc.serviceList.KafkaConsumer {
			log.Info(ctx, "closing kafka consumer")
			if err := svc.consumer.Close(ctx); err != nil {
				log.Error(ctx, "error closing kafka consumer", err)
				hasShutdownError = true
			}
			log.Info(ctx, "closed kafka consumer")
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

func registerCheckers(ctx context.Context, hc HealthChecker,
	consumer kafka.IConsumerGroup, producer kafka.IProducer, observationCli graph.DB, datasetCli DatasetAPI, filterStore FilterStore, publicUploader, privateUploader file.Uploader, zebedeeCli *health.Client) (err error) {
	hasErrors := false

	if err := hc.AddCheck("Kafka consumer", consumer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Kafka consumer", err)
	}

	if err := hc.AddCheck("Kafka producer", producer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Kafka producer", err)
	}

	if err := hc.AddCheck("Observation store", observationCli.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Observation store", err)
	}

	if err := hc.AddCheck("Dataset API", datasetCli.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Dataset API", err)
	}

	if err := hc.AddCheck("Filter store", filterStore.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Filter store", err)
	}

	if err = hc.AddCheck("Zebedee", zebedeeCli.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for zebedee", err)
	}

	if err = hc.AddCheck(fmt.Sprintf("S3 %s bucket", publicUploader.BucketName()), publicUploader.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for public s3 bucket", err)
	}

	if err = hc.AddCheck(fmt.Sprintf("S3 %s private bucket", privateUploader.BucketName()), privateUploader.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for private s3 bucket", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
