// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package eventtest

import (
	"io"
	"sync"
)

var (
	lockFileStoreMockPutFile sync.RWMutex
)

// FileStoreMock is a mock implementation of FileStore.
//
//     func TestSomethingThatUsesFileStore(t *testing.T) {
//
//         // make and configure a mocked FileStore
//         mockedFileStore := &FileStoreMock{
//             PutFileFunc: func(reader io.Reader, filename string, isPublished bool) (string, error) {
// 	               panic("TODO: mock out the PutFile method")
//             },
//         }
//
//         // TODO: use mockedFileStore in code that requires FileStore
//         //       and then make assertions.
//
//     }
type FileStoreMock struct {
	// PutFileFunc mocks the PutFile method.
	PutFileFunc func(reader io.Reader, filename string, isPublished bool) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// PutFile holds details about calls to the PutFile method.
		PutFile []struct {
			// Reader is the reader argument value.
			Reader io.Reader
			// Filename is the filename argument value.
			Filename string
			// IsPublished is the isPublished argument value.
			IsPublished bool
		}
	}
}

// PutFile calls PutFileFunc.
func (mock *FileStoreMock) PutFile(reader io.Reader, filename string, isPublished bool) (string, error) {
	if mock.PutFileFunc == nil {
		panic("FileStoreMock.PutFileFunc: method is nil but FileStore.PutFile was just called")
	}
	callInfo := struct {
		Reader      io.Reader
		Filename    string
		IsPublished bool
	}{
		Reader:      reader,
		Filename:    filename,
		IsPublished: isPublished,
	}
	lockFileStoreMockPutFile.Lock()
	mock.calls.PutFile = append(mock.calls.PutFile, callInfo)
	lockFileStoreMockPutFile.Unlock()
	return mock.PutFileFunc(reader, filename, isPublished)
}

// PutFileCalls gets all the calls that were made to PutFile.
// Check the length with:
//     len(mockedFileStore.PutFileCalls())
func (mock *FileStoreMock) PutFileCalls() []struct {
	Reader      io.Reader
	Filename    string
	IsPublished bool
} {
	var calls []struct {
		Reader      io.Reader
		Filename    string
		IsPublished bool
	}
	lockFileStoreMockPutFile.RLock()
	calls = mock.calls.PutFile
	lockFileStoreMockPutFile.RUnlock()
	return calls
}
