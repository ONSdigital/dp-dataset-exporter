package filter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"

	"github.com/ONSdigital/log.go/v2/log"
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

// Event represents a structure with type and time
type Event struct {
	Type string    `bson:"type,omitempty" json:"type"`
	Time time.Time `bson:"time,omitempty" json:"time"`
}

// ErrFilterJobNotFound returned when the filter job count not be found for the given ID.
var ErrFilterJobNotFound = errors.New("failed to find filter job")

// ErrFilterAPIError returned when an unrecognised error occurs.
var ErrFilterAPIError = errors.New("internal error from the filter api")

// ErrUnrecognisedAPIError returned when an unrecognised error occurs.
var ErrUnrecognisedAPIError = errors.New("unrecognised error from the filter api")

// NewStore returns a new instance of a filter store.
func NewStore(cli Client, serviceAuthToken string) *Store {
	return &Store{cli, serviceAuthToken}
}

// PutCSVData allows the filtered file data to be sent back to the filter store when complete.
func (store *Store) PutCSVData(ctx context.Context, filterJobID string, csv filter.Download) error {

	// Add the CSV file to the filter job, the filter api will update the state when all formats are completed
	putBody := filter.Model{
		Downloads: map[string]filter.Download{
			"CSV": {
				URL:     csv.URL,
				Private: csv.Private,
				Public:  csv.Public,
				Size:    csv.Size,
			},
		},
	}

	return store.updateFilterOutput(ctx, filterJobID, putBody)
}

// PutStateAsEmpty sets the filter output as empty
func (store *Store) PutStateAsEmpty(ctx context.Context, filterJobID string) error {
	putBody := filter.Model{
		State: "completed",
	}

	return store.updateFilterOutput(ctx, filterJobID, putBody)
}

// PutStateAsError set the filter output as an error, we shouldn't state the type of error as
// this will be displayed to the public
func (store *Store) PutStateAsError(ctx context.Context, filterJobID string) error {
	putBody := filter.Model{
		State: "failed",
	}

	return store.updateFilterOutput(ctx, filterJobID, putBody)
}

func (store *Store) updateFilterOutput(ctx context.Context, filterJobID string, filter filter.Model) error {

	payload, err := json.Marshal(filter)
	if err != nil {
		return err
	}

	err = store.UpdateFilterOutputBytes(ctx, "", store.serviceAuthToken, "", filterJobID, payload)
	return handleInvalidFilterAPIResponse(ctx, err)
}

// GetFilter returns filter data from the filter API for the given ID
func (store *Store) GetFilter(ctx context.Context, filterOutputID string) (filter *filter.Model, err error) {

	bytes, err := store.GetOutputBytes(ctx, "", store.serviceAuthToken, "", "", filterOutputID)
	if err != nil {
		return nil, handleInvalidFilterAPIResponse(ctx, err)
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
func handleInvalidFilterAPIResponse(ctx context.Context, err error) error {
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
			log.Info(ctx, "unrecognised status code returned from the filter api",
				log.Data{
					"status_code": statusErr.Code(),
				})
			return ErrUnrecognisedAPIError
		}
	}
	return err
}
