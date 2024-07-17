// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-dataset-exporter/service"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"sync"
)

// Ensure, that DatasetAPIMock does implement service.DatasetAPI.
// If this is not the case, regenerate this file with moq.
var _ service.DatasetAPI = &DatasetAPIMock{}

// DatasetAPIMock is a mock implementation of service.DatasetAPI.
//
//	func TestSomethingThatUsesDatasetAPI(t *testing.T) {
//
//		// make and configure a mocked service.DatasetAPI
//		mockedDatasetAPI := &DatasetAPIMock{
//			CheckerFunc: func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
//				panic("mock out the Checker method")
//			},
//			GetInstanceFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
//				panic("mock out the GetInstance method")
//			},
//			GetMetadataURLFunc: func(id string, edition string, version string) string {
//				panic("mock out the GetMetadataURL method")
//			},
//			GetOptionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, q *dataset.QueryParams) (dataset.Options, error) {
//				panic("mock out the GetOptions method")
//			},
//			GetOptionsInBatchesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, batchSize int, maxWorkers int) (dataset.Options, error) {
//				panic("mock out the GetOptionsInBatches method")
//			},
//			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
//				panic("mock out the GetVersion method")
//			},
//			GetVersionDimensionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.VersionDimensions, error) {
//				panic("mock out the GetVersionDimensions method")
//			},
//			GetVersionMetadataFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.Metadata, error) {
//				panic("mock out the GetVersionMetadata method")
//			},
//			PutVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string, edition string, version string, m dataset.Version) error {
//				panic("mock out the PutVersion method")
//			},
//		}
//
//		// use mockedDatasetAPI in code that requires service.DatasetAPI
//		// and then make assertions.
//
//	}
type DatasetAPIMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error

	// GetInstanceFunc mocks the GetInstance method.
	GetInstanceFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error)

	// GetMetadataURLFunc mocks the GetMetadataURL method.
	GetMetadataURLFunc func(id string, edition string, version string) string

	// GetOptionsFunc mocks the GetOptions method.
	GetOptionsFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, q *dataset.QueryParams) (dataset.Options, error)

	// GetOptionsInBatchesFunc mocks the GetOptionsInBatches method.
	GetOptionsInBatchesFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, batchSize int, maxWorkers int) (dataset.Options, error)

	// GetVersionFunc mocks the GetVersion method.
	GetVersionFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error)

	// GetVersionDimensionsFunc mocks the GetVersionDimensions method.
	GetVersionDimensionsFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.VersionDimensions, error)

	// GetVersionMetadataFunc mocks the GetVersionMetadata method.
	GetVersionMetadataFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.Metadata, error)

	// PutVersionFunc mocks the PutVersion method.
	PutVersionFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string, edition string, version string, m dataset.Version) error

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// CheckState is the checkState argument value.
			CheckState *healthcheck.CheckState
		}
		// GetInstance holds details about calls to the GetInstance method.
		GetInstance []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// InstanceID is the instanceID argument value.
			InstanceID string
			// IfMatch is the ifMatch argument value.
			IfMatch string
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
		// GetOptions holds details about calls to the GetOptions method.
		GetOptions []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
			// Dimension is the dimension argument value.
			Dimension string
			// Q is the q argument value.
			Q *dataset.QueryParams
		}
		// GetOptionsInBatches holds details about calls to the GetOptionsInBatches method.
		GetOptionsInBatches []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
			// Dimension is the dimension argument value.
			Dimension string
			// BatchSize is the batchSize argument value.
			BatchSize int
			// MaxWorkers is the maxWorkers argument value.
			MaxWorkers int
		}
		// GetVersion holds details about calls to the GetVersion method.
		GetVersion []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// DownloadServiceAuthToken is the downloadServiceAuthToken argument value.
			DownloadServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// DatasetID is the datasetID argument value.
			DatasetID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
		}
		// GetVersionDimensions holds details about calls to the GetVersionDimensions method.
		GetVersionDimensions []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
		}
		// GetVersionMetadata holds details about calls to the GetVersionMetadata method.
		GetVersionMetadata []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
		}
		// PutVersion holds details about calls to the PutVersion method.
		PutVersion []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// DatasetID is the datasetID argument value.
			DatasetID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
			// M is the m argument value.
			M dataset.Version
		}
	}
	lockChecker              sync.RWMutex
	lockGetInstance          sync.RWMutex
	lockGetMetadataURL       sync.RWMutex
	lockGetOptions           sync.RWMutex
	lockGetOptionsInBatches  sync.RWMutex
	lockGetVersion           sync.RWMutex
	lockGetVersionDimensions sync.RWMutex
	lockGetVersionMetadata   sync.RWMutex
	lockPutVersion           sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *DatasetAPIMock) Checker(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("DatasetAPIMock.CheckerFunc: method is nil but DatasetAPI.Checker was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}{
		ContextMoqParam: contextMoqParam,
		CheckState:      checkState,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(contextMoqParam, checkState)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//
//	len(mockedDatasetAPI.CheckerCalls())
func (mock *DatasetAPIMock) CheckerCalls() []struct {
	ContextMoqParam context.Context
	CheckState      *healthcheck.CheckState
} {
	var calls []struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// GetInstance calls GetInstanceFunc.
func (mock *DatasetAPIMock) GetInstance(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, instanceID string, ifMatch string) (dataset.Instance, string, error) {
	if mock.GetInstanceFunc == nil {
		panic("DatasetAPIMock.GetInstanceFunc: method is nil but DatasetAPI.GetInstance was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		InstanceID       string
		IfMatch          string
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		InstanceID:       instanceID,
		IfMatch:          ifMatch,
	}
	mock.lockGetInstance.Lock()
	mock.calls.GetInstance = append(mock.calls.GetInstance, callInfo)
	mock.lockGetInstance.Unlock()
	return mock.GetInstanceFunc(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, ifMatch)
}

// GetInstanceCalls gets all the calls that were made to GetInstance.
// Check the length with:
//
//	len(mockedDatasetAPI.GetInstanceCalls())
func (mock *DatasetAPIMock) GetInstanceCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	InstanceID       string
	IfMatch          string
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		InstanceID       string
		IfMatch          string
	}
	mock.lockGetInstance.RLock()
	calls = mock.calls.GetInstance
	mock.lockGetInstance.RUnlock()
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
	mock.lockGetMetadataURL.Lock()
	mock.calls.GetMetadataURL = append(mock.calls.GetMetadataURL, callInfo)
	mock.lockGetMetadataURL.Unlock()
	return mock.GetMetadataURLFunc(id, edition, version)
}

// GetMetadataURLCalls gets all the calls that were made to GetMetadataURL.
// Check the length with:
//
//	len(mockedDatasetAPI.GetMetadataURLCalls())
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
	mock.lockGetMetadataURL.RLock()
	calls = mock.calls.GetMetadataURL
	mock.lockGetMetadataURL.RUnlock()
	return calls
}

// GetOptions calls GetOptionsFunc.
func (mock *DatasetAPIMock) GetOptions(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, q *dataset.QueryParams) (dataset.Options, error) {
	if mock.GetOptionsFunc == nil {
		panic("DatasetAPIMock.GetOptionsFunc: method is nil but DatasetAPI.GetOptions was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
		Dimension        string
		Q                *dataset.QueryParams
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ID:               id,
		Edition:          edition,
		Version:          version,
		Dimension:        dimension,
		Q:                q,
	}
	mock.lockGetOptions.Lock()
	mock.calls.GetOptions = append(mock.calls.GetOptions, callInfo)
	mock.lockGetOptions.Unlock()
	return mock.GetOptionsFunc(ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension, q)
}

// GetOptionsCalls gets all the calls that were made to GetOptions.
// Check the length with:
//
//	len(mockedDatasetAPI.GetOptionsCalls())
func (mock *DatasetAPIMock) GetOptionsCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ID               string
	Edition          string
	Version          string
	Dimension        string
	Q                *dataset.QueryParams
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
		Dimension        string
		Q                *dataset.QueryParams
	}
	mock.lockGetOptions.RLock()
	calls = mock.calls.GetOptions
	mock.lockGetOptions.RUnlock()
	return calls
}

