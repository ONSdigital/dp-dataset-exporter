package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var filterSubmittedEvent = `{
  "type": "record",
  "name": "filter-output-submitted",
  "fields": [
    {"name": "filter_output_id", "type": "string", "default": ""},
    {"name": "instance_id", "type": "string", "default": ""},
    {"name": "dataset_id", "type": "string", "default": ""},
    {"name": "edition", "type": "string", "default": ""},
    {"name": "version", "type": "string", "default": ""}
  ]
}`

// FilterSubmittedEvent the Avro schema for FilterOutputSubmitted messages.
var FilterSubmittedEvent = &avro.Schema{
	Definition: filterSubmittedEvent,
}

var csvExportedEvent = `{
  "type": "record",
  "name": "csv-exported",
  "fields": [
    {"name": "filter_output_id", "type": "string", "default": ""},
    {"name": "file_url", "type": "string", "default": ""},
    {"name": "instance_id", "type": "string", "default": ""},
    {"name": "dataset_id", "type": "string", "default": ""},
    {"name": "edition", "type": "string", "default": ""},
    {"name": "version", "type": "string", "default": ""},
    {"name": "filename", "type": "string", "default": ""},
    {"name": "row_count", "type": "int", "default": 0}
  ]
}`

// CSVExportedEvent the Avro schema for CSV exported messages.
var CSVExportedEvent = &avro.Schema{
	Definition: csvExportedEvent,
}
