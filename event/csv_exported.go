package event

// CSVExported provides event data for a single exported CSV
type CSVExported struct {
	FilterID   string `avro:"filter_output_id"`
	FileURL    string `avro:"file_url"`
	InstanceID string `avro:"instance_id"`
	DatasetID  string `avro:"dataset_id"`
	Edition    string `avro:"edition"`
	Version    string `avro:"version"`
	Filename   string `avro:"filename"`
}
