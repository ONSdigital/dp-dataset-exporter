package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/log.go/v2/log"
)

const serviceName = "dp-dataset-exporter"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	// Get Config
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error getting config", err)
		os.Exit(1)
	}

	// Avoid logging the neo4j FileURL as it may contain a password
	log.Info(ctx, "loaded config", log.Data{"config": cfg})

	pConfig := &kafka.ProducerConfig{
		KafkaVersion: &cfg.KafkaConfig.Version,
	}

	if cfg.KafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaConfig.SecCACerts,
			cfg.KafkaConfig.SecClientCert,
			cfg.KafkaConfig.SecClientKey,
			cfg.KafkaConfig.SecSkipVerify,
		)
	}

	kafkaProducer, err := kafka.NewProducer(ctx, pConfig)
	if err != nil {
		log.Fatal(ctx, "fatal error trying to create kafka producer", err, log.Data{"topic": cfg.KafkaConfig.FilterConsumerTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.LogErrors(ctx)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		e := scanEvent(scanner)
		log.Info(ctx, "sending filter-submitted event", log.Data{"FilterSubmitted": e})

		bytes, err := schema.FilterSubmittedEvent.Marshal(e)
		if err != nil {
			log.Fatal(ctx, "filter-submitted event error", err)
			os.Exit(1)
		}

		// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
		if err := kafkaProducer.Initialise(ctx); err != nil {
			log.Fatal(ctx, "fatal error trying to initialise kafka producer", err, log.Data{"topic": cfg.KafkaConfig.FilterConsumerTopic})
			os.Exit(1)
		}

		kafkaProducer.Channels().Output <- kafka.BytesMessage{Value: bytes, Context: ctx}
	}
}

// scanEvent creates a FilterSubmitted event according to the user input
func scanEvent(scanner *bufio.Scanner) *event.FilterSubmitted {
	fmt.Println("--- [Send Kafka FilterSubmitted] ---")

	e := &event.FilterSubmitted{}

	fmt.Println("Please type the instance_id")
	fmt.Printf("$ ")
	scanner.Scan()
	e.InstanceID = scanner.Text()

	fmt.Println("Please type the dataset_id")
	fmt.Printf("$ ")
	scanner.Scan()
	e.DatasetID = scanner.Text()

	fmt.Println("Please type the edition")
	fmt.Printf("$ ")
	scanner.Scan()
	e.Edition = scanner.Text()

	fmt.Println("Please type the version")
	fmt.Printf("$ ")
	scanner.Scan()
	e.Version = scanner.Text()

	return e

}
