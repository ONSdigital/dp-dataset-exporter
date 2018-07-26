package reader

import (
	"bufio"
	"io"
)

//go:generate moq -out readertest/reader.go -pkg readertest . WrappedReader

type WrappedReader interface {
	io.Reader
}

type wrappedReader struct {
	r              io.Reader
	buf            []byte
	useBuf         bool
	bufR           *bufio.Reader
	totalBytesRead int
}

//New returns a new Reader whose first line can be read without advancing
//the offset, allowing all of the data to still be read from Read
func New(r io.Reader) *wrappedReader {
	return &wrappedReader{
		r: r,
	}
}

//Read reads data into p. It returns the number of bytes read into p.
//The bytes are taken from at most one Read on the underlying Reader,
//hence n may be less than len(p). At EOF, the count will be zero and
//err will be io.EOF.
func (r *wrappedReader) Read(p []byte) (n int, err error) {
	defer func() {
		r.totalBytesRead += n
	}()

	if !r.useBuf {
		if r.bufR != nil {
			return r.bufR.Read(p)
		}
		return r.r.Read(p)
	}

	n = copy(p, r.buf)
	r.useBuf = false
	r.buf = []byte{}

	if n < len(p) {
		l := io.LimitReader(r.bufR, int64(len(p)-n))
		b, err := l.Read(p[n:])
		n = n + b
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

// PeekBytes reads until the first occurrence of delim in the input,
// returning a slice containing the data up to but not including the delimiter.
// If PeekBytes encounters an error before finding a delimiter,
// it returns the error itself (often io.EOF) and no data.
func (r *wrappedReader) PeekBytes(delim byte) (string, error) {
	r.bufR = bufio.NewReader(r.r)
	b, err := r.bufR.ReadBytes(delim)
	if err != nil {
		return "", err
	}

	r.buf = b
	r.useBuf = true

	return string(b[:len(b)-1]), nil
}

//TotalBytesRead returns the number of bytes which have so far been read
func (r *wrappedReader) TotalBytesRead() int {
	return r.totalBytesRead
}
