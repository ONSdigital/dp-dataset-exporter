package event

// CSVExported provides event data for a single exported CSV
type CSVExported struct {
	FilterID string `avro:"filter_output_id"`
	FileURL  string `avro:"file_url"`
}
