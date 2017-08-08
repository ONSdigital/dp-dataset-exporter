// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package eventtest

import (
	"sync"
	"github.com/ONSdigital/dp-dataset-exporter/observation"
)

var (
	lockFilterStoreMockGetFilter	sync.RWMutex
	lockFilterStoreMockPutCSVData	sync.RWMutex
)

// FilterStoreMock is a mock implementation of FilterStore.
//
//     func TestSomethingThatUsesFilterStore(t *testing.T) {
//
//         // make and configure a mocked FilterStore
//         mockedFilterStore := &FilterStoreMock{ 
//             GetFilterFunc: func(filterJobID string) (*observation.Filter, error) {
// 	               panic("TODO: mock out the GetFilter method")
//             },
//             PutCSVDataFunc: func(filterJobID string, url string, size int64) error {
// 	               panic("TODO: mock out the PutCSVData method")
//             },
//         }
//
//         // TODO: use mockedFilterStore in code that requires FilterStore
//         //       and then make assertions.
// 
//     }
type FilterStoreMock struct {
	// GetFilterFunc mocks the GetFilter method.
	GetFilterFunc func(filterJobID string) (*observation.Filter, error)

	// PutCSVDataFunc mocks the PutCSVData method.
	PutCSVDataFunc func(filterJobID string, url string, size int64) error

	// calls tracks calls to the methods.
	calls struct {
		// GetFilter holds details about calls to the GetFilter method.
		GetFilter []struct {
			// FilterJobID is the filterJobID argument value.
			FilterJobID string
		}
		// PutCSVData holds details about calls to the PutCSVData method.
		PutCSVData []struct {
			// FilterJobID is the filterJobID argument value.
			FilterJobID string
			// Url is the url argument value.
			Url string
			// Size is the size argument value.
			Size int64
		}
	}
}

// GetFilter calls GetFilterFunc.
func (mock *FilterStoreMock) GetFilter(filterJobID string) (*observation.Filter, error) {
	if mock.GetFilterFunc == nil {
		panic("moq: FilterStoreMock.GetFilterFunc is nil but FilterStore.GetFilter was just called")
	}
	callInfo := struct {
		FilterJobID string
	}{
		FilterJobID: filterJobID,
	}
	lockFilterStoreMockGetFilter.Lock()
	mock.calls.GetFilter = append(mock.calls.GetFilter, callInfo)
	lockFilterStoreMockGetFilter.Unlock()
	return mock.GetFilterFunc(filterJobID)
}

// GetFilterCalls gets all the calls that were made to GetFilter.
// Check the length with:
//     len(mockedFilterStore.GetFilterCalls())
func (mock *FilterStoreMock) GetFilterCalls() []struct {
		FilterJobID string
	} {
	var calls []struct {
		FilterJobID string
	}
	lockFilterStoreMockGetFilter.RLock()
	calls = mock.calls.GetFilter
	lockFilterStoreMockGetFilter.RUnlock()
	return calls
}

// PutCSVData calls PutCSVDataFunc.
func (mock *FilterStoreMock) PutCSVData(filterJobID string, url string, size int64) error {
	if mock.PutCSVDataFunc == nil {
		panic("moq: FilterStoreMock.PutCSVDataFunc is nil but FilterStore.PutCSVData was just called")
	}
	callInfo := struct {
		FilterJobID string
		Url string
		Size int64
	}{
		FilterJobID: filterJobID,
		Url: url,
		Size: size,
	}
	lockFilterStoreMockPutCSVData.Lock()
	mock.calls.PutCSVData = append(mock.calls.PutCSVData, callInfo)
	lockFilterStoreMockPutCSVData.Unlock()
	return mock.PutCSVDataFunc(filterJobID, url, size)
}

// PutCSVDataCalls gets all the calls that were made to PutCSVData.
// Check the length with:
//     len(mockedFilterStore.PutCSVDataCalls())
func (mock *FilterStoreMock) PutCSVDataCalls() []struct {
		FilterJobID string
		Url string
		Size int64
	} {
	var calls []struct {
		FilterJobID string
		Url string
		Size int64
	}
	lockFilterStoreMockPutCSVData.RLock()
	calls = mock.calls.PutCSVData
	lockFilterStoreMockPutCSVData.RUnlock()
	return calls
}
