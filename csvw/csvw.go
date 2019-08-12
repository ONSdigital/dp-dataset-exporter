package csvw

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/log"
)

//CSVW provides a structure for describing a CSV through a JSON metadata file.
//The JSON tags feature web vocabularies like Dublin Core, DCAT and Stat-DCAT
//to help further contextualize and define the metadata being provided.
//The URL field in the CSVW must reference a CSV file, and all other data
//should describe that CSVs contents.
type CSVW struct {
	Context     string    `json:"@context"`
	URL         string    `json:"url"`
	Title       string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Issued      string    `json:"datePublished,omitempty"`
	Publisher   Publisher `json:"creator"`
	Contact     []Contact `json:"contactPoint"`
	Spacial     string    `json:"spacial"`
	Temporal    string    `json:"temporal"`
	TableSchema Columns   `json:"tableSchema"`
	Theme       string    `json:"includedInDataCatalog,omitempty"`
	License     string    `json:"license,omitempty"`
	Frequency   string    `json:"dct:accrualPeriodicity,omitempty"`
	Notes       []Note    `json:"notes,omitempty"`
}

// Contact represents a response model within a dataset
type Contact struct {
	Name      string `json:"vcard:fn"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}

//Publisher defines the entity primarily responsible for the dataset
//https://www.w3.org/TR/vocab-dcat/#class-catalog
type Publisher struct {
	Name string `json:"name,omitempty"`
	Type string `json:"@type,omitempty"`
	ID   string `json:"@id"` //a URL where more info is available
}

//Columns provides the nested structure expected within the tableSchema of a CSVW
type Columns struct {
	C              []Column `json:"columns"`
	About          string   `json:"aboutUrl"`
	dimensionNames []string
}

//Column provides the ability to define the JSON tags required specific
//to each column within the CSVW
type Column map[string]interface{}

//Note can include alerts, corrections or usage notes which exist in the
//dataset metadata and help describe the contents of the CSV
type Note struct {
	Type       string `json:"type"` // is this an enum?
	Target     string `json:"target,omitempty"`
	Body       string `json:"body"`
	Motivation string `json:"motivation,omitempty"` // how is this different from type? do we need this? is this an enum?
}

var errInvalidHeader = errors.New("invalid header row - no V4_X cell")
var errMissingDimensions = errors.New("no dimensions in provided metadata")

//New CSVW returned with top level fields populated based on provided metadata
func New(m *dataset.Metadata, csvURL string) *CSVW {
	csvw := &CSVW{
		Context:     "http://www.w3.org/ns/csvw",
		Title:       m.Title,
		Description: m.Description,
		Issued:      m.ReleaseDate,
		Theme:       m.Theme,
		License:     "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3",
		Frequency:   m.ReleaseFrequency,
		URL:         csvURL,
	}

	for _, c := range m.Contacts {
		csvw.Contact = append(csvw.Contact, Contact{
			Name:      c.Name,
			Telephone: c.Telephone,
			Email:     c.Email,
		})
	}

	if m.Publisher != nil {
		csvw.Publisher = Publisher{
			Name: m.Publisher.Name,
			Type: "https://schema.org/GovernmentOrganization",
			ID:   m.Publisher.URL,
		}
	}

	return csvw
}

//Generate the CSVW structured metadata file to describe a CSV
func Generate(metadata *dataset.Metadata, header, downloadURL, aboutURL, apiDomain string) ([]byte, error) {
	if len(metadata.Dimensions) == 0 {
		return nil, errMissingDimensions
	}

	h, offset, err := splitHeader(header)
	if err != nil {
		return nil, err
	}
	log.Info("header split for CSVW generation", log.Data{"header": header, "column_offset": strconv.Itoa(offset)})

	csvw := New(metadata, downloadURL)

	obs := csvw.newObservationColumn(h[0], metadata.UnitOfMeasure)
	log.Info("added observation column to CSVW", log.Data{"column": obs})

	//add data markings columns
	if offset != 0 {
		for i := 1; i <= offset; i++ {
			csvw.addNewColumn(h[i], "", false)
			log.Info("added observation metadata column to CSVW", log.Data{"csv": downloadURL})
		}
	}

	offset++
	h = h[offset:]

	//add dimension columns
	for i := 0; i < len(h); i = i + 2 {
		csvw.newCodeAndLabelColumns(i, apiDomain, h, metadata.Dimensions)
		log.Info("added pair of dimension columns to CSVW", log.Data{"csv": downloadURL})
	}

	csvw.TableSchema.About, err = formatAboutURL(apiDomain, aboutURL, csvw.TableSchema.dimensionNames)
	if err != nil {
		return nil, err
	}

	log.Info("all columns added to CSVW", log.Data{"number_of_columns": strconv.Itoa(len(csvw.TableSchema.C))})
	csvw.AddNotes(metadata, downloadURL)

	b, err := json.Marshal(csvw)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func formatAboutURL(domain, aboutURL string, queryParams []string) (string, error) {
	if len(queryParams) == 0 {
		return "", errors.New("cannot build aboutURL without query parameter keys")
	}
	about, err := url.Parse(aboutURL)
	if err != nil {
		return "", err
	}

	d, err := url.Parse(domain)
	if err != nil {
		return "", err
	}

	d.Path = d.Path + about.Path
	q := "?"
	for i := 0; i < len(queryParams); i = i + 2 {
		q += queryParams[i] + "={" + queryParams[i+1] + "}"
	}

	return d.String() + q, nil
}

//AddNotes to CSVW from alerts or usage notes in provided metadata
func (csvw *CSVW) AddNotes(metadata *dataset.Metadata, url string) {
	if metadata.Alerts != nil {
		for _, a := range *metadata.Alerts {
			csvw.Notes = append(csvw.Notes, Note{
				Type:   a.Type,
				Body:   a.Description,
				Target: url,
			})
		}
	}

	if metadata.UsageNotes != nil {
		for _, u := range *metadata.UsageNotes {
			csvw.Notes = append(csvw.Notes, Note{
				Type: u.Title,
				Body: u.Note,
				// TODO: store column number against usage notes in Dataset API
				//Target: url + "#col=need-to-store",
			})
		}
	}
}

func splitHeader(header string) ([]string, int, error) {
	h := strings.Split(header, ",")

	parts := strings.Split(h[0], "_")
	if len(parts) != 2 {
		return nil, 0, errInvalidHeader
	}

	offset, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, errInvalidHeader
	}

	return h, offset, err
}

func (csvw *CSVW) newObservationColumn(title, unitOfMeasure string) Column {
	c := newColumn(title, "", true)

	c["datatype"] = "string"
	c["variable_measured"] = unitOfMeasure

	log.Info("adding observations column", log.Data{"column": c})
	csvw.TableSchema.C = append(csvw.TableSchema.C, c)
	return c
}

func (csvw *CSVW) newCodeAndLabelColumns(i int, apiDomain string, header []string, dims []dataset.Dimension) {
	codeHeader := header[i]
	dimHeader := header[i+1]
	dimHeader = strings.ToLower(dimHeader)

	var dim dataset.Dimension

	for _, d := range dims {
		if d.Name == dimHeader {
			dim = d
			break
		}
	}

	dimURL := dim.URL
	if len(apiDomain) > 0 {
		uri, err := url.Parse(dim.URL)
		if err != nil {
			return
		}

		//TODO: this should not be needed once the dataset API dimensions refer to code list editions
		if !strings.Contains(uri.Path, "/editions/") {
			uri.Path += "/editions/one-off"
		}

		dimURL = fmt.Sprintf("%s%s", apiDomain, uri.Path)
	}

	if dimHeader == "time" {
		csvw.Temporal = dimURL
	}

	if dimHeader == "geography" {
		csvw.Spacial = dimURL
	}

	codeCol := newColumn("", codeHeader, true)
	codeCol["valueURL"] = dimURL + "/codes/{" + codeHeader + "}"

	labelCol := newColumn(dim.Name, dim.Label, false)
	labelCol["description"] = dim.Description

	csvw.TableSchema.C = append(csvw.TableSchema.C, codeCol, labelCol)
	csvw.TableSchema.dimensionNames = append(csvw.TableSchema.dimensionNames, labelCol["titles"].(string), codeCol["name"].(string))
}

func (csvw *CSVW) addNewColumn(title, name string, required bool) {
	csvw.TableSchema.C = append(csvw.TableSchema.C, newColumn(title, name, required))
}

func newColumn(title, name string, required bool) Column {
	c := make(Column)
	if len(title) > 0 {
		c["titles"] = title
	}

	if len(name) > 0 {
		c["name"] = name
	}

	if required {
		c["required"] = true
	}

	return c
}
