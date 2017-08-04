package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var filterJobSubmittedEvent = `{
  "type": "record",
  "name": "filter-job-submitted",
  "namespace": "",
  "fields": [
    {"name": "filter_job_id", "type": "string"}
  ]
}`

// FilterJobSubmittedEvent the Avro schema for FilterJobSubmitted messages.
var FilterJobSubmittedEvent *avro.Schema = &avro.Schema{
	Definition: filterJobSubmittedEvent,
}