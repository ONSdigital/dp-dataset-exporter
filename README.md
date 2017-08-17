dp-dataset-exporter
================

Takes a filter job and produces a filtered dataset.

### Getting started

### Configuration

An overview of the configuration options available, either as a table of
environment variables, or with a link to a configuration guide.

| Environment variable        | Default                              | Description
| --------------------------- | ------------------------------------ | -----------
| BIND_ADDR                   | :22500                               | The host and port to bind to
| KAFKA_ADDR                  | localhost:9092                       | The address of Kafka
| FILTER_JOB_CONSUMER_TOPIC   | filter-job-submitted                 | The name of the topic to consume messages from
| FILTER_JOB_CONSUMER_GROUP   | dp-dataset-exporter                  | The consumer group this application to consume filter job messages
| DATABASE_ADDRESS            | bolt://localhost:7687                | The address of the database to retrieve dataset data from
| FILTER_API_URL              | http://localhost:22100               | The URL of the filter API
| FILTER_API_AUTH_TOKEN       | FD0108EA-825D-411C-9B1D-41EF7727F465 | The auth token for the filter API
| AWS_REGION                  | eu-west-1                            | The AWS region to use
| S3_BUCKET_NAME              | http://localhost:22100               | The name of the S3 bucket to store exported files
| CSV_EXPORTED_PRODUCER_TOPIC | csv-exported                         | The topic to add messages to when a job is complete

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
