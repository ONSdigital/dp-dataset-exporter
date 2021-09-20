// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package errorstest

import (
	"context"
	"github.com/ONSdigital/dp-dataset-exporter/errors"
	"sync"
)

// Ensure, that HandlerMock does implement errors.Handler.
// If this is not the case, regenerate this file with moq.
var _ errors.Handler = &HandlerMock{}

// HandlerMock is a mock implementation of errors.Handler.
//
// 	func TestSomethingThatUsesHandler(t *testing.T) {
//
// 		// make and configure a mocked errors.Handler
// 		mockedHandler := &HandlerMock{
// 			HandleFunc: func(ctx context.Context, filterID string, err error)  {
// 				panic("mock out the Handle method")
// 			},
// 		}
//
// 		// use mockedHandler in code that requires errors.Handler
// 		// and then make assertions.
//
// 	}
type HandlerMock struct {
	// HandleFunc mocks the Handle method.
	HandleFunc func(ctx context.Context, filterID string, err error)

	// calls tracks calls to the methods.
	calls struct {
		// Handle holds details about calls to the Handle method.
		Handle []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterID is the filterID argument value.
			FilterID string
			// Err is the err argument value.
			Err error
		}
	}
	lockHandle sync.RWMutex
}

// Handle calls HandleFunc.
func (mock *HandlerMock) Handle(ctx context.Context, filterID string, err error) {
	if mock.HandleFunc == nil {
		panic("HandlerMock.HandleFunc: method is nil but Handler.Handle was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		FilterID string
		Err      error
	}{
		Ctx:      ctx,
		FilterID: filterID,
		Err:      err,
	}
	mock.lockHandle.Lock()
	mock.calls.Handle = append(mock.calls.Handle, callInfo)
	mock.lockHandle.Unlock()
	mock.HandleFunc(ctx, filterID, err)
}

// HandleCalls gets all the calls that were made to Handle.
// Check the length with:
//     len(mockedHandler.HandleCalls())
func (mock *HandlerMock) HandleCalls() []struct {
	Ctx      context.Context
	FilterID string
	Err      error
} {
	var calls []struct {
		Ctx      context.Context
		FilterID string
		Err      error
	}
	mock.lockHandle.RLock()
	calls = mock.calls.Handle
	mock.lockHandle.RUnlock()
	return calls
}
