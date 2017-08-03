package event

// FilterJobSubmitted is the structure of each event consumed.
type FilterJobSubmitted struct {
	FilterJobID string `avro:"filter_job_id"`
}
