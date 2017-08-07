// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package eventtest

import (
	"sync"
	"github.com/ONSdigital/dp-dataset-exporter/observation"
)

var (
	lockObservationStoreMockGetCSVRows	sync.RWMutex
)

// ObservationStoreMock is a mock implementation of ObservationStore.
//
//     func TestSomethingThatUsesObservationStore(t *testing.T) {
//
//         // make and configure a mocked ObservationStore
//         mockedObservationStore := &ObservationStoreMock{ 
//             GetCSVRowsFunc: func(filter *observation.Filter) (observation.CSVRowReader, error) {
// 	               panic("TODO: mock out the GetCSVRows method")
//             },
//         }
//
//         // TODO: use mockedObservationStore in code that requires ObservationStore
//         //       and then make assertions.
// 
//     }
type ObservationStoreMock struct {
	// GetCSVRowsFunc mocks the GetCSVRows method.
	GetCSVRowsFunc func(filter *observation.Filter) (observation.CSVRowReader, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetCSVRows holds details about calls to the GetCSVRows method.
		GetCSVRows []struct {
			// Filter is the filter argument value.
			Filter *observation.Filter
		}
	}
}

// GetCSVRows calls GetCSVRowsFunc.
func (mock *ObservationStoreMock) GetCSVRows(filter *observation.Filter) (observation.CSVRowReader, error) {
	if mock.GetCSVRowsFunc == nil {
		panic("moq: ObservationStoreMock.GetCSVRowsFunc is nil but ObservationStore.GetCSVRows was just called")
	}
	callInfo := struct {
		Filter *observation.Filter
	}{
		Filter: filter,
	}
	lockObservationStoreMockGetCSVRows.Lock()
	mock.calls.GetCSVRows = append(mock.calls.GetCSVRows, callInfo)
	lockObservationStoreMockGetCSVRows.Unlock()
	return mock.GetCSVRowsFunc(filter)
}

// GetCSVRowsCalls gets all the calls that were made to GetCSVRows.
// Check the length with:
//     len(mockedObservationStore.GetCSVRowsCalls())
func (mock *ObservationStoreMock) GetCSVRowsCalls() []struct {
		Filter *observation.Filter
	} {
	var calls []struct {
		Filter *observation.Filter
	}
	lockObservationStoreMockGetCSVRows.RLock()
	calls = mock.calls.GetCSVRows
	lockObservationStoreMockGetCSVRows.RUnlock()
	return calls
}
