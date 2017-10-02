package errors

// Event to contain details of something that happened in the filter / export process.
type Event struct {
	FilterID  string `avro:"filters_id"`
	EventType string `avro:"event_type"`
	EventMsg  string `avro:"event_message"`
}
