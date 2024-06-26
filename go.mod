module github.com/ONSdigital/dp-dataset-exporter

go 1.22

toolchain go1.22.0

// to avoid 'sonatype-2021-4899' non-CVE Vulnerability
exclude github.com/gorilla/sessions v1.2.1

// to avoid 'sonatype-2020-1055' non-CVE vulnerability introduced by github.com/go-ldap/ldap/v3@v3.1.10
replace github.com/go-ldap/ldap/v3 v3.1.10 => github.com/go-ldap/ldap/v3 v3.4.3

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.11.1

//to avoid  [CVE-2022-29153] CWE-918: Server-Side Request Forgery (SSRF)
exclude github.com/hashicorp/consul/api v1.1.0

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.260.1
	github.com/ONSdigital/dp-graph/v2 v2.17.0
	github.com/ONSdigital/dp-healthcheck v1.6.3
	github.com/ONSdigital/dp-net v1.5.0
	github.com/ONSdigital/dp-otel-go v0.0.7
	github.com/ONSdigital/dp-s3 v1.10.0
	github.com/ONSdigital/dp-vault v1.3.1
	github.com/ONSdigital/go-ns v0.0.0-20210916104633-ac1c1c52327e
	github.com/ONSdigital/log.go/v2 v2.4.3
	github.com/aws/aws-sdk-go v1.54.8
	github.com/gorilla/mux v1.8.1
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/vault/api v1.14.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.8.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.52.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.52.0
	golang.org/x/time v0.5.0 // indirect
)

require github.com/ONSdigital/dp-kafka/v4 v4.1.0

require (
	github.com/ONSdigital/dp-api-clients-go v1.43.0 // indirect
	github.com/ONSdigital/dp-net/v2 v2.11.2 // indirect
	github.com/ONSdigital/golang-neo4j-bolt-driver v0.0.0-20210408132126-c2323ff08bf1 // indirect
	github.com/ONSdigital/graphson v0.3.0 // indirect
	github.com/ONSdigital/gremgo-neptune v1.1.0 // indirect
	github.com/Shopify/sarama v1.38.1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/eapache/go-resiliency v1.6.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-avro/avro v0.0.0-20171219232920-444163702c11 // indirect
	github.com/go-jose/go-jose/v4 v4.0.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-test/deep v1.1.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.8 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.6 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/smarty/assertions v1.16.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama v0.43.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.52.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.27.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.27.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.27.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.27.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/sdk v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240624140628-dc46fd24d27d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240624140628-dc46fd24d27d // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
