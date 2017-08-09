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
	"strings"
	"testing"
)

const filterAPIURL string = "http://filter-api:8765"
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

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

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

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

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

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

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

		filterStore := filter.NewStore(filterAPIURL, mockHTTPClient)

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

type iOReadCloser struct {
	io.Reader
}

func (iOReadCloser) Close() error { return nil }
