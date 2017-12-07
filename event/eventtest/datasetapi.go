// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package eventtest

import (
	"github.com/ONSdigital/go-ns/clients/dataset"
	"sync"
)

var (
	lockDatasetAPIMockGetVersion          sync.RWMutex
	lockDatasetAPIMockPutVersionDownloads sync.RWMutex
)

// DatasetAPIMock is a mock implementation of DatasetAPI.
//
//     func TestSomethingThatUsesDatasetAPI(t *testing.T) {
//
//         // make and configure a mocked DatasetAPI
//         mockedDatasetAPI := &DatasetAPIMock{
//             GetVersionFunc: func(id string, edition string, version string) (dataset.Version, error) {
// 	               panic("TODO: mock out the GetVersion method")
//             },
//             PutVersionDownloadsFunc: func(datasetID string, edition string, version string, downloadList map[string]dataset.Download) error {
// 	               panic("TODO: mock out the PutVersionDownloads method")
//             },
//         }
//
//         // TODO: use mockedDatasetAPI in code that requires DatasetAPI
//         //       and then make assertions.
//
//     }
type DatasetAPIMock struct {
	// GetVersionFunc mocks the GetVersion method.
	GetVersionFunc func(id string, edition string, version string) (dataset.Version, error)

	// PutVersionDownloadsFunc mocks the PutVersionDownloads method.
	PutVersionDownloadsFunc func(datasetID string, edition string, version string, downloadList map[string]dataset.Download) error

	// calls tracks calls to the methods.
	calls struct {
		// GetVersion holds details about calls to the GetVersion method.
		GetVersion []struct {
			// Id is the id argument value.
			Id string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
		}
		// PutVersionDownloads holds details about calls to the PutVersionDownloads method.
		PutVersionDownloads []struct {
			// DatasetID is the datasetID argument value.
			DatasetID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
			// DownloadList is the downloadList argument value.
			DownloadList map[string]dataset.Download
		}
	}
}

// GetVersion calls GetVersionFunc.
func (mock *DatasetAPIMock) GetVersion(id string, edition string, version string) (dataset.Version, error) {
	if mock.GetVersionFunc == nil {
		panic("moq: DatasetAPIMock.GetVersionFunc is nil but DatasetAPI.GetVersion was just called")
	}
	callInfo := struct {
		Id      string
		Edition string
		Version string
	}{
		Id:      id,
		Edition: edition,
		Version: version,
	}
	lockDatasetAPIMockGetVersion.Lock()
	mock.calls.GetVersion = append(mock.calls.GetVersion, callInfo)
	lockDatasetAPIMockGetVersion.Unlock()
	return mock.GetVersionFunc(id, edition, version)
}

// GetVersionCalls gets all the calls that were made to GetVersion.
// Check the length with:
//     len(mockedDatasetAPI.GetVersionCalls())
func (mock *DatasetAPIMock) GetVersionCalls() []struct {
	Id      string
	Edition string
	Version string
} {
	var calls []struct {
		Id      string
		Edition string
		Version string
	}
	lockDatasetAPIMockGetVersion.RLock()
	calls = mock.calls.GetVersion
	lockDatasetAPIMockGetVersion.RUnlock()
	return calls
}

// PutVersionDownloads calls PutVersionDownloadsFunc.
func (mock *DatasetAPIMock) PutVersionDownloads(datasetID string, edition string, version string, downloadList map[string]dataset.Download) error {
	if mock.PutVersionDownloadsFunc == nil {
		panic("moq: DatasetAPIMock.PutVersionDownloadsFunc is nil but DatasetAPI.PutVersionDownloads was just called")
	}
	callInfo := struct {
		DatasetID    string
		Edition      string
		Version      string
		DownloadList map[string]dataset.Download
	}{
		DatasetID:    datasetID,
		Edition:      edition,
		Version:      version,
		DownloadList: downloadList,
	}
	lockDatasetAPIMockPutVersionDownloads.Lock()
	mock.calls.PutVersionDownloads = append(mock.calls.PutVersionDownloads, callInfo)
	lockDatasetAPIMockPutVersionDownloads.Unlock()
	return mock.PutVersionDownloadsFunc(datasetID, edition, version, downloadList)
}

// PutVersionDownloadsCalls gets all the calls that were made to PutVersionDownloads.
// Check the length with:
//     len(mockedDatasetAPI.PutVersionDownloadsCalls())
func (mock *DatasetAPIMock) PutVersionDownloadsCalls() []struct {
	DatasetID    string
	Edition      string
	Version      string
	DownloadList map[string]dataset.Download
} {
	var calls []struct {
		DatasetID    string
		Edition      string
		Version      string
		DownloadList map[string]dataset.Download
	}
	lockDatasetAPIMockPutVersionDownloads.RLock()
	calls = mock.calls.PutVersionDownloads
	lockDatasetAPIMockPutVersionDownloads.RUnlock()
	return calls
}