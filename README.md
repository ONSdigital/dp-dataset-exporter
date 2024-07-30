# dp-dataset-exporter

Takes a filter job and produces a filtered dataset.

## Getting started

Ensure you have vault running.

`brew install vault`
`vault server -dev`

* Setup AWS credentials. The app uses the default provider chain. When running locally this typically means they are provided by the `~/.aws/credentials` file.  Alternatively you can inject the credentials via environment variables as described in the configuration section
* Run `make debug`
* Run the auth-stub-api

### Kafka scripts

Scripts for updating and debugging Kafka can be found [here](https://github.com/ONSdigital/dp-data-tools)(dp-data-tools)

### Configuration

An overview of the configuration options available, either as a table of
environment variables, or with a link to a configuration guide.

| Environment variable         | Default (example)                    | Description
| ---------------------------- | ------------------------------------ | -----------
| BIND_ADDR                    | :22500                               | The host and port to bind to
| KAFKA_ADDR                   | `localhost:9092`                     | The address of (TLS-ready) Kafka brokers (comma-separated values)
| KAFKA_VERSION                | `1.0.2`                              | The version of (TLS-ready) Kafka
| KAFKA_SEC_PROTO              | _unset_               (only `TLS`)   | if set to `TLS`, kafka connections will use TLS
| KAFKA_SEC_CLIENT_KEY         | _unset_                              | PEM [2] for the client key (optional, used for client auth) [1]
| KAFKA_SEC_CLIENT_CERT        | _unset_                              | PEM [2] for the client certificate (optional, used for client auth) [1]
| KAFKA_SEC_CA_CERTS           | _unset_                              | PEM [2] of CA cert chain if using private CA for the server cert [1]
| KAFKA_SEC_SKIP_VERIFY        | false                                | ignore server certificate issues if set to `true` [1]
| FILTER_JOB_CONSUMER_TOPIC    | filter-job-submitted                 | The name of the topic to consume messages from
| FILTER_JOB_CONSUMER_GROUP    | dp-dataset-exporter                  | The consumer group this application to consume filter job messages
| DATASET_API_URL              | http://localhost:22000               | The URL of the dataset API
| DATASET_API_AUTH_TOKEN       | FD0108EA-825D-411C-9B1D-41EF7727F465 | The auth token for the dataset API
| FILTER_API_URL               | http://localhost:22100               | The URL of the filter API
| FILTER_API_AUTH_TOKEN        | FD0108EA-825D-411C-9B1D-41EF7727F465 | The auth token for the filter API
| AWS_REGION                   | eu-west-1                            | The AWS region to use
| S3_BUCKET_URL                | _unset_     (e.g. `https://cf.host`) | If set, the URL prefix for public, exported downloads
| S3_BUCKET_NAME               | csv-exported                         | The name of the public S3 bucket to store exported files
| S3_PRIVATE_BUCKET_NAME       | csv-exported                         | The name of the private s3 bucket to store exported files
| CSV_EXPORTED_PRODUCER_TOPIC  | common-output-created                | The topic to add messages to when a job is complete
| ERROR_PRODUCER_TOPIC         | filter-error                         | The topic to add messages to when an error occurs
| HEALTHCHECK_INTERVAL         | 30s                                  | The time between doing health checks
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                                  | The time taken for the health changes from warning state to critical due to subsystem check failures
| GRACEFUL_SHUTDOWN_TIMEOUT    | time.Second * 10                     | How long to wait for the service to shutdown gracefully
| DOWNLOAD_SERVICE_URL         | http://localhost:23600               | The URL of the download service
| SERVICE_AUTH_TOKEN           | 0f49d57b-c551-4d33-af1e-a442801dd851 | The service token for this app
| ZEBEDEE_URL                  | http://localhost:8082                | The URL to zebedee
| AWS_ACCESS_KEY_ID            | -                                    | The AWS access key credential
| AWS_SECRET_ACCESS_KEY        | -                                    | The AWS secret key credential
| FULL_DATASET_FILE_PREFIX     | full-datasets                        | The prefix added to full dataset download files
| FILTERED_DATASET_FILE_PREFIX | filtered-dataset                     | The prefix added to filtered dataset download files
| OTEL_EXPORTER_OTLP_ENDPOINT  | localhost:4317                       | Endpoint for OpenTelemetry service
| OTEL_SERVICE_NAME            | dp-dataset-exporter                  | Label of service for OpenTelemetry service
| OTEL_BATCH_TIMEOUT           | 5s                                   | Timeout for OpenTelemetry
| OTEL_ENABLED                 | false                                | Feature flag to enable OpenTelemetry

Notes:

1. Ignored unless using TLS (i.e. `KAFKA_SEC_PROTO` has a value enabling TLS)

2. PEM values are identified as those starting with `-----BEGIN`
    and can use `\n` (sic) instead of newlines (they will be converted to newlines before use).
    Any other value will be treated as a path to the given PEM file.

### Healthcheck

 The `/health` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:22500/health`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2019, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
