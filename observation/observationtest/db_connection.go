// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package observationtest

import (
	"sync"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

var (
	lockDBConnectionMockQueryNeo	sync.RWMutex
)

// DBConnectionMock is a mock implementation of DBConnection.
//
//     func TestSomethingThatUsesDBConnection(t *testing.T) {
//
//         // make and configure a mocked DBConnection
//         mockedDBConnection := &DBConnectionMock{ 
//             QueryNeoFunc: func(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Rows, error) {
// 	               panic("TODO: mock out the QueryNeo method")
//             },
//         }
//
//         // TODO: use mockedDBConnection in code that requires DBConnection
//         //       and then make assertions.
// 
//     }
type DBConnectionMock struct {
	// QueryNeoFunc mocks the QueryNeo method.
	QueryNeoFunc func(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Rows, error)

	// calls tracks calls to the methods.
	calls struct {
		// QueryNeo holds details about calls to the QueryNeo method.
		QueryNeo []struct {
			// Query is the query argument value.
			Query string
			// Params is the params argument value.
			Params map[string]interface{}
		}
	}
}

// QueryNeo calls QueryNeoFunc.
func (mock *DBConnectionMock) QueryNeo(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Rows, error) {
	if mock.QueryNeoFunc == nil {
		panic("moq: DBConnectionMock.QueryNeoFunc is nil but DBConnection.QueryNeo was just called")
	}
	callInfo := struct {
		Query string
		Params map[string]interface{}
	}{
		Query: query,
		Params: params,
	}
	lockDBConnectionMockQueryNeo.Lock()
	mock.calls.QueryNeo = append(mock.calls.QueryNeo, callInfo)
	lockDBConnectionMockQueryNeo.Unlock()
	return mock.QueryNeoFunc(query, params)
}

// QueryNeoCalls gets all the calls that were made to QueryNeo.
// Check the length with:
//     len(mockedDBConnection.QueryNeoCalls())
func (mock *DBConnectionMock) QueryNeoCalls() []struct {
		Query string
		Params map[string]interface{}
	} {
	var calls []struct {
		Query string
		Params map[string]interface{}
	}
	lockDBConnectionMockQueryNeo.RLock()
	calls = mock.calls.QueryNeo
	lockDBConnectionMockQueryNeo.RUnlock()
	return calls
}
