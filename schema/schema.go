package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var filterSubmittedEvent = `{
  "type": "record",
  "name": "filter-output-submitted",
  "namespace": "",
  "fields": [
    {"name": "filter_output_id", "type": "string"},
    {"name": "instance_id", "type": "string"},
    {"name": "dataset_id", "type": "string"},
    {"name": "edition", "type": "string"},
    {"name": "version", "type": "string"}
  ]
}`

// FilterSubmittedEvent the Avro schema for FilterOutputSubmitted messages.
var FilterSubmittedEvent = &avro.Schema{
	Definition: filterSubmittedEvent,
}

var csvExportedEvent = `{
  "type": "record",
  "name": "csv-exported",
  "namespace": "",
  "fields": [
    {"name": "filter_output_id", "type": "string"},
    {"name": "file_url", "type": "string"},
    {"name": "instance_id", "type": "string"},
    {"name": "dataset_id", "type": "string"},
    {"name": "edition", "type": "string"},
    {"name": "version", "type": "string"},
    {"name": "filename", "type": "string"}
  ]
}`

// CSVExportedEvent the Avro schema for CSV exported messages.
var CSVExportedEvent = &avro.Schema{
	Definition: csvExportedEvent,
}
