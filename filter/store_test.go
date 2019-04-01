package filter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-graph/observation"
	"github.com/ONSdigital/go-ns/common/commontest"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	filterAPIURL   = "http://filter-api:8765"
	filterOutputID = "123456784432"

	fileHRef    = "download-url"
	privateLink = "s3-private-link"
	publicLink  = "s3-public-link"
	fileSize    = "12345"
)

var mockDimensionListData = []*observation.DimensionFilter{{
	Name:    "Sex",
	Options: []string{"male"},
}}

var mockFilterData = &observation.Filter{
	FilterID:         filterOutputID,
	InstanceID:       "123",
	DimensionFilters: mockDimensionListData,
}

func TestStore_GetFilter(t *testing.T) {

	mockFilterJSON, _ := json.Marshal(mockFilterData)
	mockFilterBody := iOReadCloser{bytes.NewReader(mockFilterJSON)}

	Convey("Given a store with mocked HTTP responses", t, func() {

		mockHTTPClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: mockFilterBody}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

		Convey("When GetFilter is called", func() {

			filter, err := filterStore.GetFilter(filterOutputID)

			Convey("The expected filter data is returned", func() {
				So(err, ShouldBeNil)

				So(filter.FilterID, ShouldEqual, filterOutputID)
				So(filter.DimensionFilters[0].Name, ShouldEqual, "Sex")
				So(filter.DimensionFilters[0].Options, ShouldResemble, []string{"male"})
			})
		})
	})
}

func TestStore_GetFilter_DimensionListCallError(t *testing.T) {

	Convey("Given a mock http client that returns a 500 error when getting filter output", t, func() {

		mockHTTPClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError, Body: nil}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

		Convey("When GetFilter is called", func() {

			actualFilter, err := filterStore.GetFilter(filterOutputID)

			Convey("The expected error is returned", func() {
				So(actualFilter, ShouldBeNil)
				So(err, ShouldNotBeNil)

				So(err, ShouldEqual, filter.ErrFilterAPIError)
			})
		})
	})
}

func TestStore_GetFilter_FilterCallError(t *testing.T) {

	Convey("Given a mock http client that returns a 404 error when getting filter output", t, func() {

		mockHTTPClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusBadGateway}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

		Convey("When GetFilter is called", func() {

			actualFilter, err := filterStore.GetFilter(filterOutputID)

			Convey("The expected error is returned", func() {
				So(actualFilter, ShouldBeNil)
				So(err, ShouldNotBeNil)

				So(err, ShouldEqual, filter.ErrUnrecognisedAPIError)
			})
		})
	})
}

func TestStore_PutCSVData(t *testing.T) {

	Convey("Given a store with a mocked HTTP response", t, func() {

		mockResponseBody := iOReadCloser{bytes.NewReader([]byte(""))}
		mockHTTPClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: mockResponseBody}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

		Convey("When PutCSVData is called with csv private link", func() {
			csv := &observation.DownloadItem{
				HRef:    fileHRef,
				Private: privateLink,
				Size:    fileSize,
			}

			err := filterStore.PutCSVData(filterOutputID, *csv)

			Convey("The expected body data is sent", func() {
				So(err, ShouldBeNil)

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)

				httpReq := mockHTTPClient.DoCalls()[0].Req
				buf := bytes.Buffer{}
				_, _ = buf.ReadFrom(httpReq.Body)

				actualFilter := &filter.FilterOuput{}
				err := json.Unmarshal(buf.Bytes(), actualFilter)
				So(err, ShouldBeNil)

				So(actualFilter.Downloads.CSV.HRef, ShouldEqual, fileHRef)
				So(actualFilter.Downloads.CSV.Private, ShouldEqual, privateLink)
				So(actualFilter.Downloads.CSV.Public, ShouldBeEmpty)
				So(actualFilter.Downloads.CSV.Size, ShouldEqual, fileSize)
				So(httpReq.URL.Path, ShouldEndWith, filterOutputID)
			})
		})

		Convey("When PutCSVData is called with csv public link", func() {
			csv := &observation.DownloadItem{
				HRef:   fileHRef,
				Public: publicLink,
				Size:   fileSize,
			}

			err := filterStore.PutCSVData(filterOutputID, *csv)

			Convey("The expected body data is sent", func() {
				So(err, ShouldBeNil)

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)

				httpReq := mockHTTPClient.DoCalls()[0].Req
				buf := bytes.Buffer{}
				_, _ = buf.ReadFrom(httpReq.Body)

				actualFilter := &filter.FilterOuput{}
				err := json.Unmarshal(buf.Bytes(), actualFilter)
				So(err, ShouldBeNil)

				So(actualFilter.Downloads.CSV.HRef, ShouldEqual, fileHRef)
				So(actualFilter.Downloads.CSV.Private, ShouldBeEmpty)
				So(actualFilter.Downloads.CSV.Public, ShouldEqual, publicLink)
				So(actualFilter.Downloads.CSV.Size, ShouldEqual, fileSize)
				So(httpReq.URL.Path, ShouldEndWith, filterOutputID)
			})
		})
	})
}

func TestStore_PutCSVData_HTTPNotFoundError(t *testing.T) {

	Convey("Given a store with a mocked HTTP error response", t, func() {
		csv := &observation.DownloadItem{
			HRef:    fileHRef,
			Private: privateLink,
			Size:    fileSize,
		}

		mockHTTPClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusNotFound, Body: nil}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(filterOutputID, *csv)

			Convey("The expected error is returned", func() {

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, filter.ErrFilterJobNotFound)
			})
		})
	})
}

func TestStore_PutCSVData_HTTPInternalServerError(t *testing.T) {

	Convey("Given a store with a mocked HTTP error response", t, func() {
		csv := &observation.DownloadItem{
			HRef:    fileHRef,
			Private: privateLink,
			Size:    fileSize,
		}

		mockHTTPClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError, Body: nil}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(filterOutputID, *csv)

			Convey("The expected error is returned", func() {

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, filter.ErrFilterAPIError)
			})
		})
	})
}

func TestStore_PutCSVData_HTTPUnrecognisedError(t *testing.T) {

	Convey("Given a store with a mocked HTTP error response", t, func() {
		csv := &observation.DownloadItem{
			HRef:    fileHRef,
			Private: privateLink,
			Size:    fileSize,
		}

		mockHTTPClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusBadGateway, Body: nil}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(filterOutputID, *csv)

			Convey("The expected error is returned", func() {

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, filter.ErrUnrecognisedAPIError)
			})
		})
	})
}

func TestStore_PutStateAsEmpty(t *testing.T) {
	Convey("Given a store with a mocked HTTP error response", t, func() {

		mockResponseBody := iOReadCloser{bytes.NewReader([]byte(""))}

		mockHTTPClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: mockResponseBody}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

		Convey("When PutStateAsEmpty is called", func() {

			err := filterStore.PutStateAsEmpty(filterOutputID)

			Convey("Then no error is returned", func() {

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestStore_PutStateAsError(t *testing.T) {
	Convey("Given a store with a mocked HTTP error response", t, func() {

		mockResponseBody := iOReadCloser{bytes.NewReader([]byte(""))}

		mockHTTPClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: mockResponseBody}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

		Convey("When PutStateAsError is called", func() {

			err := filterStore.PutStateAsError(filterOutputID)

			Convey("Then no error is returned", func() {

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)
				So(err, ShouldBeNil)
			})
		})
	})
}

type iOReadCloser struct {
	io.Reader
}

func (iOReadCloser) Close() error { return nil }
