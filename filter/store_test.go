package filter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	filterCli "github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/filter/filtertest"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	filterOutputID = "123456784432"
	fileURL        = "download-url"
	privateLink    = "s3-private-link"
	publicLink     = "s3-public-link"
	fileSize       = "12345"
)

var mockDimensionListData = []filterCli.ModelDimension{{
	Name:    "Sex",
	Options: []string{"male"},
}}

var mockFilterData = &filterCli.Model{
	FilterID:   filterOutputID,
	InstanceID: "123",
	Dimensions: mockDimensionListData,
}

var ctx = context.Background()

func TestStore_GetFilter(t *testing.T) {
	mockFilterJSON, _ := json.Marshal(mockFilterData)
	validServiceToken := "validServiceAuthToken"

	Convey("Given a store with mocked filter client responses", t, func() {

		mockFilterClient := &filtertest.ClientMock{
			GetOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, collectionID string, filterOutputID string) ([]byte, error) {
				return mockFilterJSON, nil
			},
		}

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When GetFilter is called", func() {

			filter, err := filterStore.GetFilter(ctx, filterOutputID)

			Convey("The expected filter data is returned", func() {
				So(err, ShouldBeNil)

				So(filter.FilterID, ShouldEqual, filterOutputID)
				So(filter.Dimensions[0].Name, ShouldEqual, "Sex")
				So(filter.Dimensions[0].Options, ShouldResemble, []string{"male"})
			})

			Convey("Filter client is called with the valid serviceAuthToken", func() {
				So(len(mockFilterClient.GetOutputBytesCalls()), ShouldEqual, 1)
				So(mockFilterClient.GetOutputBytesCalls()[0].FilterOutputID, ShouldEqual, filterOutputID)
				So(mockFilterClient.GetOutputBytesCalls()[0].ServiceAuthToken, ShouldEqual, validServiceToken)
			})
		})
	})
}

func TestStore_GetFilter_DimensionListCallError(t *testing.T) {
	Convey("Given a mock filter client that returns a 500 error when getting filter output", t, func() {

		mockFilterClient := &filtertest.ClientMock{
			GetOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, collectionID string, filterOutputID string) ([]byte, error) {
				uri := fmt.Sprintf("%s/filter-outputs/%s", "myURL", filterOutputID)
				return nil, &filterCli.ErrInvalidFilterAPIResponse{
					ExpectedCode: http.StatusOK,
					ActualCode:   http.StatusInternalServerError,
					URI:          uri}
			},
		}
		validServiceToken := "validServiceAuthToken"

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When GetFilter is called", func() {

			actualFilter, err := filterStore.GetFilter(ctx, filterOutputID)

			Convey("The expected error is returned", func() {
				So(actualFilter, ShouldBeNil)
				So(err, ShouldEqual, filter.ErrFilterAPIError)
			})
		})
	})
}

func TestStore_GetFilter_FilterCallError(t *testing.T) {

	Convey("Given a mock filter client that returns a StatusBadGateway error when getting filter output", t, func() {

		mockFilterClient := &filtertest.ClientMock{
			GetOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, collectionID string, filterOutputID string) ([]byte, error) {
				uri := fmt.Sprintf("%s/filter-outputs/%s", "myURL", filterOutputID)
				return nil, &filterCli.ErrInvalidFilterAPIResponse{
					ExpectedCode: http.StatusOK,
					ActualCode:   http.StatusBadGateway,
					URI:          uri}
			},
		}
		validServiceToken := "validServiceAuthToken"

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When GetFilter is called", func() {

			actualFilter, err := filterStore.GetFilter(ctx, filterOutputID)

			Convey("The expected error is returned", func() {
				So(actualFilter, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, filter.ErrUnrecognisedAPIError)
			})
		})
	})
}

