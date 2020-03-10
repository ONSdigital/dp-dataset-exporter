// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package filtertest

import (
	"context"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"sync"
)

var (
	lockClientMockGetOutputBytes          sync.RWMutex
	lockClientMockUpdateFilterOutputBytes sync.RWMutex
)

// Ensure, that ClientMock does implement filter.Client.
// If this is not the case, regenerate this file with moq.
var _ filter.Client = &ClientMock{}

// ClientMock is a mock implementation of filter.Client.
//
//     func TestSomethingThatUsesClient(t *testing.T) {
//
//         // make and configure a mocked filter.Client
//         mockedClient := &ClientMock{
//             GetOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, collectionID string, filterOutputID string) ([]byte, error) {
// 	               panic("mock out the GetOutputBytes method")
//             },
//             UpdateFilterOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, filterJobID string, b []byte) error {
// 	               panic("mock out the UpdateFilterOutputBytes method")
//             },
//         }
//
//         // use mockedClient in code that requires filter.Client
//         // and then make assertions.
//
//     }
type ClientMock struct {
	// GetOutputBytesFunc mocks the GetOutputBytes method.
	GetOutputBytesFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, collectionID string, filterOutputID string) ([]byte, error)

	// UpdateFilterOutputBytesFunc mocks the UpdateFilterOutputBytes method.
	UpdateFilterOutputBytesFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, filterJobID string, b []byte) error

	// calls tracks calls to the methods.
	calls struct {
		// GetOutputBytes holds details about calls to the GetOutputBytes method.
		GetOutputBytes []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// DownloadServiceToken is the downloadServiceToken argument value.
			DownloadServiceToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// FilterOutputID is the filterOutputID argument value.
			FilterOutputID string
		}
		// UpdateFilterOutputBytes holds details about calls to the UpdateFilterOutputBytes method.
		UpdateFilterOutputBytes []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// DownloadServiceToken is the downloadServiceToken argument value.
			DownloadServiceToken string
			// FilterJobID is the filterJobID argument value.
			FilterJobID string
			// B is the b argument value.
			B []byte
		}
	}
}

// GetOutputBytes calls GetOutputBytesFunc.
func (mock *ClientMock) GetOutputBytes(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, collectionID string, filterOutputID string) ([]byte, error) {
	if mock.GetOutputBytesFunc == nil {
		panic("ClientMock.GetOutputBytesFunc: method is nil but Client.GetOutputBytes was just called")
	}
	callInfo := struct {
		Ctx                  context.Context
		UserAuthToken        string
		ServiceAuthToken     string
		DownloadServiceToken string
		CollectionID         string
		FilterOutputID       string
	}{
		Ctx:                  ctx,
		UserAuthToken:        userAuthToken,
		ServiceAuthToken:     serviceAuthToken,
		DownloadServiceToken: downloadServiceToken,
		CollectionID:         collectionID,
		FilterOutputID:       filterOutputID,
	}
	lockClientMockGetOutputBytes.Lock()
	mock.calls.GetOutputBytes = append(mock.calls.GetOutputBytes, callInfo)
	lockClientMockGetOutputBytes.Unlock()
	return mock.GetOutputBytesFunc(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID)
}

// GetOutputBytesCalls gets all the calls that were made to GetOutputBytes.
// Check the length with:
//     len(mockedClient.GetOutputBytesCalls())
func (mock *ClientMock) GetOutputBytesCalls() []struct {
	Ctx                  context.Context
	UserAuthToken        string
	ServiceAuthToken     string
	DownloadServiceToken string
	CollectionID         string
	FilterOutputID       string
} {
	var calls []struct {
		Ctx                  context.Context
		UserAuthToken        string
		ServiceAuthToken     string
		DownloadServiceToken string
		CollectionID         string
		FilterOutputID       string
	}
	lockClientMockGetOutputBytes.RLock()
	calls = mock.calls.GetOutputBytes
	lockClientMockGetOutputBytes.RUnlock()
	return calls
}

// UpdateFilterOutputBytes calls UpdateFilterOutputBytesFunc.
func (mock *ClientMock) UpdateFilterOutputBytes(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, filterJobID string, b []byte) error {
	if mock.UpdateFilterOutputBytesFunc == nil {
		panic("ClientMock.UpdateFilterOutputBytesFunc: method is nil but Client.UpdateFilterOutputBytes was just called")
	}
	callInfo := struct {
		Ctx                  context.Context
		UserAuthToken        string
		ServiceAuthToken     string
		DownloadServiceToken string
		FilterJobID          string
		B                    []byte
	}{
		Ctx:                  ctx,
		UserAuthToken:        userAuthToken,
		ServiceAuthToken:     serviceAuthToken,
		DownloadServiceToken: downloadServiceToken,
		FilterJobID:          filterJobID,
		B:                    b,
	}
	lockClientMockUpdateFilterOutputBytes.Lock()
	mock.calls.UpdateFilterOutputBytes = append(mock.calls.UpdateFilterOutputBytes, callInfo)
	lockClientMockUpdateFilterOutputBytes.Unlock()
	return mock.UpdateFilterOutputBytesFunc(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, filterJobID, b)
}

// UpdateFilterOutputBytesCalls gets all the calls that were made to UpdateFilterOutputBytes.
// Check the length with:
//     len(mockedClient.UpdateFilterOutputBytesCalls())
func (mock *ClientMock) UpdateFilterOutputBytesCalls() []struct {
	Ctx                  context.Context
	UserAuthToken        string
	ServiceAuthToken     string
	DownloadServiceToken string
	FilterJobID          string
	B                    []byte
} {
	var calls []struct {
		Ctx                  context.Context
		UserAuthToken        string
		ServiceAuthToken     string
		DownloadServiceToken string
		FilterJobID          string
		B                    []byte
	}
	lockClientMockUpdateFilterOutputBytes.RLock()
	calls = mock.calls.UpdateFilterOutputBytes
	lockClientMockUpdateFilterOutputBytes.RUnlock()
	return calls
}