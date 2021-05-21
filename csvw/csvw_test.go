package csvw

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	. "github.com/smartystreets/goconvey/convey"
)

var fileURL = "ons/file.csv"
var apiURL = "api.example.com"

var ctx = context.Background()

func TestNew(t *testing.T) {
	Convey("Given a complete metadata struct", t, func() {
		m := &dataset.Metadata{
			Version: dataset.Version{
				ReleaseDate: "1 Jan 2000",
			},
			DatasetDetails: dataset.DatasetDetails{
				Title:            "title",
				Description:      "description",
				Theme:            "theme",
				License:          "license",
				ReleaseFrequency: "annual",
				Contacts: &[]dataset.Contact{
					dataset.Contact{
						Name:      "contact name",
						Telephone: "1234",
						Email:     "a@b.com",
					},
					dataset.Contact{
						Name:      "contact name2",
						Telephone: "5678",
						Email:     "y@z.com",
					}},
				Publisher: &dataset.Publisher{
					Name: "ONS",
					Type: "Gov Org",
					URL:  "ons.gov.uk",
				},
			},
		}

		Convey("When the New csvw function is called", func() {
			csvw := New(m, fileURL)

			Convey("Then the values should be set to the expected fields", func() {
				So(csvw.Context, ShouldEqual, "http://www.w3.org/ns/csvw")

				So(csvw.Title, ShouldEqual, m.Title)
				So(csvw.Description, ShouldEqual, m.Description)
				So(csvw.Issued, ShouldEqual, m.ReleaseDate)
				So(csvw.Theme, ShouldEqual, m.Theme)
				So(csvw.License, ShouldEqual, m.License)
				So(csvw.Frequency, ShouldEqual, m.ReleaseFrequency)
				So(csvw.URL, ShouldEqual, fileURL)

				mContacts := *m.Contacts
				So(csvw.Contact[0].Name, ShouldEqual, mContacts[0].Name)
				So(csvw.Contact[0].Telephone, ShouldEqual, mContacts[0].Telephone)
				So(csvw.Contact[0].Email, ShouldEqual, mContacts[0].Email)

				So(csvw.Publisher.Name, ShouldResemble, m.Publisher.Name)
				So(csvw.Publisher.Type, ShouldResemble, m.Publisher.Type)
				So(csvw.Publisher.ID, ShouldResemble, m.Publisher.URL)
			})
		})

		Convey("When the Publisher and Contact are empty", func() {
			m.Contacts = nil
			m.Publisher = nil

			Convey("And the New csvw function is called", func() {
				csvw := New(m, fileURL)

				Convey("Then the other values still are populated", func() {
					So(csvw.Context, ShouldEqual, "http://www.w3.org/ns/csvw")

					So(csvw.Title, ShouldEqual, m.Title)
					So(csvw.Description, ShouldEqual, m.Description)
					So(csvw.Issued, ShouldEqual, m.ReleaseDate)
					So(csvw.Theme, ShouldEqual, m.Theme)
					So(csvw.License, ShouldEqual, m.License)
					So(csvw.Frequency, ShouldEqual, m.ReleaseFrequency)
					So(csvw.URL, ShouldEqual, fileURL)

					So(csvw.Contact, ShouldBeEmpty)
					So(csvw.Publisher, ShouldBeZeroValue)
				})
			})
		})
	})
}

