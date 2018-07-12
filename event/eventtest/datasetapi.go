// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package eventtest

import (
	"github.com/ONSdigital/go-ns/clients/dataset"
	"sync"
)

var (
	lockDatasetAPIMockGetInstance        sync.RWMutex
	lockDatasetAPIMockGetMetadataURL     sync.RWMutex
	lockDatasetAPIMockGetVersion         sync.RWMutex
	lockDatasetAPIMockGetVersionMetadata sync.RWMutex
	lockDatasetAPIMockPutVersion         sync.RWMutex
)

// DatasetAPIMock is a mock implementation of DatasetAPI.
//
//     func TestSomethingThatUsesDatasetAPI(t *testing.T) {
//
//         // make and configure a mocked DatasetAPI
//         mockedDatasetAPI := &DatasetAPIMock{
//             GetInstanceFunc: func(instanceID string, cfg ...dataset.Config) (dataset.Instance, error) {
// 	               panic("TODO: mock out the GetInstance method")
//             },
//             GetMetadataURLFunc: func(id string, edition string, version string) string {
// 	               panic("TODO: mock out the GetMetadataURL method")
//             },
//             GetVersionFunc: func(id string, edition string, version string, cfg ...dataset.Config) (dataset.Version, error) {
// 	               panic("TODO: mock out the GetVersion method")
//             },
//             GetVersionMetadataFunc: func(id string, edition string, version string, cfg ...dataset.Config) (dataset.Metadata, error) {
// 	               panic("TODO: mock out the GetVersionMetadata method")
//             },
//             PutVersionFunc: func(id string, edition string, version string, m dataset.Version, cfg ...dataset.Config) error {
// 	               panic("TODO: mock out the PutVersion method")
//             },
//         }
//
//         // TODO: use mockedDatasetAPI in code that requires DatasetAPI
//         //       and then make assertions.
//
//     }
type DatasetAPIMock struct {
	// GetInstanceFunc mocks the GetInstance method.
	GetInstanceFunc func(instanceID string, cfg ...dataset.Config) (dataset.Instance, error)

	// GetMetadataURLFunc mocks the GetMetadataURL method.
	GetMetadataURLFunc func(id string, edition string, version string) string

	// GetVersionFunc mocks the GetVersion method.
	GetVersionFunc func(id string, edition string, version string, cfg ...dataset.Config) (dataset.Version, error)

	// GetVersionMetadataFunc mocks the GetVersionMetadata method.
	GetVersionMetadataFunc func(id string, edition string, version string, cfg ...dataset.Config) (dataset.Metadata, error)

	// PutVersionFunc mocks the PutVersion method.
	PutVersionFunc func(id string, edition string, version string, m dataset.Version, cfg ...dataset.Config) error

	// calls tracks calls to the methods.
	calls struct {
		// GetInstance holds details about calls to the GetInstance method.
		GetInstance []struct {
			// InstanceID is the instanceID argument value.
			InstanceID string
			// Cfg is the cfg argument value.
			Cfg []dataset.Config
		}
		// GetMetadataURL holds details about calls to the GetMetadataURL method.
		GetMetadataURL []struct {
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
		}
		// GetVersion holds details about calls to the GetVersion method.
		GetVersion []struct {
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
			// Cfg is the cfg argument value.
			Cfg []dataset.Config
		}
		// GetVersionMetadata holds details about calls to the GetVersionMetadata method.
		GetVersionMetadata []struct {
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
			// Cfg is the cfg argument value.
			Cfg []dataset.Config
		}
		// PutVersion holds details about calls to the PutVersion method.
		PutVersion []struct {
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
			// M is the m argument value.
			M dataset.Version
			// Cfg is the cfg argument value.
			Cfg []dataset.Config
		}
	}
}

// GetInstance calls GetInstanceFunc.
func (mock *DatasetAPIMock) GetInstance(instanceID string, cfg ...dataset.Config) (dataset.Instance, error) {
	if mock.GetInstanceFunc == nil {
		panic("DatasetAPIMock.GetInstanceFunc: method is nil but DatasetAPI.GetInstance was just called")
	}
	callInfo := struct {
		InstanceID string
		Cfg        []dataset.Config
	}{
		InstanceID: instanceID,
		Cfg:        cfg,
	}
	lockDatasetAPIMockGetInstance.Lock()
	mock.calls.GetInstance = append(mock.calls.GetInstance, callInfo)
	lockDatasetAPIMockGetInstance.Unlock()
	return mock.GetInstanceFunc(instanceID, cfg...)
}

// GetInstanceCalls gets all the calls that were made to GetInstance.
// Check the length with:
//     len(mockedDatasetAPI.GetInstanceCalls())
func (mock *DatasetAPIMock) GetInstanceCalls() []struct {
	InstanceID string
	Cfg        []dataset.Config
} {
	var calls []struct {
		InstanceID string
		Cfg        []dataset.Config
	}
	lockDatasetAPIMockGetInstance.RLock()
	calls = mock.calls.GetInstance
	lockDatasetAPIMockGetInstance.RUnlock()
	return calls
}

// GetMetadataURL calls GetMetadataURLFunc.
func (mock *DatasetAPIMock) GetMetadataURL(id string, edition string, version string) string {
	if mock.GetMetadataURLFunc == nil {
		panic("DatasetAPIMock.GetMetadataURLFunc: method is nil but DatasetAPI.GetMetadataURL was just called")
	}
	callInfo := struct {
		ID      string
		Edition string
		Version string
	}{
		ID:      id,
		Edition: edition,
		Version: version,
	}
	lockDatasetAPIMockGetMetadataURL.Lock()
	mock.calls.GetMetadataURL = append(mock.calls.GetMetadataURL, callInfo)
	lockDatasetAPIMockGetMetadataURL.Unlock()
	return mock.GetMetadataURLFunc(id, edition, version)
}

