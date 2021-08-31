package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"github.com/ONSdigital/dp-dataset-exporter/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

// The purpose of this app is to trigger various filters in order to time how long they take
// my examining the logs

func main() {
	ctx := context.Background()
	log.Namespace = "dp-dataset-exporter"

	config, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error getting config", err)
		os.Exit(1)
	}

	// Avoid logging the neo4j FileURL as it may contain a password
	log.Info(ctx, "loaded config", log.Data{"config": config})

	// Create Kafka Producer
	pChannels := kafka.CreateProducerChannels()
	pConfig := &kafka.ProducerConfig{KafkaVersion: &config.KafkaVersion}
	kafkaProducer, err := kafka.NewProducer(ctx, config.KafkaAddr, config.FilterConsumerTopic, pChannels, pConfig)
	if err != nil {
		log.Fatal(ctx, "fatal error trying to create kafka producer", err, log.Data{"topic": config.FilterConsumerTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.Channels().LogErrors(ctx, "kafka producer")

	type KafakaQuerry struct {
		InstanceID string `avro:"instance_id"`
		DatasetID  string `avro:"dataset_id"`
		Edition    string `avro:"edition"`
		Version    string `avro:"version"`
	}

	filterItems := []KafakaQuerry{
		/*{
			"08eda1d1-df67-499a-a545-b5223292b19f", // this works ?, 1 => no, not in acceptable state
			"cpih01",
			"time-series",
			"1",
		},*/
		{
			"460b5039-bb09-4038-b8eb-9091713f4497", // this works ?, 7
			"older-people-economic-activity",
			"time-series",
			"1",
		},
		{
			"8b6e93bb-d40d-4370-b75f-a25aecf9fb7f", // this works ?, 8
			"older-people-economic-activity",
			"time-series",
			"1",
		},
		/*{
			"c3a5bf00-c975-46f9-83e9-748f5076e47a", // this works ? ~ 30 seconds 136
			"life-expectancy-by-local-authority",
			"time-series",
			"2",
		},*/
		{
			"d49aaa19-1daa-43d4-9861-dbe88d5b9d54", // this works ? 151
			"suicides-in-the-uk",
			"2019",
			"2",
		},
		{
			"e9cbd134-9066-4a05-a366-b8573ebb53d0", // this works ? 209
			"2011-census-1pc-6var",
			"2011",
			"1",
		},
		{
			"85578044-27a2-4126-8e98-176a92a6a54c", // this works ? 243
			"suicides-in-the-uk",
			"time-series",
			"1",
		},
		{
			"b6fb647b-3dd8-4b2e-98bb-533577580938", // this works ? 244
			"suicides-in-the-uk",
			"time-series",
			"2",
		},
		{
			"92b35c5d-ea94-4622-be87-c25d672faf3e", // this works ? 248
			"wellbeing-quarterly",
			"time-series",
			"2",
		},
	}

	var sendCount int

	for i := 0; i < 1; i++ {
		var count int
		// launch a load (possibly: 269) of filter requests to do performance timings.
		for _, item := range filterItems {
			sendFilter(ctx, kafkaProducer, item.InstanceID, item.DatasetID, item.Edition, item.Version)
			count++
			if count >= 1 {
				break
			}
			sendCount++
			fmt.Printf("sendCount: %d\n", sendCount)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		filterID := scanner.Text()

		sendFilter(ctx, kafkaProducer, filterID, "1", "2", "3")
		time.Sleep(1 * time.Second)
	}
}

func sendFilter(ctx context.Context, kafkaProducer *kafka.Producer, filterID, datasetID, edition, version string) {
	log.Info(ctx, "sending filter output event", log.Data{"filter_ouput_id": filterID})

	event := event.FilterSubmitted{
		FilterID:  filterID,
		DatasetID: datasetID,
		Edition:   edition,
		Version:   version,
	}

	bytes, err := schema.FilterSubmittedEvent.Marshal(event)
	if err != nil {
		log.Fatal(ctx, "filter submitted event error", err)
		os.Exit(1)
	}

	// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
	kafkaProducer.Initialise(ctx)
	kafkaProducer.Channels().Output <- bytes
}
