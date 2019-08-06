// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package internal

import (
	"context"
	"github.com/ONSdigital/dp-graph/neptune/driver"
	"github.com/ONSdigital/graphson"
	"github.com/ONSdigital/gremgo-neptune"
	"sync"
)

var (
	lockNeptunePoolMockClose            sync.RWMutex
	lockNeptunePoolMockExecute          sync.RWMutex
	lockNeptunePoolMockGet              sync.RWMutex
	lockNeptunePoolMockGetCount         sync.RWMutex
	lockNeptunePoolMockGetE             sync.RWMutex
	lockNeptunePoolMockGetStringList    sync.RWMutex
	lockNeptunePoolMockOpenStreamCursor sync.RWMutex
)

// Ensure, that NeptunePoolMock does implement NeptunePool.
// If this is not the case, regenerate this file with moq.
var _ driver.NeptunePool = &NeptunePoolMock{}

// NeptunePoolMock is a mock implementation of NeptunePool.
//
//     func TestSomethingThatUsesNeptunePool(t *testing.T) {
//
//         // make and configure a mocked NeptunePool
//         mockedNeptunePool := &NeptunePoolMock{
//             CloseFunc: func()  {
// 	               panic("mock out the Close method")
//             },
//             ExecuteFunc: func(query string, bindings map[string]string, rebindings map[string]string) ([]gremgo.Response, error) {
// 	               panic("mock out the Execute method")
//             },
//             GetFunc: func(query string, bindings map[string]string, rebindings map[string]string) ([]graphson.Vertex, error) {
// 	               panic("mock out the Get method")
//             },
//             GetCountFunc: func(q string, bindings map[string]string, rebindings map[string]string) (int64, error) {
// 	               panic("mock out the GetCount method")
//             },
//             GetEFunc: func(q string, bindings map[string]string, rebindings map[string]string) (interface{}, error) {
// 	               panic("mock out the GetE method")
//             },
//             GetStringListFunc: func(query string, bindings map[string]string, rebindings map[string]string) ([]string, error) {
// 	               panic("mock out the GetStringList method")
//             },
//             OpenStreamCursorFunc: func(ctx context.Context, query string, bindings map[string]string, rebindings map[string]string) (*gremgo.Stream, error) {
// 	               panic("mock out the OpenStreamCursor method")
//             },
//         }
//
//         // use mockedNeptunePool in code that requires NeptunePool
//         // and then make assertions.
//
//     }
type NeptunePoolMock struct {
	// CloseFunc mocks the Close method.
	CloseFunc func()

	// ExecuteFunc mocks the Execute method.
	ExecuteFunc func(query string, bindings map[string]string, rebindings map[string]string) ([]gremgo.Response, error)

	// GetFunc mocks the Get method.
	GetFunc func(query string, bindings map[string]string, rebindings map[string]string) ([]graphson.Vertex, error)

	// GetCountFunc mocks the GetCount method.
	GetCountFunc func(q string, bindings map[string]string, rebindings map[string]string) (int64, error)

	// GetEFunc mocks the GetE method.
	GetEFunc func(q string, bindings map[string]string, rebindings map[string]string) (interface{}, error)

	// GetStringListFunc mocks the GetStringList method.
	GetStringListFunc func(query string, bindings map[string]string, rebindings map[string]string) ([]string, error)

	// OpenStreamCursorFunc mocks the OpenStreamCursor method.
	OpenStreamCursorFunc func(ctx context.Context, query string, bindings map[string]string, rebindings map[string]string) (*gremgo.Stream, error)

	// calls tracks calls to the methods.
	calls struct {
		// Close holds details about calls to the Close method.
		Close []struct {
		}
		// Execute holds details about calls to the Execute method.
		Execute []struct {
			// Query is the query argument value.
			Query string
			// Bindings is the bindings argument value.
			Bindings map[string]string
			// Rebindings is the rebindings argument value.
			Rebindings map[string]string
		}
		// Get holds details about calls to the Get method.
		Get []struct {
			// Query is the query argument value.
			Query string
			// Bindings is the bindings argument value.
			Bindings map[string]string
			// Rebindings is the rebindings argument value.
			Rebindings map[string]string
		}
		// GetCount holds details about calls to the GetCount method.
		GetCount []struct {
			// Q is the q argument value.
			Q string
			// Bindings is the bindings argument value.
			Bindings map[string]string
			// Rebindings is the rebindings argument value.
			Rebindings map[string]string
		}
		// GetE holds details about calls to the GetE method.
		GetE []struct {
			// Q is the q argument value.
			Q string
			// Bindings is the bindings argument value.
			Bindings map[string]string
			// Rebindings is the rebindings argument value.
			Rebindings map[string]string
		}
		// GetStringList holds details about calls to the GetStringList method.
		GetStringList []struct {
			// Query is the query argument value.
			Query string
			// Bindings is the bindings argument value.
			Bindings map[string]string
			// Rebindings is the rebindings argument value.
			Rebindings map[string]string
		}
		// OpenStreamCursor holds details about calls to the OpenStreamCursor method.
		OpenStreamCursor []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Query is the query argument value.
			Query string
			// Bindings is the bindings argument value.
			Bindings map[string]string
			// Rebindings is the rebindings argument value.
			Rebindings map[string]string
		}
	}
}

