package observation

import (
	"bytes"
	"fmt"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

//go:generate moq -out observationtest/db_connection.go -pkg observationtest . DBConnection

// Store represents storage for observation data.
type Store struct {
	dBConnection DBConnection
}

// DBConnection provides a connection to the database.
type DBConnection interface {
	QueryNeo(query string, params map[string]interface{}) (bolt.Rows, error)
}

// NewStore returns a new store instace using the given DB connection.
func NewStore(dBConnection DBConnection) *Store {
	return &Store{
		dBConnection: dBConnection,
	}
}

// GetCSVRows returns a reader allowing individual CSV rows to be read.
func (store *Store) GetCSVRows(filter *Filter) (CSVRowReader, error) {

	query := createObservationQuery(filter)

	rows, err := store.dBConnection.QueryNeo(query, nil)
	if err != nil {
		return nil, err
	}

	return NewBoltRowReader(rows), nil
}

func createObservationQuery(filter *Filter) string {

	matchDimensions := "MATCH "
	where := " WHERE "
	with := " WITH "
	match := " MATCH "

	for index, dimension := range filter.DimensionFilters {

		if index != 0 {
			matchDimensions += ", "
			where += " AND "
			with += ", "
			match += ", "
		}

		optionList := createOptionList(dimension.Values)
		matchDimensions += fmt.Sprintf("(%s:_%s_%s)", dimension.Name, filter.DataSetFilterID, dimension.Name)
		where += fmt.Sprintf("%s.value IN [%s]", dimension.Name, optionList)
		with += dimension.Name
		match += fmt.Sprintf("(o:_%s_observation)-[:isValueOf]->(%s)", filter.DataSetFilterID, dimension.Name)
	}

	return matchDimensions + where + with + match
}

func createOptionList(options []string) string {

	var buffer bytes.Buffer

	for index, option := range options {

		if index != 0 {
			buffer.WriteString(", ")
		}

		buffer.WriteString(fmt.Sprintf("'%s'", option))
	}

	return buffer.String()
}
