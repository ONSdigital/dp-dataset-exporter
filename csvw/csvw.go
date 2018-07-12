package csvw

import (
	"encoding/json"
	"errors"
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
	Context     string          `json:"@context"`
	URL         string          `json:"url"`
	Title       string          `json:"dct:title"`
	Description string          `json:"dct:description,omitempty"`
	Issued      string          `json:"dct:issued,omitempty"`
	Publisher   Publisher       `json:"dct:publisher"`
	Contact     dataset.Contact `json:"dcat:contactPoint,omitempty"` //// TODO: handle array as per spec
	TableSchema Columns         `json:"tableSchema"`
	Theme       string          `json:"dcat:theme,omitempty"`
	License     string          `json:"dct:license,omitempty"`
	Frequency   string          `json:"dct:accrualPeriodicity,omitempty"`
	Notes       []Note          `json:"notes,omitempty"`
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
	C     []Column `json:"columns"`
	About string   `json:"aboutUrl"`
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

//New CSVW returned with top level fields populated based on provided metadata
func New(m *dataset.Metadata, csvURL string) *CSVW {
	csvw := &CSVW{
		Context:     "http://www.w3.org/ns/csvw",
		Title:       m.Title,
		Description: m.Description,
		Issued:      m.ReleaseDate,
		Theme:       m.Theme,
		License:     m.License,
		Frequency:   m.ReleaseFrequency,
		URL:         csvURL,
	}

	//// TODO: handle multiple contacts
	if len(m.Contacts) != 0 {
		csvw.Contact = m.Contacts[0]
	}

	if m.Publisher != nil {
		csvw.Publisher = Publisher{
			Name: m.Publisher.Name,
			Type: m.Publisher.Type,
			ID:   m.Publisher.URL,
		}
	}

	return csvw
}

//Generate the CSVW structured metadata file to describe a CSV
func Generate(metadata *dataset.Metadata, header, downloadURL, aboutURL string) ([]byte, error) {
	if len(metadata.Dimensions) == 0 {
		return nil, errors.New("no dimensions in provided metadata")
	}

	h, offset, err := splitHeader(header)
	if err != nil {
		return nil, err
	}
	log.Info("header split for CSVW generation", log.Data{"header": header, "column_offset": strconv.Itoa(offset)})

	csvw := New(metadata, downloadURL)

	var list []Column
	obs := newObservationColumn(downloadURL, h[0], metadata.UnitOfMeasure)
	list = append(list, obs)
	log.Info("added observation column to CSVW", log.Data{"column": obs})

	//add data markings columns
	if offset != 0 {
		for i := 1; i <= offset; i++ {
			c := newColumn(i, downloadURL, h[i], "")
			list = append(list, c)
			log.Info("added data marking column to CSVW", log.Data{"column": c})
		}
	}

	offset++
	h = h[offset:]

	//add dimensions columns
	for i := 0; i < len(h); i = i + 2 {
		c, l := newCodeAndLabelColumns(offset, i, downloadURL, h, metadata.Dimensions)
		log.Info("added dimension columns to CSVW", log.Data{"code_column": c, "label_column": l})
		list = append(list, c, l)
	}

	csvw.TableSchema = Columns{
		About: aboutURL,
		C:     list,
	}

	log.Info("all columns added to CSVW", log.Data{"number_of_columns": strconv.Itoa(len(list))})
	csvw.AddNotes(metadata, downloadURL)

	b, err := json.Marshal(csvw)
	if err != nil {
		return nil, err
	}

	return b, nil
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
		return nil, 0, errors.New("invalid header row - no V4_X cell")
	}

	offset, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, errors.New("invalid header row - no V4_X cell")
	}

	return h, offset, err
}

func newObservationColumn(url, title, name string) Column {
	c := newColumn(0, url, title, name)

	c["datatype"] = "number"
	c["required"] = true

	log.Info("adding observations column", log.Data{"column": c})
	return c
}

func newCodeAndLabelColumns(offset, i int, downloadURL string, header []string, dims []dataset.Dimension) (Column, Column) {
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

	codeCol := newColumn(offset+i, downloadURL, "", codeHeader)
	codeCol["valueURL"] = dim.HRef + "/codes/{" + codeHeader + "}" //// TODO: could this link to the code list or API?
	codeCol["required"] = true
	// TODO: determine what could go in c["datatype"]

	labelCol := newColumn(offset+i+1, downloadURL, dim.Label, dimHeader)
	labelCol["description"] = dim.Description
	// TODO: determine what could go in c["datatype"] and c["required"]

	return codeCol, labelCol
}

func newColumn(n int, url, title, name string) Column {
	c := Column{
		"@id": url + "#col=" + strconv.Itoa(n),
	}

	if len(title) > 0 {
		c["titles"] = title
	}

	if len(name) > 0 {
		c["name"] = name
	}
	return c
}