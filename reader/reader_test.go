package reader

import (
	"bytes"
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRead(t *testing.T) {
	data := []byte(`Header line
First line
Second line
`)

	Convey("Reader can read all data", t, func() {
		inner := bytes.NewReader(data)
		r := New(inner)

		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, len(data))
		So(buf[:len(data)], ShouldResemble, data)
		So(r.TotalBytesRead(), ShouldEqual, 35)
	})

	Convey("Reader can read all data after calling Peek", t, func() {
		inner := bytes.NewReader(data)
		r := New(inner)

		header, err := r.PeekBytes('\n')
		So(header, ShouldEqual, "Header line")
		So(err, ShouldBeNil)

		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, len(data))
		So(buf[:len(data)], ShouldResemble, data)
	})

	Convey("A second call to Read returns an error", t, func() {
		inner := bytes.NewReader(data)
		r := New(inner)

		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, len(data))
		So(buf[:len(data)], ShouldResemble, data)

		n, err = r.Read(buf)
		So(err, ShouldEqual, io.EOF)
		So(n, ShouldBeZeroValue)
	})

	Convey("Calling Read on empty reader returns an error", t, func() {
		inner := bytes.NewReader(make([]byte, 0))
		r := New(inner)

		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		So(err, ShouldEqual, io.EOF)
		So(n, ShouldBeZeroValue)
	})
}

func TestPeek(t *testing.T) {
	data := []byte(`Header line
First line
Second line
`)

	upToS := []byte(`Header line
First line
`)

	Convey("Reader can successfully peek up to the provided character", t, func() {
		inner := bytes.NewReader(data)
		r := New(inner)

		chars := []struct {
			in  byte
			out string
		}{
			{'\n', "Header line"},
			{' ', "Header"},
			{'S', string(upToS)},
		}

		for _, c := range chars {
			Convey("When the character is "+string(c.in), func() {
				header, err := r.PeekBytes(c.in)
				So(header, ShouldEqual, c.out)
				So(err, ShouldBeNil)
			})
		}

	})

	Convey("PeekBytes returns error", t, func() {

		Convey("When byte not found", func() {
			inner := bytes.NewReader(data)
			r := New(inner)

			header, err := r.PeekBytes('X')
			So(header, ShouldBeEmpty)
			So(err, ShouldEqual, io.EOF)
		})

		Convey("When no data in underlying reader", func() {
			inner := bytes.NewReader(make([]byte, 0))
			r := New(inner)

			header, err := r.PeekBytes('\n')
			So(header, ShouldBeEmpty)
			So(err, ShouldEqual, io.EOF)
		})

		Convey("When called after Read", func() {
			inner := bytes.NewReader(data)
			r := New(inner)

			buf := make([]byte, 1024)
			n, err := r.Read(buf)
			So(err, ShouldEqual, nil)
			So(n, ShouldEqual, len(data))
			So(buf[:len(data)], ShouldResemble, data)

			header, err := r.PeekBytes('\n')
			So(header, ShouldBeEmpty)
			So(err, ShouldEqual, io.EOF)
		})
	})
}
