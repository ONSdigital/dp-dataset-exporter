package observation

// Filter represents a structure for a filter job
type Filter struct {
	FilterID         string             `json:"filter_id,omitempty"`
	InstanceID       string             `json:"instance_id"`
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
	Name    string   `json:"name,omitempty"`
	Options []string `json:"options,omitempty"`
}

// DownloadItem represents an object containing information for the download item
type DownloadItem struct {
	Size string `json:"size,omitempty"`
	URL  string `json:"url,omitempty"`
}
