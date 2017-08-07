package observation

import "io"

// Check that the reader conforms to the io.reader interface.
var _ io.Reader = (*Reader)(nil)

// Reader is an io.Reader implementation that wraps a csvRowReader
type Reader struct {
	csvRowReader CSVRowReader
	buffer       []byte // buffer a portion of the current line
	eof          bool   // are we at the end of the csv rows?
}

// NewReader returns a new io.Reader for the given csvRowReader.
func NewReader(csvRowReader CSVRowReader) *Reader {
	return &Reader{
		csvRowReader: csvRowReader,
	}
}

// Read bytes from the underlying csvRowReader
func (reader *Reader) Read(p []byte) (n int, err error) {

	// check if the next line needs to be read.
	if reader.buffer == nil || len(reader.buffer) == 0 {
		csvRow, err := reader.csvRowReader.Read()
		if err == io.EOF {
			reader.eof = true
		} else if err != nil {
			return 0, err
		}

		reader.buffer = []byte(csvRow)
	}

	// copy into the given byte array.
	copied := copy(p, reader.buffer)

	// if the line is bigger than the array, slice the line to account for bytes read
	if len(reader.buffer) > len(p) {
		reader.buffer = reader.buffer[copied:]
	} else { // the line is smaller than the array - clear the current line as it has all been read.
		reader.buffer = nil

		if reader.eof {
			return copied, io.EOF
		}
	}

	return copied, nil
}

// Close the reader.
func (reader *Reader) Close() (err error) {
	return reader.csvRowReader.Close()
}
