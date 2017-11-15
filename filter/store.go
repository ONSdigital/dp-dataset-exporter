package filter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/go-ns/log"
	"strconv"
)

//go:generate moq -out filtertest/http_client.go -pkg filtertest . HTTPClient

// Store provides access to stored dimension data.
type Store struct {
	filterAPIURL       string
	filterAPIAuthToken string
	httpClient         HTTPClient
}

// FilterOuput represents a structure used to update the filter api
type FilterOuput struct {
	FilterID   string     `json:"filter_id,omitempty"`
	InstanceID string     `json:"instance_id"`
	State      string     `json:"state,omitempty"`
	Downloads  *Downloads `json:"downloads,omitempty"`
}

// Downloads represents the CSV which has been generated
type Downloads struct {
	CSV *DownloadItem `json:"csv,omitempty"`
}

// DownloadItem represents an object containing information for the download item
type DownloadItem struct {
	Size string `json:"size,omitempty"`
	URL  string `json:"url,omitempty"`
}

// ErrFilterJobNotFound returned when the filter job count not be found for the given ID.
var ErrFilterJobNotFound = errors.New("failed to find filter job")

// ErrFilterAPIError returned when an unrecognised error occurs.
var ErrFilterAPIError = errors.New("Internal error from the filter api")

// ErrUnrecognisedAPIError returned when an unrecognised error occurs.
var ErrUnrecognisedAPIError = errors.New("Unrecognised error from the filter api")

// NewStore returns a new instance of a filter store.
func NewStore(filterAPIURL string, filterAPIAuthToken string, httpClient HTTPClient) *Store {
	return &Store{
		filterAPIURL:       filterAPIURL,
		filterAPIAuthToken: filterAPIAuthToken,
		httpClient:         httpClient,
	}
}

// HTTPClient dependency.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// PutCSVData allows the filtered file data to be sent back to the filter store when complete.
func (store *Store) PutCSVData(filterJobID string, fileURL string, size int64) error {

	url := store.filterAPIURL + "/filter-outputs/" + filterJobID

	// Add the CSV file to the filter job, the filter api will update the state when all formats are completed
	putBody := &FilterOuput{
		Downloads: &Downloads{
			CSV: &DownloadItem{
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
	filter, err := store.getFilterData(filterJobID)
	if err != nil {
		return nil, err
	}

	return filter, nil
}

// call the filter API for filter data.
func (store *Store) getFilterData(filterOutputID string) (*observation.Filter, error) {

	url := store.filterAPIURL + "/filter-outputs/" + filterOutputID

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

// common function for handling a HTTP request and response codes.
func (store *Store) makeRequest(method, url string, body io.Reader) ([]byte, error) {

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Internal-Token", store.filterAPIAuthToken)

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
		log.Debug("unrecognised status code returned from the filter api",
			log.Data{
				"status_code": response.StatusCode,
			})
		return nil, ErrUnrecognisedAPIError
	}
}
