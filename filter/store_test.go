package filter_test

import (
	"bytes"
	"encoding/json"
	"github.com/ONSdigital/dp-dataset-exporter/filter"
	"github.com/ONSdigital/dp-dataset-exporter/filter/filtertest"
	"github.com/ONSdigital/dp-dataset-exporter/observation"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

const filterAPIURL string = "http://filter-api:8765"
const filterAPIAuthToken string = "dgt-dsfgrtyedf-gesrtrt"
const filterJobID string = "123456784432"

var mockFilterData = &observation.Filter{
	JobID:            filterJobID,
	DataSetFilterID:  "123",
	DimensionListURL: filterAPIURL + "/filter/" + filterJobID + "/dimensions",
}

var mockDimensionListData = []*observation.DimensionFilter{{
	URL:  filterAPIURL + "/filter/" + filterJobID + "/dimensions/234",
	Name: "Sex",
}}

var mockDimensionData = &observation.DimensionFilter{
	URL:    filterAPIURL + "/filter/" + filterJobID + "/dimensions/234",
	Name:   "Sex",
	Values: []string{"Male", "Female"},
}

func TestStore_GetFilter(t *testing.T) {

	mockFilterJSON, _ := json.Marshal(mockFilterData)
	mockFilterBody := iOReadCloser{bytes.NewReader(mockFilterJSON)}

	mockDimensionListJSON, _ := json.Marshal(mockDimensionListData)
	mockDimensionListBody := iOReadCloser{bytes.NewReader(mockDimensionListJSON)}

	mockDimensionJSON, _ := json.Marshal(mockDimensionData)
	mockDimensionBody := iOReadCloser{bytes.NewReader(mockDimensionJSON)}

	Convey("Given a store with mocked HTTP responses", t, func() {

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {

				if strings.Contains(req.URL.Path, "/options") {
					return &http.Response{StatusCode: http.StatusOK, Body: mockDimensionBody}, nil
				}

				if strings.Contains(req.URL.Path, "/dimensions") {
					return &http.Response{StatusCode: http.StatusOK, Body: mockDimensionListBody}, nil
				}

				return &http.Response{StatusCode: http.StatusOK, Body: mockFilterBody}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

		Convey("When GetFilter is called", func() {

			filter, err := filterStore.GetFilter(filterJobID)

			Convey("The expected filter data is returned", func() {
				So(err, ShouldBeNil)

				So(filter.JobID, ShouldEqual, filterJobID)
				So(filter.DimensionFilters[0].Name, ShouldEqual, mockDimensionData.Name)
				So(filter.DimensionFilters[0].Values[0], ShouldEqual, mockDimensionData.Values[0])
				So(filter.DimensionFilters[0].Values[1], ShouldEqual, mockDimensionData.Values[1])
			})
		})
	})
}

func TestStore_GetFilter_DimensionCallError(t *testing.T) {

	mockFilterJSON, _ := json.Marshal(mockFilterData)
	mockFilterBody := iOReadCloser{bytes.NewReader(mockFilterJSON)}

	mockDimensionListJSON, _ := json.Marshal(mockDimensionListData)
	mockDimensionListBody := iOReadCloser{bytes.NewReader(mockDimensionListJSON)}

	Convey("Given a mock http client that returns a 404 error when getting dimension options", t, func() {

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {

				if strings.Contains(req.URL.Path, "/options") {
					return &http.Response{StatusCode: http.StatusNotFound}, nil
				}

				if strings.Contains(req.URL.Path, "/dimensions") {
					return &http.Response{StatusCode: http.StatusOK, Body: mockDimensionListBody}, nil
				}

				return &http.Response{StatusCode: http.StatusOK, Body: mockFilterBody}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

		Convey("When GetFilter is called", func() {

			actualFilter, err := filterStore.GetFilter(filterJobID)

			Convey("The expected error is returned", func() {
				So(actualFilter, ShouldBeNil)
				So(err, ShouldNotBeNil)

				So(err, ShouldEqual, filter.ErrFilterJobNotFound)
			})
		})
	})
}

func TestStore_GetFilter_DimensionListCallError(t *testing.T) {

	mockFilterJSON, _ := json.Marshal(mockFilterData)
	mockFilterBody := iOReadCloser{bytes.NewReader(mockFilterJSON)}

	Convey("Given a mock http client that returns a 500 error when getting the dimension list", t, func() {

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {

				if strings.Contains(req.URL.Path, "/dimensions") {
					return &http.Response{StatusCode: http.StatusInternalServerError}, nil
				}

				return &http.Response{StatusCode: http.StatusOK, Body: mockFilterBody}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

		Convey("When GetFilter is called", func() {

			actualFilter, err := filterStore.GetFilter(filterJobID)

			Convey("The expected error is returned", func() {
				So(actualFilter, ShouldBeNil)
				So(err, ShouldNotBeNil)

				So(err, ShouldEqual, filter.ErrFilterAPIError)
			})
		})
	})
}

func TestStore_GetFilter_FilterCallError(t *testing.T) {

	Convey("Given a mock http client that returns a 404 error when getting the dimension list", t, func() {

		mockHTTPClient := &filtertest.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusBadGateway}, nil
			},
		}

		filterStore := filter.NewStore(filterAPIURL, filterAPIAuthToken, mockHTTPClient)

		Convey("When GetFilter is called", func() {

			actualFilter, err := filterStore.GetFilter(filterJobID)

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

			err := filterStore.PutCSVData(filterJobID, fileURL, fileSize)

			Convey("The expected body data is sent", func() {
				So(err, ShouldBeNil)

				So(len(mockHTTPClient.DoCalls()), ShouldEqual, 1)

				httpReq := mockHTTPClient.DoCalls()[0].Req
				buf := bytes.Buffer{}
				buf.ReadFrom(httpReq.Body)

				actualFilter := &observation.Filter{}
				err := json.Unmarshal(buf.Bytes(), actualFilter)
				So(err, ShouldBeNil)

				So(actualFilter.Downloads.CSV.URL, ShouldEqual, fileURL)
				So(actualFilter.Downloads.CSV.Size, ShouldEqual, strconv.FormatInt(fileSize, 10))
				So(httpReq.URL.Path, ShouldEndWith, filterJobID)
				So(httpReq.Header.Get("internal_token"), ShouldEqual, filterAPIAuthToken)
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

			err := filterStore.PutCSVData(filterJobID, fileURL, fileSize)

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

			err := filterStore.PutCSVData(filterJobID, fileURL, fileSize)

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

			err := filterStore.PutCSVData(filterJobID, fileURL, fileSize)

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