func TestAddNotes(t *testing.T) {
	Convey("Given metadata including alerts and usage notes", t, func() {
		n := &dataset.Metadata{
			Version: dataset.Version{
				Alerts: &[]dataset.Alert{
					{
						Date:        "1 Jan 2000",
						Description: "this is an alert",
						Type:        "correction",
					},
					{
						Date:        "31 Dec 2000",
						Description: "this alert came later",
						Type:        "correction",
					},
				},
			},
			DatasetDetails: dataset.DatasetDetails{
				UsageNotes: &[]dataset.UsageNote{
					dataset.UsageNote{
						Note:  "use it this way",
						Title: "first note",
					},
					dataset.UsageNote{
						Note:  "use it that way",
						Title: "second note",
					},
				},
			},
		}

		csvw := &CSVW{}

		Convey("When the AddNotes function is called", func() {
			csvw.AddNotes(n, fileURL)

			Convey("Then the values should be set to the expected fields", func() {
				So(csvw.Notes, ShouldHaveLength, 4)

				count := 0
				for _, a := range *n.Alerts {
					So(csvw.Notes[count].Type, ShouldEqual, a.Type)
					So(csvw.Notes[count].Body, ShouldEqual, a.Description)
					So(csvw.Notes[count].Target, ShouldEqual, fileURL)
					count++
				}

				for _, a := range *n.DatasetDetails.UsageNotes {
					So(csvw.Notes[count].Type, ShouldEqual, a.Title)
					So(csvw.Notes[count].Body, ShouldEqual, a.Note)
					So(csvw.Notes[count].Target, ShouldBeEmpty)
					count++
				}
			})
		})

		Convey("When there are no alerts and the AddNotes function is called", func() {
			n.Alerts = nil
			csvw.AddNotes(n, fileURL)

			Convey("Then the values should be set to the expected fields", func() {
				So(csvw.Notes, ShouldHaveLength, 2)

				for i, a := range *n.DatasetDetails.UsageNotes {
					So(csvw.Notes[i].Type, ShouldEqual, a.Title)
					So(csvw.Notes[i].Body, ShouldEqual, a.Note)
					So(csvw.Notes[i].Target, ShouldBeEmpty)
				}
			})
		})

		Convey("When there are no usage notes and the AddNotes function is called", func() {
			n.DatasetDetails.UsageNotes = nil
			csvw.AddNotes(n, fileURL)

			Convey("Then the values should be set to the expected fields", func() {
				So(csvw.Notes, ShouldHaveLength, 2)

				for i, a := range *n.Alerts {
					So(csvw.Notes[i].Type, ShouldEqual, a.Type)
					So(csvw.Notes[i].Body, ShouldEqual, a.Description)
					So(csvw.Notes[i].Target, ShouldEqual, fileURL)
				}
			})
		})

		Convey("When there are no usage notes or and the AddNotes function is called", func() {
			n.DatasetDetails.UsageNotes = nil
			n.Alerts = nil
			csvw.AddNotes(n, fileURL)

			Convey("Then the notes field should be empty", func() {
				So(csvw.Notes, ShouldBeEmpty)
			})
		})
	})
}

func TestFormatAboutURL(t *testing.T) {
	Convey("Given a valid domain config and url", t, func() {
		domain := "http://api.example.com/v1"
		url := "http://localhost:22000/datasets/1/editions/2/version/3/metadata"

		Convey("When the formatAboutURL function is called", func() {
			url, err := formatAboutURL(url, domain)

			Convey("Then the returned values should be as expected", func() {
				So(url, ShouldEqual, "http://api.example.com/v1/datasets/1/editions/2/version/3/metadata")
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestSplitHeader(t *testing.T) {
	Convey("Given a valid string version of a csv header", t, func() {
		h := "V4_2, a, b, c, d"

		Convey("When the splitHeader function is called", func() {
			split, offset, err := splitHeader(h)

			Convey("Then the returned values should be as expected", func() {
				So(split, ShouldHaveLength, 5)
				So(offset, ShouldEqual, 2)
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a valid csv row without a V4_X cell", t, func() {
		h := "bad, a, b, c, d"

		Convey("When the splitHeader function is called", func() {
			split, offset, err := splitHeader(h)

			Convey("Then the returned values should be as expected", func() {
				So(split, ShouldHaveLength, 0)
				So(offset, ShouldEqual, 0)
				So(err, ShouldEqual, errInvalidHeader)
			})
		})
	})

	Convey("Given a valid csv row with an invalid V4_X cell", t, func() {
		h := "V4_A, a, b, c, d"

		Convey("When the splitHeader function is called", func() {
			split, offset, err := splitHeader(h)

			Convey("Then the returned values should be as expected", func() {
				So(split, ShouldHaveLength, 0)
				So(offset, ShouldEqual, 0)
				So(err, ShouldEqual, errInvalidHeader)
			})
		})
	})
}

func TestGenerate(t *testing.T) {

	header := "V4_2, a, b, c, d"
	Convey("Given metadata that includes a dimension", t, func() {
		m := &dataset.Metadata{
			Version: dataset.Version{
				ReleaseDate: "1 Jan 2000",
				Dimensions: []dataset.VersionDimension{
					dataset.VersionDimension{
						Name: "geography",
						Links: dataset.Links{
							Self: dataset.Link{
								URL: "api/versions/self",
							},
						},
						Description: "areas included in dataset",
						Label:       "Geographic areas",
					},
				},
			},
			DatasetDetails: dataset.DatasetDetails{},
		}

		Convey("When the Generate csvw function is called", func() {
			data, err := Generate(ctx, m, header, fileURL, fileURL, apiURL)

			Convey("Then results should be returned with no errors", func() {
				So(data, ShouldHaveLength, 382)
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given metadata that does not include a dimension", t, func() {
		m := &dataset.Metadata{
			Version: dataset.Version{
				ReleaseDate: "1 Jan 2000",
			},
			DatasetDetails: dataset.DatasetDetails{},
		}

		Convey("When the Generate csvw function is called", func() {
			data, err := Generate(ctx, m, header, fileURL, fileURL, apiURL)

			Convey("Then results should be returned with no errors", func() {
				So(data, ShouldHaveLength, 0)
				So(err, ShouldEqual, errMissingDimensions)
			})
		})
	})
}
