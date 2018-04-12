// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package filetest

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"sync"
)

var (
	lockUploaderMockUpload sync.RWMutex
)

// UploaderMock is a mock implementation of Uploader.
//
//     func TestSomethingThatUsesUploader(t *testing.T) {
//
//         // make and configure a mocked Uploader
//         mockedUploader := &UploaderMock{
//             UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
// 	               panic("TODO: mock out the Upload method")
//             },
//         }
//
//         // TODO: use mockedUploader in code that requires Uploader
//         //       and then make assertions.
//
//     }
type UploaderMock struct {
	// UploadFunc mocks the Upload method.
	UploadFunc func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)

	// calls tracks calls to the methods.
	calls struct {
		// Upload holds details about calls to the Upload method.
		Upload []struct {
			// Input is the input argument value.
			Input *s3manager.UploadInput
			// Options is the options argument value.
			Options []func(*s3manager.Uploader)
		}
	}
}

// Upload calls UploadFunc.
func (mock *UploaderMock) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	if mock.UploadFunc == nil {
		panic("moq: UploaderMock.UploadFunc is nil but Uploader.Upload was just called")
	}
	callInfo := struct {
		Input   *s3manager.UploadInput
		Options []func(*s3manager.Uploader)
	}{
		Input:   input,
		Options: options,
	}
	lockUploaderMockUpload.Lock()
	mock.calls.Upload = append(mock.calls.Upload, callInfo)
	lockUploaderMockUpload.Unlock()
	return mock.UploadFunc(input, options...)
}

// UploadCalls gets all the calls that were made to Upload.
// Check the length with:
//     len(mockedUploader.UploadCalls())
func (mock *UploaderMock) UploadCalls() []struct {
	Input   *s3manager.UploadInput
	Options []func(*s3manager.Uploader)
} {
	var calls []struct {
		Input   *s3manager.UploadInput
		Options []func(*s3manager.Uploader)
	}
	lockUploaderMockUpload.RLock()
	calls = mock.calls.Upload
	lockUploaderMockUpload.RUnlock()
	return calls
}
