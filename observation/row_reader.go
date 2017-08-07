package observation

import (
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
)

//go:generate moq -out observationtest/bolt_rows.go -pkg observationtest . BoltRows
//go:generate moq -out observationtest/row_reader.go -pkg observationtest . CSVRowReader
type BoltRows bolt.Rows

type CSVRowReader interface {
	Read() (string, error)
	Close() error
}

// Neo4JRowReader translates Neo4j rows to CSV rows.
type BoltRowReader struct {
	rows BoltRows
}

// NewBoltRowReader returns a new reader instace for the given bolt rows.
func NewBoltRowReader(rows BoltRows) *BoltRowReader {
	return &BoltRowReader{
		rows: rows,
	}
}

// NoDataReturnedError is returned if a Neo4j row has no data.
var NoDataReturnedError = errors.New("No data returned in this row.")

// UnrecognisedTypeError is returned if a Neo4j row does not have the expected string value.
var UnrecognisedTypeError = errors.New("The value returned was not a string.")

// Read the next row, or return io.EOF
func (reader *BoltRowReader) Read() (string, error) {
	data, _, err := reader.rows.NextNeo()
	if err != nil {
		return "", err
	}

	if len(data) < 1 {
		return "", NoDataReturnedError
	}

	if csvRow, ok := data[0].(string); ok {
		return csvRow, nil
	}

	return "", UnrecognisedTypeError
}

// Close the reader.
func (reader *BoltRowReader) Close() error {
	return reader.rows.Close()
}
