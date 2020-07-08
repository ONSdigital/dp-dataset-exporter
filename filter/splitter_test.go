package filter

import (
	filterCli "github.com/ONSdigital/dp-api-clients-go/filter"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSplitter_Split(t *testing.T) {

	filter := &filterCli.Model{
		Dimensions: []filterCli.ModelDimension{
			{Name: "age", Options: []string{"29", "30"}},
			{Name: "sex", Options: []string{"male", "female"}},
		},
	}


	Convey("Given a splitter with a batch size smaller than the number of options", t, func() {

		batchSize := 1
		splitter := NewSplitter(batchSize)

		Convey("When split is called", func() {
			filterParts := splitter.Split(filter)

			Convey("The expected number of filter parts are returned", func() {
				So(len(filterParts), ShouldEqual, 1)
			})
		})
	})


	Convey("Given a splitter with a batch size the same as the number of options", t, func() {

		batchSize := 2
		splitter := NewSplitter(batchSize)

		Convey("When split is called", func() {
			filterParts := splitter.Split(filter)

			Convey("The expected number of filter parts are returned", func() {
				So(len(filterParts), ShouldEqual, 2)
			})
		})
	})

	Convey("Given a splitter with a batch size larger than the number of options", t, func() {

		batchSize := 3
		splitter := NewSplitter(batchSize)

		Convey("When split is called", func() {
			filterParts := splitter.Split(filter)

			Convey("The expected number of filter parts are returned", func() {
				So(len(filterParts), ShouldEqual, 2)
			})
		})
	})
}
