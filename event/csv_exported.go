package event

// CSVExported provides event data for a single exported CSV
type CSVExported struct {
	FilterJobID string `avro:"filter_job_id"`
}
