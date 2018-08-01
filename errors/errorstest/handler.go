// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package errorstest

import (
	"sync"
)

var (
	lockHandlerMockHandle sync.RWMutex
)

// HandlerMock is a mock implementation of Handler.
//
//     func TestSomethingThatUsesHandler(t *testing.T) {
//
//         // make and configure a mocked Handler
//         mockedHandler := &HandlerMock{
//             HandleFunc: func(filterID string, err error)  {
// 	               panic("TODO: mock out the Handle method")
//             },
//         }
//
//         // TODO: use mockedHandler in code that requires Handler
//         //       and then make assertions.
//
//     }
type HandlerMock struct {
	// HandleFunc mocks the Handle method.
	HandleFunc func(filterID string, err error)

	// calls tracks calls to the methods.
	calls struct {
		// Handle holds details about calls to the Handle method.
		Handle []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Err is the err argument value.
			Err error
		}
	}
}

// Handle calls HandleFunc.
func (mock *HandlerMock) Handle(filterID string, err error) {
	if mock.HandleFunc == nil {
		panic("HandlerMock.HandleFunc: method is nil but Handler.Handle was just called")
	}
	callInfo := struct {
		FilterID string
		Err      error
	}{
		FilterID: filterID,
		Err:      err,
	}
	lockHandlerMockHandle.Lock()
	mock.calls.Handle = append(mock.calls.Handle, callInfo)
	lockHandlerMockHandle.Unlock()
	mock.HandleFunc(filterID, err)
}

// HandleCalls gets all the calls that were made to Handle.
// Check the length with:
//     len(mockedHandler.HandleCalls())
func (mock *HandlerMock) HandleCalls() []struct {
	FilterID string
	Err      error
} {
	var calls []struct {
		FilterID string
		Err      error
	}
	lockHandlerMockHandle.RLock()
	calls = mock.calls.Handle
	lockHandlerMockHandle.RUnlock()
	return calls
}
