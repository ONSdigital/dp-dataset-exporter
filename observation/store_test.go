package observation_test

import (
	"github.com/ONSdigital/dp-dataset-exporter/observation"
	"github.com/ONSdigital/dp-dataset-exporter/observation/observationtest"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSpec(t *testing.T) {

	Convey("Given an store with a mock DB connection", t, func() {

		filter := &observation.Filter{
			DataSetFilterID: "888",
			DimensionFilters: []observation.DimensionFilter{
				{Name: "age", Values: []string{"29", "30"}},
				{Name: "sex", Values: []string{"male", "female"}},
			},
		}

		expectedQuery := "MATCH (age:_888_age), (sex:_888_sex) " +
			"WHERE age.value IN ['29', '30'] AND sex.value IN ['male', 'female'] " +
			"WITH age, sex " +
			"MATCH (o:_888_observation)-[:isValueOf]->(age), (o:_888_observation)-[:isValueOf]->(sex)"

		expectedCSVRow := "the,csv,row"

		mockBoltRows := &observationtest.BoltRowsMock{
			CloseFunc: func() error {
				return nil
			},
			NextNeoFunc: func() ([]interface{}, map[string]interface{}, error) {
				return []interface{}{expectedCSVRow}, nil, nil
			},
		}

		mockedDBConnection := &observationtest.DBConnectionMock{
			QueryNeoFunc: func(query string, params map[string]interface{}) (golangNeo4jBoltDriver.Rows, error) {
				return mockBoltRows, nil
			},
		}

		store := observation.NewStore(mockedDBConnection)

		Convey("When GetCSVRows is called", func() {

			rowReader, err := store.GetCSVRows(filter)

			Convey("The expected query is sent to the database", func() {

				actualQuery := mockedDBConnection.QueryNeoCalls()[0].Query

				So(len(mockedDBConnection.QueryNeoCalls()), ShouldEqual, 1)
				So(actualQuery, ShouldEqual, expectedQuery)
			})

			Convey("There is a row reader returned for the rows given by the database.", func() {
				So(err, ShouldBeNil)
				So(rowReader, ShouldNotBeNil)
			})
		})
	})
}
