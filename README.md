dp-dataset-exporter
================

Takes a filter job and produces a filtered dataset.

### Getting started

### Configuration

An overview of the configuration options available, either as a table of
environment variables, or with a link to a configuration guide.

| Environment variable       | Default                | Description
| -------------------------- | ---------------------- | -----------
| BIND_ADDR                  | :22500                 | The host and port to bind to
| KAFKA_ADDR                 | localhost:9092         | The address of Kafka
| FILTER_JOB_CONSUMER_TOPIC  | filter-job-submitted   | The name of the topic to consume messages from
| FILTER_JOB_CONSUMER_GROUP  | dp-dataset-exporter    | The consumer group this application to consume filter job messages
| DATABASE_ADDRESS           | bolt://localhost:7687  | The address of the database to retrieve dataset data from
| FILTER_API_URL             | http://localhost:22100 | The URL of the filter API
| RESULT_PRODUCER_TOPIC      | csv-filtered           | The topic to add messages to when a job is complete

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