func TestStore_GetFilter_FilterNotFound(t *testing.T) {

	Convey("Given a mock filter client that returns a 404 error when getting filter output", t, func() {

		mockFilterClient := &filtertest.ClientMock{
			GetOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, collectionID string, filterOutputID string) ([]byte, error) {
				uri := fmt.Sprintf("%s/filter-outputs/%s", "myURL", filterOutputID)
				return nil, &filterCli.ErrInvalidFilterAPIResponse{
					ExpectedCode: http.StatusOK,
					ActualCode:   http.StatusNotFound,
					URI:          uri}
			},
		}
		validServiceToken := "validServiceAuthToken"

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When GetFilter is called", func() {

			actualFilter, err := filterStore.GetFilter(ctx, filterOutputID)

			Convey("The expected error is returned", func() {
				So(actualFilter, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, filter.ErrFilterJobNotFound)
			})
		})
	})
}

func TestStore_PutCSVData(t *testing.T) {

	Convey("Given a store with a mocked filter client response", t, func() {

		mockFilterClient := &filtertest.ClientMock{
			UpdateFilterOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, filterJobID string, b []byte) error {
				return nil
			},
		}
		validServiceToken := "validServiceAuthToken"

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When PutCSVData is called with csv private link", func() {
			csv := &filterCli.Download{
				URL:     fileURL,
				Private: privateLink,
				Size:    fileSize,
			}

			err := filterStore.PutCSVData(ctx, filterOutputID, *csv)

			Convey("The expected body data is sent", func() {
				So(err, ShouldBeNil)

				// Validate filter output id
				So(len(mockFilterClient.UpdateFilterOutputBytesCalls()), ShouldEqual, 1)
				So(mockFilterClient.UpdateFilterOutputBytesCalls()[0].FilterJobID, ShouldEqual, filterOutputID)

				// Validate filter output sent
				actualFilter := &filterCli.Model{}
				err := json.Unmarshal(mockFilterClient.UpdateFilterOutputBytesCalls()[0].B, actualFilter)
				So(err, ShouldBeNil)
				So(actualFilter.Downloads["CSV"].URL, ShouldEqual, fileURL)
				So(actualFilter.Downloads["CSV"].Private, ShouldEqual, privateLink)
				So(actualFilter.Downloads["CSV"].Public, ShouldBeEmpty)
				So(actualFilter.Downloads["CSV"].Size, ShouldEqual, fileSize)
			})
		})

		Convey("When PutCSVData is called with csv public link", func() {
			csv := &filterCli.Download{
				URL:    fileURL,
				Public: publicLink,
				Size:   fileSize,
			}

			err := filterStore.PutCSVData(ctx, filterOutputID, *csv)

			Convey("The expected body data is sent", func() {
				So(err, ShouldBeNil)

				// Validate filter output id
				So(len(mockFilterClient.UpdateFilterOutputBytesCalls()), ShouldEqual, 1)
				So(mockFilterClient.UpdateFilterOutputBytesCalls()[0].FilterJobID, ShouldEqual, filterOutputID)

				// Validate filter output sent
				actualFilter := &filterCli.Model{}
				err := json.Unmarshal(mockFilterClient.UpdateFilterOutputBytesCalls()[0].B, actualFilter)
				So(err, ShouldBeNil)
				So(actualFilter.Downloads["CSV"].URL, ShouldEqual, fileURL)
				So(actualFilter.Downloads["CSV"].Private, ShouldBeEmpty)
				So(actualFilter.Downloads["CSV"].Public, ShouldEqual, publicLink)
				So(actualFilter.Downloads["CSV"].Size, ShouldEqual, fileSize)
			})
		})
	})
}

func TestStore_PutCSVData_HTTPNotFoundError(t *testing.T) {

	Convey("Given a store with a mocked filter client with StatusNotFound error response", t, func() {
		csv := &filterCli.Download{
			URL:     fileURL,
			Private: privateLink,
			Size:    fileSize,
		}

		mockFilterClient := &filtertest.ClientMock{
			UpdateFilterOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, filterJobID string, b []byte) error {
				uri := fmt.Sprintf("%s/filter-outputs/%s", "myURL", filterJobID)
				return &filterCli.ErrInvalidFilterAPIResponse{
					ExpectedCode: http.StatusOK,
					ActualCode:   http.StatusNotFound,
					URI:          uri}
			},
		}
		validServiceToken := "validServiceAuthToken"

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(ctx, filterOutputID, *csv)

			Convey("The expected error is returned", func() {
				So(len(mockFilterClient.UpdateFilterOutputBytesCalls()), ShouldEqual, 1)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, filter.ErrFilterJobNotFound)
			})
		})
	})
}

