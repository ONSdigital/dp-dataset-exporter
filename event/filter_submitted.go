package event

// FilterSubmitted is the structure of each event consumed.
type FilterSubmitted struct {
	FilterID string `avro:"filter_output_id"`
}
