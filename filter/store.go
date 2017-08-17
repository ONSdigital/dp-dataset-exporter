package filter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-dataset-exporter/observation"
)

//go:generate moq -out filtertest/http_client.go -pkg filtertest . HTTPClient

// Store provides access to stored dimension data.
type Store struct {
	filterAPIURL string
	httpClient   HTTPClient
}

// ErrFilterJobNotFound returned when the filter job count not be found for the given ID.
var ErrFilterJobNotFound = errors.New("failed to find filter job")

// ErrFilterAPIError returned when an unrecognised error occurs.
var ErrFilterAPIError = errors.New("Internal error from the import api")

// ErrUnrecognisedAPIError returned when an unrecognised error occurs.
var ErrUnrecognisedAPIError = errors.New("Unrecognised error from the import api")

// NewStore returns a new instance of a filter store.
func NewStore(filterAPIURL string, httpClient HTTPClient) *Store {
	return &Store{
		filterAPIURL: filterAPIURL,
		httpClient:   httpClient,
	}
}

// HTTPClient dependency.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// PutCSVData allows the filtered file data to be sent back to the filter store when complete.
func (store *Store) PutCSVData(filterJobID string, fileURL string, size int64) error {

	url := store.filterAPIURL + "/filters/" + filterJobID

	putBody := &observation.Filter{
		State: "completed",
		Downloads: &observation.Downloads{
			CSV: &observation.DownloadItem{
				Size: strconv.FormatInt(size, 10),
				URL:  fileURL,
			},
		},
	}

	json, err := json.Marshal(putBody)
	if err != nil {
		return err
	}

	_, err = store.makeRequest("PUT", url, bytes.NewReader(json))

	return err
}

// GetFilter returns filter data from the filter API for the given ID
func (store *Store) GetFilter(filterJobID string) (*observation.Filter, error) {
	filter, err := store.getFilterMetaData(filterJobID)
	if err != nil {
		return nil, err
	}

	dimensionList, err := store.getFilterDimensionList(filter)
	if err != nil {
		return nil, err
	}

	for _, dimension := range dimensionList {

		dimensionFilter, err := store.getFilterDimensionOptions(dimension.URL)
		if err != nil {
			return nil, err
		}

		filter.DimensionFilters = append(filter.DimensionFilters, dimensionFilter)
	}

	return filter, nil
}

// call the filter API for filter meta data.
func (store *Store) getFilterMetaData(filterJobID string) (*observation.Filter, error) {

	url := store.filterAPIURL + "/filters/" + filterJobID

	bytes, err := store.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var filter *observation.Filter
	err = json.Unmarshal(bytes, &filter)
	if err != nil {
		return nil, err
	}

	return filter, nil
}

// call the filter API for filter meta data.
func (store *Store) getFilterDimensionOptions(dimensionURL string) (*observation.DimensionFilter, error) {
	dimensionOptionsURL := store.filterAPIURL + dimensionURL + "/options"

	bytes, err := store.makeRequest("GET", dimensionOptionsURL, nil)
	if err != nil {
		return nil, err
	}

	var dimensionFilter *observation.DimensionFilter
	err = json.Unmarshal(bytes, &dimensionFilter)
	if err != nil {
		return nil, err
	}

	return dimensionFilter, nil
}

func (store *Store) getFilterDimensionList(filter *observation.Filter) ([]*observation.DimensionFilter, error) {
	url := store.filterAPIURL + filter.DimensionListURL

	bytes, err := store.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var dimensionURLs []*observation.DimensionFilter
	err = json.Unmarshal(bytes, &dimensionURLs)
	if err != nil {
		return nil, err
	}

	return dimensionURLs, nil
}

// common function for handling a HTTP request and response codes.
func (store *Store) makeRequest(method, url string, body io.Reader) ([]byte, error) {

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	response, responseError := store.httpClient.Do(request)
	if responseError != nil {
		return nil, responseError
	}

	switch response.StatusCode {
	case http.StatusOK:
		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read http body into bytes")
		}
		return bytes, nil
	case http.StatusNotFound:
		return nil, ErrFilterJobNotFound
	case http.StatusInternalServerError:
		return nil, ErrFilterAPIError
	default:
		return nil, ErrUnrecognisedAPIError
	}
}
