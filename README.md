dp-dataset-exporter
================

Takes a filter job and produces a filtered dataset.

### Getting started

`make debug`

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
| DATASET_API_URL             | http://localhost:22000               | The URL of the dataset API
| DATASET_API_AUTH_TOKEN      | FD0108EA-825D-411C-9B1D-41EF7727F465 | The auth token for the dataset API
| NEO4J_POOL_SIZE             | 5                                    | The number of neo4j connections to pool
| FILTER_API_URL              | http://localhost:22100               | The URL of the filter API
| FILTER_API_AUTH_TOKEN       | FD0108EA-825D-411C-9B1D-41EF7727F465 | The auth token for the filter API
| AWS_REGION                  | eu-west-1                            | The AWS region to use
| S3_BUCKET_NAME              | http://localhost:22100               | The name of the public S3 bucket to store exported files
| S3_PRIVATE_BUCKET_NAME      | csv-exported                         | The name of the private s3 bucket to store exported files
| CSV_EXPORTED_PRODUCER_TOPIC | csv-exported                         | The topic to add messages to when a job is complete
| ERROR_PRODUCER_TOPIC        | filter-error                         | The topic to add messages to when an error occurs
| HEALTHCHECK_INTERVAL        | time.Minute                          | How often to run a health check
| GRACEFUL_SHUTDOWN_TIMEOUT   | time.Second * 10                     | How long to wait for the service to shutdown gracefully
| DOWNLOAD_SERVICE_URL        | http://localhost:23600               | The URL of the download service

### Healthcheck

 The `/healthcheck` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:22500/healthcheck`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
