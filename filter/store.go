package filter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/filter"

	"github.com/ONSdigital/dp-graph/observation"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out filtertest/client.go -pkg filtertest . Client

// Client interface contains the method signatures expected to be implemented by dp-api-clients-go/filter
type Client interface {
	UpdateFilterOutputBytes(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, filterJobID string, b []byte) error
	GetOutputBytes(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) ([]byte, error)
}

// Store provides access to stored dimension data using the filter dp-api-client
type Store struct {
	Client
	serviceAuthToken string
}

// FilterOutput represents a structure used to update the filter api
type FilterOutput struct {
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
func NewStore(cli Client, serviceAuthToken string) *Store {
	return &Store{cli, serviceAuthToken}
}

// PutCSVData allows the filtered file data to be sent back to the filter store when complete.
func (store *Store) PutCSVData(filterJobID string, csv observation.DownloadItem) error {

	// Add the CSV file to the filter job, the filter api will update the state when all formats are completed
	putBody := FilterOutput{
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
	putBody := FilterOutput{
		State: "completed",
	}

	return store.updateFilterOutput(filterJobID, &putBody)
}

// PutStateAsError set the filter output as an error, we shouldn't state the type of error as
// this will be displayed to the public
func (store *Store) PutStateAsError(filterJobID string) error {
	putBody := FilterOutput{
		State: "failed",
	}

	return store.updateFilterOutput(filterJobID, &putBody)
}

func (store *Store) updateFilterOutput(filterJobID string, filter *FilterOutput) error {

	payload, err := json.Marshal(filter)
	if err != nil {
		return err
	}

	err = store.UpdateFilterOutputBytes(context.Background(), "", store.serviceAuthToken, "", filterJobID, payload)
	return handleInvalidFilterAPIResponse(err)
}

// GetFilter returns filter data from the filter API for the given ID
func (store *Store) GetFilter(filterOutputID string) (filter *observation.Filter, err error) {

	bytes, err := store.GetOutputBytes(context.Background(), "", store.serviceAuthToken, "", "", filterOutputID)
	if err != nil {
		return nil, handleInvalidFilterAPIResponse(err)
	}

	if err = json.Unmarshal(bytes, &filter); err != nil {
		return nil, err
	}
	return filter, nil
}

// handleInvalidFilterAPIResponse checks if the error type is ErrInvalidFilterAPIResponse,
// and if that is the case, it translates the code to the required Error type:
// - StatusNotFound -> ErrFilterJobNotFound
// - StatusInternalServerError -> ErrFilterAPIError
// - Any other status -> ErrUnrecognisedAPIError
func handleInvalidFilterAPIResponse(err error) error {
	if err == nil {
		return nil
	}
	if statusErr, ok := err.(*filter.ErrInvalidFilterAPIResponse); ok {
		switch statusErr.Code() {
		case http.StatusNotFound:
			return ErrFilterJobNotFound
		case http.StatusInternalServerError:
			return ErrFilterAPIError
		default:
			log.Event(nil, "unrecognised status code returned from the filter api", log.INFO,
				log.Data{
					"status_code": statusErr.Code(),
				})
			return ErrUnrecognisedAPIError
		}
	}
	return err
}
