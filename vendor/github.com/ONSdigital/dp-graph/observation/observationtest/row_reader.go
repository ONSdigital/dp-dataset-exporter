// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package observationtest

import (
	"context"
	"github.com/ONSdigital/dp-graph/observation"
	"sync"
)

var (
	lockStreamRowReaderMockClose sync.RWMutex
	lockStreamRowReaderMockRead  sync.RWMutex
)

// Ensure, that StreamRowReaderMock does implement StreamRowReader.
// If this is not the case, regenerate this file with moq.
var _ observation.StreamRowReader = &StreamRowReaderMock{}

// StreamRowReaderMock is a mock implementation of StreamRowReader.
//
//     func TestSomethingThatUsesStreamRowReader(t *testing.T) {
//
//         // make and configure a mocked StreamRowReader
//         mockedStreamRowReader := &StreamRowReaderMock{
//             CloseFunc: func(in1 context.Context) error {
// 	               panic("mock out the Close method")
//             },
//             ReadFunc: func() (string, error) {
// 	               panic("mock out the Read method")
//             },
//         }
//
//         // use mockedStreamRowReader in code that requires StreamRowReader
//         // and then make assertions.
//
//     }
type StreamRowReaderMock struct {
	// CloseFunc mocks the Close method.
	CloseFunc func(in1 context.Context) error

	// ReadFunc mocks the Read method.
	ReadFunc func() (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Close holds details about calls to the Close method.
		Close []struct {
			// In1 is the in1 argument value.
			In1 context.Context
		}
		// Read holds details about calls to the Read method.
		Read []struct {
		}
	}
}

// Close calls CloseFunc.
func (mock *StreamRowReaderMock) Close(in1 context.Context) error {
	if mock.CloseFunc == nil {
		panic("StreamRowReaderMock.CloseFunc: method is nil but StreamRowReader.Close was just called")
	}
	callInfo := struct {
		In1 context.Context
	}{
		In1: in1,
	}
	lockStreamRowReaderMockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	lockStreamRowReaderMockClose.Unlock()
	return mock.CloseFunc(in1)
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//     len(mockedStreamRowReader.CloseCalls())
func (mock *StreamRowReaderMock) CloseCalls() []struct {
	In1 context.Context
} {
	var calls []struct {
		In1 context.Context
	}
	lockStreamRowReaderMockClose.RLock()
	calls = mock.calls.Close
	lockStreamRowReaderMockClose.RUnlock()
	return calls
}

// Read calls ReadFunc.
func (mock *StreamRowReaderMock) Read() (string, error) {
	if mock.ReadFunc == nil {
		panic("StreamRowReaderMock.ReadFunc: method is nil but StreamRowReader.Read was just called")
	}
	callInfo := struct {
	}{}
	lockStreamRowReaderMockRead.Lock()
	mock.calls.Read = append(mock.calls.Read, callInfo)
	lockStreamRowReaderMockRead.Unlock()
	return mock.ReadFunc()
}

// ReadCalls gets all the calls that were made to Read.
// Check the length with:
//     len(mockedStreamRowReader.ReadCalls())
func (mock *StreamRowReaderMock) ReadCalls() []struct {
} {
	var calls []struct {
	}
	lockStreamRowReaderMockRead.RLock()
	calls = mock.calls.Read
	lockStreamRowReaderMockRead.RUnlock()
	return calls
}