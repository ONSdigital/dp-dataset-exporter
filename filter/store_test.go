package filter_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/filter/filtertest"
	"github.com/ONSdigital/dp-dataset-exporter/observation"
	. "github.com/smartystreets/goconvey/convey"
)

const filterAPIURL string = "http://filter-api:8765"
const filterAPIAuthToken string = "dgt-dsfgrtyedf-gesrtrt"
const filterOutputID string = "123456784432"

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

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: mockFilterBody}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

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

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError, Body: nil}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

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

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusBadGateway}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

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

		fileURL := ""
		fileSize := int64(12345)

		mockResponseBody := iOReadCloser{bytes.NewReader([]byte(""))}
		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: mockResponseBody}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(filterOutputID, fileURL, fileSize)

			Convey("The expected body data is sent", func() {
				So(err, ShouldBeNil)

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)

				httpReq := mockHTTPClient.DoCalls()[0].Req
				buf := bytes.Buffer{}
				_, _ = buf.ReadFrom(httpReq.Body)

				actualFilter := &observation.Filter{}
				err := json.Unmarshal(buf.Bytes(), actualFilter)
				So(err, ShouldBeNil)

				So(actualFilter.Downloads.CSV.URL, ShouldEqual, fileURL)
				So(actualFilter.Downloads.CSV.Size, ShouldEqual, strconv.FormatInt(fileSize, 10))
				So(httpReq.URL.Path, ShouldEndWith, filterOutputID)
				So(httpReq.Header.Get("Internal-Token"), ShouldEqual, filterAPIAuthToken)
			})
		})
	})
}

func TestStore_PutCSVData_HTTPNotFoundError(t *testing.T) {

	Convey("Given a store with a mocked HTTP error response", t, func() {

		fileURL := ""
		fileSize := int64(12345)

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusNotFound, Body: nil}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(filterOutputID, fileURL, fileSize)

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

		fileURL := ""
		fileSize := int64(12345)

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError, Body: nil}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(filterOutputID, fileURL, fileSize)

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

		fileURL := ""
		fileSize := int64(12345)

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusBadGateway, Body: nil}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

		Convey("When PutCSVData is called", func() {

			err := filterStore.PutCSVData(filterOutputID, fileURL, fileSize)

			Convey("The expected error is returned", func() {

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, filter.ErrUnrecognisedAPIError)
			})
		})
	})
}

type iOReadCloser struct {
	io.Reader
}

func (iOReadCloser) Close() error { return nil }
