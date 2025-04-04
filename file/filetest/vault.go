// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package filetest

import (
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"sync"
)

// Ensure, that VaultClientMock does implement file.VaultClient.
// If this is not the case, regenerate this file with moq.
var _ file.VaultClient = &VaultClientMock{}

// VaultClientMock is a mock implementation of file.VaultClient.
//
//	func TestSomethingThatUsesVaultClient(t *testing.T) {
//
//		// make and configure a mocked file.VaultClient
//		mockedVaultClient := &VaultClientMock{
//			WriteKeyFunc: func(path string, key string, value string) error {
//				panic("mock out the WriteKey method")
//			},
//		}
//
//		// use mockedVaultClient in code that requires file.VaultClient
//		// and then make assertions.
//
//	}
type VaultClientMock struct {
	// WriteKeyFunc mocks the WriteKey method.
	WriteKeyFunc func(path string, key string, value string) error

	// calls tracks calls to the methods.
	calls struct {
		// WriteKey holds details about calls to the WriteKey method.
		WriteKey []struct {
			// Path is the path argument value.
			Path string
			// Key is the key argument value.
			Key string
			// Value is the value argument value.
			Value string
		}
	}
	lockWriteKey sync.RWMutex
}

// WriteKey calls WriteKeyFunc.
func (mock *VaultClientMock) WriteKey(path string, key string, value string) error {
	if mock.WriteKeyFunc == nil {
		panic("VaultClientMock.WriteKeyFunc: method is nil but VaultClient.WriteKey was just called")
	}
	callInfo := struct {
		Path  string
		Key   string
		Value string
	}{
		Path:  path,
		Key:   key,
		Value: value,
	}
	mock.lockWriteKey.Lock()
	mock.calls.WriteKey = append(mock.calls.WriteKey, callInfo)
	mock.lockWriteKey.Unlock()
	return mock.WriteKeyFunc(path, key, value)
}

// WriteKeyCalls gets all the calls that were made to WriteKey.
// Check the length with:
//
//	len(mockedVaultClient.WriteKeyCalls())
func (mock *VaultClientMock) WriteKeyCalls() []struct {
	Path  string
	Key   string
	Value string
} {
	var calls []struct {
		Path  string
		Key   string
		Value string
	}
	mock.lockWriteKey.RLock()
	calls = mock.calls.WriteKey
	mock.lockWriteKey.RUnlock()
	return calls
}