// GetOptionsInBatches calls GetOptionsInBatchesFunc.
func (mock *DatasetAPIMock) GetOptionsInBatches(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, batchSize int, maxWorkers int) (dataset.Options, error) {
	if mock.GetOptionsInBatchesFunc == nil {
		panic("DatasetAPIMock.GetOptionsInBatchesFunc: method is nil but DatasetAPI.GetOptionsInBatches was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
		Dimension        string
		BatchSize        int
		MaxWorkers       int
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ID:               id,
		Edition:          edition,
		Version:          version,
		Dimension:        dimension,
		BatchSize:        batchSize,
		MaxWorkers:       maxWorkers,
	}
	mock.lockGetOptionsInBatches.Lock()
	mock.calls.GetOptionsInBatches = append(mock.calls.GetOptionsInBatches, callInfo)
	mock.lockGetOptionsInBatches.Unlock()
	return mock.GetOptionsInBatchesFunc(ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension, batchSize, maxWorkers)
}

// GetOptionsInBatchesCalls gets all the calls that were made to GetOptionsInBatches.
// Check the length with:
//
//	len(mockedDatasetAPI.GetOptionsInBatchesCalls())
func (mock *DatasetAPIMock) GetOptionsInBatchesCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ID               string
	Edition          string
	Version          string
	Dimension        string
	BatchSize        int
	MaxWorkers       int
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
		Dimension        string
		BatchSize        int
		MaxWorkers       int
	}
	mock.lockGetOptionsInBatches.RLock()
	calls = mock.calls.GetOptionsInBatches
	mock.lockGetOptionsInBatches.RUnlock()
	return calls
}

