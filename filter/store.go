package filter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-graph/observation"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/log.go/log"
)

// Store provides access to stored dimension data.
type Store struct {
	filterAPIURL string
	httpClient   rchttp.Clienter
}

// FilterOuput represents a structure used to update the filter api
type FilterOuput struct {
	FilterID   string                 `json:"filter_id,omitempty"`
	InstanceID string                 `json:"instance_id"`
	State      string                 `json:"state,omitempty"`
	Downloads  *observation.Downloads `json:"downloads,omitempty"`
	Events     []*Event               `json:"events,omitempty"`
	Published  bool                   `json:"published,omitempty"`
}

// Event represents a structure with type and time
type Event struct {
	Type string    `bson:"type,omitempty" json:"type"`
	Time time.Time `bson:"time,omitempty" json:"time"`
}

// ErrFilterJobNotFound returned when the filter job count not be found for the given ID.
var ErrFilterJobNotFound = errors.New("failed to find filter job")

// ErrFilterAPIError returned when an unrecognised error occurs.
var ErrFilterAPIError = errors.New("Internal error from the filter api")

// ErrUnrecognisedAPIError returned when an unrecognised error occurs.
var ErrUnrecognisedAPIError = errors.New("Unrecognised error from the filter api")

// NewStore returns a new instance of a filter store.
func NewStore(filterAPIURL string, httpClient rchttp.Clienter) *Store {
	return &Store{
		filterAPIURL: filterAPIURL,
		httpClient:   httpClient,
	}
}

// PutCSVData allows the filtered file data to be sent back to the filter store when complete.
func (store *Store) PutCSVData(filterJobID string, csv observation.DownloadItem) error {

	// Add the CSV file to the filter job, the filter api will update the state when all formats are completed
	putBody := FilterOuput{
		Downloads: &observation.Downloads{
			CSV: &observation.DownloadItem{
				HRef:    csv.HRef,
				Private: csv.Private,
				Public:  csv.Public,
				Size:    csv.Size,
			},
		},
	}

	return store.updateFilterOutput(filterJobID, &putBody)
}

// PutStateAsEmpty sets the filter output as empty
func (store *Store) PutStateAsEmpty(filterJobID string) error {
	putBody := FilterOuput{
		State: "completed",
	}

	return store.updateFilterOutput(filterJobID, &putBody)
}

// PutStateAsError set the filter output as an error, we shouldn't state the type of error as
// this will be displayed to the public
func (store *Store) PutStateAsError(filterJobID string) error {
	putBody := FilterOuput{
		State: "failed",
	}

	return store.updateFilterOutput(filterJobID, &putBody)
}

func (store *Store) updateFilterOutput(filterJobID string, filter *FilterOuput) error {
	url := store.filterAPIURL + "/filter-outputs/" + filterJobID
	json, err := json.Marshal(filter)
	if err != nil {
		return err
	}
	_, err = store.makeRequest("PUT", url, bytes.NewReader(json))
	return err
}

// GetFilter returns filter data from the filter API for the given ID
func (store *Store) GetFilter(filterOutputID string) (*observation.Filter, error) {
	filter, err := store.getFilterData(filterOutputID)
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

	response, responseError := store.httpClient.Do(context.Background(), request)
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
		log.Event(nil, "unrecognised status code returned from the filter api",
			log.Data{
				"status_code": response.StatusCode,
			})
		return nil, ErrUnrecognisedAPIError
	}
}
