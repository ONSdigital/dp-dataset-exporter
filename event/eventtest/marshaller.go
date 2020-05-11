// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package eventtest

import (
	"github.com/ONSdigital/dp-dataset-exporter/event"
	"sync"
)

var (
	lockMarshallerMockMarshal sync.RWMutex
)

// Ensure, that MarshallerMock does implement Marshaller.
// If this is not the case, regenerate this file with moq.
var _ event.Marshaller = &MarshallerMock{}

// MarshallerMock is a mock implementation of event.Marshaller.
//
//     func TestSomethingThatUsesMarshaller(t *testing.T) {
//
//         // make and configure a mocked event.Marshaller
//         mockedMarshaller := &MarshallerMock{
//             MarshalFunc: func(s interface{}) ([]byte, error) {
// 	               panic("mock out the Marshal method")
//             },
//         }
//
//         // use mockedMarshaller in code that requires event.Marshaller
//         // and then make assertions.
//
//     }
type MarshallerMock struct {
	// MarshalFunc mocks the Marshal method.
	MarshalFunc func(s interface{}) ([]byte, error)

	// calls tracks calls to the methods.
	calls struct {
		// Marshal holds details about calls to the Marshal method.
		Marshal []struct {
			// S is the s argument value.
			S interface{}
		}
	}
}

// Marshal calls MarshalFunc.
func (mock *MarshallerMock) Marshal(s interface{}) ([]byte, error) {
	if mock.MarshalFunc == nil {
		panic("MarshallerMock.MarshalFunc: method is nil but Marshaller.Marshal was just called")
	}
	callInfo := struct {
		S interface{}
	}{
		S: s,
	}
	lockMarshallerMockMarshal.Lock()
	mock.calls.Marshal = append(mock.calls.Marshal, callInfo)
	lockMarshallerMockMarshal.Unlock()
	return mock.MarshalFunc(s)
}

// MarshalCalls gets all the calls that were made to Marshal.
// Check the length with:
//     len(mockedMarshaller.MarshalCalls())
func (mock *MarshallerMock) MarshalCalls() []struct {
	S interface{}
} {
	var calls []struct {
		S interface{}
	}
	lockMarshallerMockMarshal.RLock()
	calls = mock.calls.Marshal
	lockMarshallerMockMarshal.RUnlock()
	return calls
}