// GetVersion calls GetVersionFunc.
func (mock *DatasetAPIMock) GetVersion(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
	if mock.GetVersionFunc == nil {
		panic("DatasetAPIMock.GetVersionFunc: method is nil but DatasetAPI.GetVersion was just called")
	}
	callInfo := struct {
		Ctx                      context.Context
		UserAuthToken            string
		ServiceAuthToken         string
		DownloadServiceAuthToken string
		CollectionID             string
		DatasetID                string
		Edition                  string
		Version                  string
	}{
		Ctx:                      ctx,
		UserAuthToken:            userAuthToken,
		ServiceAuthToken:         serviceAuthToken,
		DownloadServiceAuthToken: downloadServiceAuthToken,
		CollectionID:             collectionID,
		DatasetID:                datasetID,
		Edition:                  edition,
		Version:                  version,
	}
	mock.lockGetVersion.Lock()
	mock.calls.GetVersion = append(mock.calls.GetVersion, callInfo)
	mock.lockGetVersion.Unlock()
	return mock.GetVersionFunc(ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version)
}

// GetVersionCalls gets all the calls that were made to GetVersion.
// Check the length with:
//
//	len(mockedDatasetAPI.GetVersionCalls())
func (mock *DatasetAPIMock) GetVersionCalls() []struct {
	Ctx                      context.Context
	UserAuthToken            string
	ServiceAuthToken         string
	DownloadServiceAuthToken string
	CollectionID             string
	DatasetID                string
	Edition                  string
	Version                  string
} {
	var calls []struct {
		Ctx                      context.Context
		UserAuthToken            string
		ServiceAuthToken         string
		DownloadServiceAuthToken string
		CollectionID             string
		DatasetID                string
		Edition                  string
		Version                  string
	}
	mock.lockGetVersion.RLock()
	calls = mock.calls.GetVersion
	mock.lockGetVersion.RUnlock()
	return calls
}

// GetVersionDimensions calls GetVersionDimensionsFunc.
func (mock *DatasetAPIMock) GetVersionDimensions(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.VersionDimensions, error) {
	if mock.GetVersionDimensionsFunc == nil {
		panic("DatasetAPIMock.GetVersionDimensionsFunc: method is nil but DatasetAPI.GetVersionDimensions was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ID:               id,
		Edition:          edition,
		Version:          version,
	}
	mock.lockGetVersionDimensions.Lock()
	mock.calls.GetVersionDimensions = append(mock.calls.GetVersionDimensions, callInfo)
	mock.lockGetVersionDimensions.Unlock()
	return mock.GetVersionDimensionsFunc(ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version)
}

// GetVersionDimensionsCalls gets all the calls that were made to GetVersionDimensions.
// Check the length with:
//
//	len(mockedDatasetAPI.GetVersionDimensionsCalls())
func (mock *DatasetAPIMock) GetVersionDimensionsCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ID               string
	Edition          string
	Version          string
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
	}
	mock.lockGetVersionDimensions.RLock()
	calls = mock.calls.GetVersionDimensions
	mock.lockGetVersionDimensions.RUnlock()
	return calls
}

// GetVersionMetadata calls GetVersionMetadataFunc.
func (mock *DatasetAPIMock) GetVersionMetadata(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.Metadata, error) {
	if mock.GetVersionMetadataFunc == nil {
		panic("DatasetAPIMock.GetVersionMetadataFunc: method is nil but DatasetAPI.GetVersionMetadata was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ID:               id,
		Edition:          edition,
		Version:          version,
	}
	mock.lockGetVersionMetadata.Lock()
	mock.calls.GetVersionMetadata = append(mock.calls.GetVersionMetadata, callInfo)
	mock.lockGetVersionMetadata.Unlock()
	return mock.GetVersionMetadataFunc(ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version)
}

// GetVersionMetadataCalls gets all the calls that were made to GetVersionMetadata.
// Check the length with:
//
//	len(mockedDatasetAPI.GetVersionMetadataCalls())
func (mock *DatasetAPIMock) GetVersionMetadataCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ID               string
	Edition          string
	Version          string
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
	}
	mock.lockGetVersionMetadata.RLock()
	calls = mock.calls.GetVersionMetadata
	mock.lockGetVersionMetadata.RUnlock()
	return calls
}

// PutVersion calls PutVersionFunc.
func (mock *DatasetAPIMock) PutVersion(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string, edition string, version string, m dataset.Version) error {
	if mock.PutVersionFunc == nil {
		panic("DatasetAPIMock.PutVersionFunc: method is nil but DatasetAPI.PutVersion was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		DatasetID        string
		Edition          string
		Version          string
		M                dataset.Version
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		DatasetID:        datasetID,
		Edition:          edition,
		Version:          version,
		M:                m,
	}
	mock.lockPutVersion.Lock()
	mock.calls.PutVersion = append(mock.calls.PutVersion, callInfo)
	mock.lockPutVersion.Unlock()
	return mock.PutVersionFunc(ctx, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version, m)
}

// PutVersionCalls gets all the calls that were made to PutVersion.
// Check the length with:
//
//	len(mockedDatasetAPI.PutVersionCalls())
func (mock *DatasetAPIMock) PutVersionCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	DatasetID        string
	Edition          string
	Version          string
	M                dataset.Version
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		DatasetID        string
		Edition          string
		Version          string
		M                dataset.Version
	}
	mock.lockPutVersion.RLock()
	calls = mock.calls.PutVersion
	mock.lockPutVersion.RUnlock()
	return calls
}