// GetMetadataURLCalls gets all the calls that were made to GetMetadataURL.
// Check the length with:
//     len(mockedDatasetAPI.GetMetadataURLCalls())
func (mock *DatasetAPIMock) GetMetadataURLCalls() []struct {
	ID      string
	Edition string
	Version string
} {
	var calls []struct {
		ID      string
		Edition string
		Version string
	}
	lockDatasetAPIMockGetMetadataURL.RLock()
	calls = mock.calls.GetMetadataURL
	lockDatasetAPIMockGetMetadataURL.RUnlock()
	return calls
}

// GetVersion calls GetVersionFunc.
func (mock *DatasetAPIMock) GetVersion(id string, edition string, version string, cfg ...dataset.Config) (dataset.Version, error) {
	if mock.GetVersionFunc == nil {
		panic("DatasetAPIMock.GetVersionFunc: method is nil but DatasetAPI.GetVersion was just called")
	}
	callInfo := struct {
		ID      string
		Edition string
		Version string
		Cfg     []dataset.Config
	}{
		ID:      id,
		Edition: edition,
		Version: version,
		Cfg:     cfg,
	}
	lockDatasetAPIMockGetVersion.Lock()
	mock.calls.GetVersion = append(mock.calls.GetVersion, callInfo)
	lockDatasetAPIMockGetVersion.Unlock()
	return mock.GetVersionFunc(id, edition, version, cfg...)
}

// GetVersionCalls gets all the calls that were made to GetVersion.
// Check the length with:
//     len(mockedDatasetAPI.GetVersionCalls())
func (mock *DatasetAPIMock) GetVersionCalls() []struct {
	ID      string
	Edition string
	Version string
	Cfg     []dataset.Config
} {
	var calls []struct {
		ID      string
		Edition string
		Version string
		Cfg     []dataset.Config
	}
	lockDatasetAPIMockGetVersion.RLock()
	calls = mock.calls.GetVersion
	lockDatasetAPIMockGetVersion.RUnlock()
	return calls
}

// GetVersionMetadata calls GetVersionMetadataFunc.
func (mock *DatasetAPIMock) GetVersionMetadata(id string, edition string, version string, cfg ...dataset.Config) (dataset.Metadata, error) {
	if mock.GetVersionMetadataFunc == nil {
		panic("DatasetAPIMock.GetVersionMetadataFunc: method is nil but DatasetAPI.GetVersionMetadata was just called")
	}
	callInfo := struct {
		ID      string
		Edition string
		Version string
		Cfg     []dataset.Config
	}{
		ID:      id,
		Edition: edition,
		Version: version,
		Cfg:     cfg,
	}
	lockDatasetAPIMockGetVersionMetadata.Lock()
	mock.calls.GetVersionMetadata = append(mock.calls.GetVersionMetadata, callInfo)
	lockDatasetAPIMockGetVersionMetadata.Unlock()
	return mock.GetVersionMetadataFunc(id, edition, version, cfg...)
}

// GetVersionMetadataCalls gets all the calls that were made to GetVersionMetadata.
// Check the length with:
//     len(mockedDatasetAPI.GetVersionMetadataCalls())
func (mock *DatasetAPIMock) GetVersionMetadataCalls() []struct {
	ID      string
	Edition string
	Version string
	Cfg     []dataset.Config
} {
	var calls []struct {
		ID      string
		Edition string
		Version string
		Cfg     []dataset.Config
	}
	lockDatasetAPIMockGetVersionMetadata.RLock()
	calls = mock.calls.GetVersionMetadata
	lockDatasetAPIMockGetVersionMetadata.RUnlock()
	return calls
}

// PutVersion calls PutVersionFunc.
func (mock *DatasetAPIMock) PutVersion(id string, edition string, version string, m dataset.Version, cfg ...dataset.Config) error {
	if mock.PutVersionFunc == nil {
		panic("DatasetAPIMock.PutVersionFunc: method is nil but DatasetAPI.PutVersion was just called")
	}
	callInfo := struct {
		ID      string
		Edition string
		Version string
		M       dataset.Version
		Cfg     []dataset.Config
	}{
		ID:      id,
		Edition: edition,
		Version: version,
		M:       m,
		Cfg:     cfg,
	}
	lockDatasetAPIMockPutVersion.Lock()
	mock.calls.PutVersion = append(mock.calls.PutVersion, callInfo)
	lockDatasetAPIMockPutVersion.Unlock()
	return mock.PutVersionFunc(id, edition, version, m, cfg...)
}

// PutVersionCalls gets all the calls that were made to PutVersion.
// Check the length with:
//     len(mockedDatasetAPI.PutVersionCalls())
func (mock *DatasetAPIMock) PutVersionCalls() []struct {
	ID      string
	Edition string
	Version string
	M       dataset.Version
	Cfg     []dataset.Config
} {
	var calls []struct {
		ID      string
		Edition string
		Version string
		M       dataset.Version
		Cfg     []dataset.Config
	}
	lockDatasetAPIMockPutVersion.RLock()
	calls = mock.calls.PutVersion
	lockDatasetAPIMockPutVersion.RUnlock()
	return calls
}
