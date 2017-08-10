package observation

// Filter represents a structure for a filter job
type Filter struct {
	JobID            string             `json:"filter_job_id,omitempty"`
	DataSetFilterID  string             `json:"dataset_filter_id"`
	DimensionListURL string             `json:"dimension_list_url"`
	State            string             `json:"state,omitempty"`
	DimensionFilters []*DimensionFilter `json:"dimensions,omitempty"`
}

// DimensionFilter represents an object containing a list of dimension values and the dimension name
type DimensionFilter struct {
	Name   string   `json:"name,omitempty"`
	URL    string   `json:"dimension_url,omitempty"`
	Values []string `json:"values,omitempty"`
}