// Close calls CloseFunc.
func (mock *NeptunePoolMock) Close() {
	if mock.CloseFunc == nil {
		panic("NeptunePoolMock.CloseFunc: method is nil but NeptunePool.Close was just called")
	}
	callInfo := struct {
	}{}
	lockNeptunePoolMockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	lockNeptunePoolMockClose.Unlock()
	mock.CloseFunc()
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//     len(mockedNeptunePool.CloseCalls())
func (mock *NeptunePoolMock) CloseCalls() []struct {
} {
	var calls []struct {
	}
	lockNeptunePoolMockClose.RLock()
	calls = mock.calls.Close
	lockNeptunePoolMockClose.RUnlock()
	return calls
}

// Execute calls ExecuteFunc.
func (mock *NeptunePoolMock) Execute(query string, bindings map[string]string, rebindings map[string]string) ([]gremgo.Response, error) {
	if mock.ExecuteFunc == nil {
		panic("NeptunePoolMock.ExecuteFunc: method is nil but NeptunePool.Execute was just called")
	}
	callInfo := struct {
		Query      string
		Bindings   map[string]string
		Rebindings map[string]string
	}{
		Query:      query,
		Bindings:   bindings,
		Rebindings: rebindings,
	}
	lockNeptunePoolMockExecute.Lock()
	mock.calls.Execute = append(mock.calls.Execute, callInfo)
	lockNeptunePoolMockExecute.Unlock()
	return mock.ExecuteFunc(query, bindings, rebindings)
}

// ExecuteCalls gets all the calls that were made to Execute.
// Check the length with:
//     len(mockedNeptunePool.ExecuteCalls())
func (mock *NeptunePoolMock) ExecuteCalls() []struct {
	Query      string
	Bindings   map[string]string
	Rebindings map[string]string
} {
	var calls []struct {
		Query      string
		Bindings   map[string]string
		Rebindings map[string]string
	}
	lockNeptunePoolMockExecute.RLock()
	calls = mock.calls.Execute
	lockNeptunePoolMockExecute.RUnlock()
	return calls
}

// Get calls GetFunc.
func (mock *NeptunePoolMock) Get(query string, bindings map[string]string, rebindings map[string]string) ([]graphson.Vertex, error) {
	if mock.GetFunc == nil {
		panic("NeptunePoolMock.GetFunc: method is nil but NeptunePool.Get was just called")
	}
	callInfo := struct {
		Query      string
		Bindings   map[string]string
		Rebindings map[string]string
	}{
		Query:      query,
		Bindings:   bindings,
		Rebindings: rebindings,
	}
	lockNeptunePoolMockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	lockNeptunePoolMockGet.Unlock()
	return mock.GetFunc(query, bindings, rebindings)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//     len(mockedNeptunePool.GetCalls())
func (mock *NeptunePoolMock) GetCalls() []struct {
	Query      string
	Bindings   map[string]string
	Rebindings map[string]string
} {
	var calls []struct {
		Query      string
		Bindings   map[string]string
		Rebindings map[string]string
	}
	lockNeptunePoolMockGet.RLock()
	calls = mock.calls.Get
	lockNeptunePoolMockGet.RUnlock()
	return calls
}

// GetCount calls GetCountFunc.
func (mock *NeptunePoolMock) GetCount(q string, bindings map[string]string, rebindings map[string]string) (int64, error) {
	if mock.GetCountFunc == nil {
		panic("NeptunePoolMock.GetCountFunc: method is nil but NeptunePool.GetCount was just called")
	}
	callInfo := struct {
		Q          string
		Bindings   map[string]string
		Rebindings map[string]string
	}{
		Q:          q,
		Bindings:   bindings,
		Rebindings: rebindings,
	}
	lockNeptunePoolMockGetCount.Lock()
	mock.calls.GetCount = append(mock.calls.GetCount, callInfo)
	lockNeptunePoolMockGetCount.Unlock()
	return mock.GetCountFunc(q, bindings, rebindings)
}

// GetCountCalls gets all the calls that were made to GetCount.
// Check the length with:
//     len(mockedNeptunePool.GetCountCalls())
func (mock *NeptunePoolMock) GetCountCalls() []struct {
	Q          string
	Bindings   map[string]string
	Rebindings map[string]string
} {
	var calls []struct {
		Q          string
		Bindings   map[string]string
		Rebindings map[string]string
	}
	lockNeptunePoolMockGetCount.RLock()
	calls = mock.calls.GetCount
	lockNeptunePoolMockGetCount.RUnlock()
	return calls
}

// GetE calls GetEFunc.
func (mock *NeptunePoolMock) GetE(q string, bindings map[string]string, rebindings map[string]string) (interface{}, error) {
	if mock.GetEFunc == nil {
		panic("NeptunePoolMock.GetEFunc: method is nil but NeptunePool.GetE was just called")
	}
	callInfo := struct {
		Q          string
		Bindings   map[string]string
		Rebindings map[string]string
	}{
		Q:          q,
		Bindings:   bindings,
		Rebindings: rebindings,
	}
	lockNeptunePoolMockGetE.Lock()
	mock.calls.GetE = append(mock.calls.GetE, callInfo)
	lockNeptunePoolMockGetE.Unlock()
	return mock.GetEFunc(q, bindings, rebindings)
}

// GetECalls gets all the calls that were made to GetE.
// Check the length with:
//     len(mockedNeptunePool.GetECalls())
func (mock *NeptunePoolMock) GetECalls() []struct {
	Q          string
	Bindings   map[string]string
	Rebindings map[string]string
} {
	var calls []struct {
		Q          string
		Bindings   map[string]string
		Rebindings map[string]string
	}
	lockNeptunePoolMockGetE.RLock()
	calls = mock.calls.GetE
	lockNeptunePoolMockGetE.RUnlock()
	return calls
}

// GetStringList calls GetStringListFunc.
func (mock *NeptunePoolMock) GetStringList(query string, bindings map[string]string, rebindings map[string]string) ([]string, error) {
	if mock.GetStringListFunc == nil {
		panic("NeptunePoolMock.GetStringListFunc: method is nil but NeptunePool.GetStringList was just called")
	}
	callInfo := struct {
		Query      string
		Bindings   map[string]string
		Rebindings map[string]string
	}{
		Query:      query,
		Bindings:   bindings,
		Rebindings: rebindings,
	}
	lockNeptunePoolMockGetStringList.Lock()
	mock.calls.GetStringList = append(mock.calls.GetStringList, callInfo)
	lockNeptunePoolMockGetStringList.Unlock()
	return mock.GetStringListFunc(query, bindings, rebindings)
}

// GetStringListCalls gets all the calls that were made to GetStringList.
// Check the length with:
//     len(mockedNeptunePool.GetStringListCalls())
func (mock *NeptunePoolMock) GetStringListCalls() []struct {
	Query      string
	Bindings   map[string]string
	Rebindings map[string]string
} {
	var calls []struct {
		Query      string
		Bindings   map[string]string
		Rebindings map[string]string
	}
	lockNeptunePoolMockGetStringList.RLock()
	calls = mock.calls.GetStringList
	lockNeptunePoolMockGetStringList.RUnlock()
	return calls
}

// OpenStreamCursor calls OpenStreamCursorFunc.
func (mock *NeptunePoolMock) OpenStreamCursor(ctx context.Context, query string, bindings map[string]string, rebindings map[string]string) (*gremgo.Stream, error) {
	if mock.OpenStreamCursorFunc == nil {
		panic("NeptunePoolMock.OpenStreamCursorFunc: method is nil but NeptunePool.OpenStreamCursor was just called")
	}
	callInfo := struct {
		Ctx        context.Context
		Query      string
		Bindings   map[string]string
		Rebindings map[string]string
	}{
		Ctx:        ctx,
		Query:      query,
		Bindings:   bindings,
		Rebindings: rebindings,
	}
	lockNeptunePoolMockOpenStreamCursor.Lock()
	mock.calls.OpenStreamCursor = append(mock.calls.OpenStreamCursor, callInfo)
	lockNeptunePoolMockOpenStreamCursor.Unlock()
	return mock.OpenStreamCursorFunc(ctx, query, bindings, rebindings)
}

// OpenStreamCursorCalls gets all the calls that were made to OpenStreamCursor.
// Check the length with:
//     len(mockedNeptunePool.OpenStreamCursorCalls())
func (mock *NeptunePoolMock) OpenStreamCursorCalls() []struct {
	Ctx        context.Context
	Query      string
	Bindings   map[string]string
	Rebindings map[string]string
} {
	var calls []struct {
		Ctx        context.Context
		Query      string
		Bindings   map[string]string
		Rebindings map[string]string
	}
	lockNeptunePoolMockOpenStreamCursor.RLock()
	calls = mock.calls.OpenStreamCursor
	lockNeptunePoolMockOpenStreamCursor.RUnlock()
	return calls
}
