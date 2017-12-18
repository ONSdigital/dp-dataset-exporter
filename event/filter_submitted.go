package event

// FilterSubmitted is the structure of each event consumed.
type FilterSubmitted struct {
	FilterID   string `avro:"filter_output_id"`
	InstanceID string `avro:"instance_id"`
	DatasetID  string `avro:"dataset_id"`
	Edition    string `avro:"edition"`
	Version    string `avro:"version"`
}