func TestStore_PutCSVData_HTTPInternalServerError(t *testing.T) {

	Convey("Given a store with a mocked filter client with StatusInternalServerError error response", t, func() {
		csv := &filterCli.Download{
			URL:     fileURL,
			Private: privateLink,
			Size:    fileSize,
		}

		mockFilterClient := &filtertest.ClientMock{
			UpdateFilterOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, filterJobID string, b []byte) error {
				uri := fmt.Sprintf("%s/filter-outputs/%s", "myURL", filterJobID)
				return &filterCli.ErrInvalidFilterAPIResponse{
					ExpectedCode: http.StatusOK,
					ActualCode:   http.StatusInternalServerError,
					URI:          uri}
			},
		}
		validServiceToken := "validServiceAuthToken"

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(ctx, filterOutputID, *csv)

			Convey("The expected error is returned", func() {
				So(len(mockFilterClient.UpdateFilterOutputBytesCalls()), ShouldEqual, 1)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, filter.ErrFilterAPIError)
			})
		})
	})
}

func TestStore_PutCSVData_HTTPUnrecognisedError(t *testing.T) {

	Convey("Given a store with a mocked filter client with StatusBadGateway error response", t, func() {
		csv := &filterCli.Download{
			URL:     fileURL,
			Private: privateLink,
			Size:    fileSize,
		}

		mockFilterClient := &filtertest.ClientMock{
			UpdateFilterOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, filterJobID string, b []byte) error {
				uri := fmt.Sprintf("%s/filter-outputs/%s", "myURL", filterJobID)
				return &filterCli.ErrInvalidFilterAPIResponse{
					ExpectedCode: http.StatusOK,
					ActualCode:   http.StatusBadGateway,
					URI:          uri}
			},
		}
		validServiceToken := "validServiceAuthToken"

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(ctx, filterOutputID, *csv)

			Convey("The expected error is returned", func() {
				So(len(mockFilterClient.UpdateFilterOutputBytesCalls()), ShouldEqual, 1)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, filter.ErrUnrecognisedAPIError)
			})
		})
	})
}

func TestStore_PutStateAsEmpty(t *testing.T) {
	Convey("Given a store with a mocked filter client response", t, func() {

		mockFilterClient := &filtertest.ClientMock{
			UpdateFilterOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, filterJobID string, b []byte) error {
				return nil
			},
		}
		validServiceToken := "validServiceAuthToken"

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When PutStateAsEmpty is called", func() {

			err := filterStore.PutStateAsEmpty(ctx, filterOutputID)

			Convey("Then no error is returned and filterOutput is set to 'completed' state", func() {
				So(len(mockFilterClient.UpdateFilterOutputBytesCalls()), ShouldEqual, 1)
				So(err, ShouldBeNil)

				// Validate filter output sent
				actualFilter := &filterCli.Model{}
				err := json.Unmarshal(mockFilterClient.UpdateFilterOutputBytesCalls()[0].B, actualFilter)
				So(err, ShouldBeNil)
				So(actualFilter.State, ShouldEqual, "completed")
			})
		})
	})
}

func TestStore_PutStateAsError(t *testing.T) {
	Convey("Given a store with a mocked filter client response", t, func() {
		mockFilterClient := &filtertest.ClientMock{
			UpdateFilterOutputBytesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceToken string, filterJobID string, b []byte) error {
				return nil
			},
		}
		validServiceToken := "validServiceAuthToken"

		filterStore := filter.NewStore(mockFilterClient, validServiceToken)

		Convey("When PutStateAsError is called", func() {

			err := filterStore.PutStateAsError(ctx, filterOutputID)

			Convey("Then no error is returned and the filterOutput is set to 'failed' state", func() {
				So(len(mockFilterClient.UpdateFilterOutputBytesCalls()), ShouldEqual, 1)
				So(err, ShouldBeNil)

				// Validate filter output sent
				actualFilter := &filterCli.Model{}
				err := json.Unmarshal(mockFilterClient.UpdateFilterOutputBytesCalls()[0].B, actualFilter)
				So(err, ShouldBeNil)
				So(actualFilter.State, ShouldEqual, "failed")
			})
		})
	})
}
