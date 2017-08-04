package observation

// Filter represents a structure for a filter job
type Filter struct {
	DataSetFilterID  string            `json:"dataset_filter_id"`
	FilterID         string            `json:"id,omitempty"`
	State            string            `json:"state,omitempty"`
	DimensionFilters []DimensionFilter `json:"dimensions,omitempty"`
}

// DimensionFilter represents an object containing a list of dimension values and the dimension name
type DimensionFilter struct {
	Name   string   `json:"name,omitempty"`
	Values []string `json:"values,omitempty"`
}
