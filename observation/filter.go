package observation

// Filter represents a structure for a filter job
type Filter struct {
	JobID            string             `json:"filter_job_id,omitempty"`
	DataSetFilterID  string             `json:"dataset_filter_id"`
	DimensionListURL string             `json:"dimension_list_url"`
	State            string             `json:"state,omitempty"`
	Downloads        *Downloads         `json:"downloads,omitempty"`
	DimensionFilters []*DimensionFilter `json:"dimensions,omitempty"`
}

// Downloads represents a list of file types possible to download
type Downloads struct {
	CSV *DownloadItem `json:"csv,omitempty"`
}

// DimensionFilter represents an object containing a list of dimension values and the dimension name
type DimensionFilter struct {
	Name    string             `json:"name,omitempty"`
	URL     string             `json:"dimension_url,omitempty"`
	Options []*DimensionOption `json:"options,omitempty"`
}

// DimensionOption represents a single option of a dimension
type DimensionOption struct {
	Option string `json:"option,omitempty"`
	URL    string `json:"dimension_option_url,omitempty"`
}

// DownloadItem represents an object containing information for the download item
type DownloadItem struct {
	Size string `json:"size,omitempty"`
	URL  string `json:"url,omitempty"`
}
