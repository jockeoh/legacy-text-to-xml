package xmlwriter_test

import (
	"bytes"
	"encoding/xml"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jockeoh/legacy-text-to-xml/internal/model"
	"github.com/jockeoh/legacy-text-to-xml/internal/parser"
	"github.com/jockeoh/legacy-text-to-xml/internal/xmlwriter"
)

type testPeopleXML struct {
	People []testPersonXML `xml:"person"`
}

type testPersonXML struct {
	FirstName string          `xml:"firstname"`
	LastName  string          `xml:"lastname"`
	Address   *testAddressXML `xml:"address"`
	Phone     *testPhoneXML   `xml:"phone"`
	Family    []testFamilyXML `xml:"family"`
}

type testFamilyXML struct {
	Name    string          `xml:"name"`
	Born    string          `xml:"born"`
	Address *testAddressXML `xml:"address"`
	Phone   *testPhoneXML   `xml:"phone"`
}

type testAddressXML struct {
	Street string  `xml:"street"`
	City   string  `xml:"city"`
	Zip    *string `xml:"zip"`
}

type testPhoneXML struct {
	Mobile   string `xml:"mobile"`
	Landline string `xml:"landline"`
}

func TestWriteSample(t *testing.T) {
	input, err := os.Open("../../testdata/sample-input.txt")
	if err != nil {
		t.Fatalf("open sample input: %v", err)
	}
	t.Cleanup(func() {
		if err := input.Close(); err != nil {
			t.Errorf("close sample input: %v", err)
		}
	})

	people, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("parse sample: %v", err)
	}
	var output bytes.Buffer
	if err := xmlwriter.Write(&output, people); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	want, err := os.ReadFile("../../testdata/sample-expected.xml")
	if err != nil {
		t.Fatalf("read expected XML: %v", err)
	}
	if output.String() != string(want) {
		t.Errorf("sample XML differs\ngot:\n%s\nwant:\n%s", output.String(), want)
	}

	document := decode(t, output.Bytes())
	if len(document.People) != 2 || len(document.People[0].Family) != 2 {
		t.Fatalf("decoded sample = %#v, want two people and two family members", document)
	}
	if document.People[0].Family[0].Address == nil || document.People[0].Family[1].Phone == nil {
		t.Fatalf("family ownership lost in XML: %#v", document.People[0].Family)
	}
	if document.People[1].Address == nil || document.People[1].Address.Zip != nil {
		t.Fatalf("Boris address = %#v, want omitted zip", document.People[1].Address)
	}
}

func TestWriteEscapesSpecialCharacters(t *testing.T) {
	people, err := parser.Parse(strings.NewReader("P|Anna & Eva|<Test>\nA|A > B & C|<City>|\n"))
	if err != nil {
		t.Fatalf("parse special characters: %v", err)
	}
	var output bytes.Buffer
	if err := xmlwriter.Write(&output, people); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	xmlText := output.String()
	if !strings.Contains(xmlText, "Anna &amp; Eva") || !strings.Contains(xmlText, "&lt;Test&gt;") {
		t.Errorf("XML does not contain escaped values: %s", xmlText)
	}
	document := decode(t, output.Bytes())
	person := document.People[0]
	if person.FirstName != "Anna & Eva" || person.LastName != "<Test>" {
		t.Errorf("decoded names = %q, %q, want source values", person.FirstName, person.LastName)
	}
	if person.Address == nil || person.Address.Street != "A > B & C" || person.Address.City != "<City>" {
		t.Errorf("decoded address = %#v, want source values", person.Address)
	}
}

func TestWriteUsesCanonicalOrder(t *testing.T) {
	people, err := parser.Parse(strings.NewReader("P|Ada|Lovelace\nT|mobile|landline\nA|street|city|zip\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	var output bytes.Buffer
	if err := xmlwriter.Write(&output, people); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	xmlText := output.String()
	addressAt := strings.Index(xmlText, "<address>")
	phoneAt := strings.Index(xmlText, "<phone>")
	if addressAt < 0 || phoneAt < 0 || addressAt > phoneAt {
		t.Errorf("address must precede phone regardless of input order: %s", xmlText)
	}
}

func TestWritePreservesFamilyOrder(t *testing.T) {
	people := []model.Person{{
		FirstName: "Parent",
		LastName:  "Person",
		Family: []model.FamilyMember{
			{Name: "First", Born: "2000"},
			{Name: "Second", Born: "2001"},
		},
	}}
	var output bytes.Buffer
	if err := xmlwriter.Write(&output, people); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	document := decode(t, output.Bytes())
	family := document.People[0].Family
	if len(family) != 2 || family[0].Name != "First" || family[1].Name != "Second" {
		t.Errorf("family = %#v, want input order", family)
	}
}

func TestWriteDistinguishesMissingAndEmptyZip(t *testing.T) {
	empty := ""
	people := []model.Person{
		{FirstName: "No", LastName: "Zip", Address: &model.Address{Street: "one", City: "city"}},
		{FirstName: "Empty", LastName: "Zip", Address: &model.Address{Street: "two", City: "city", Zip: &empty}},
	}
	var output bytes.Buffer
	if err := xmlwriter.Write(&output, people); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	document := decode(t, output.Bytes())
	if document.People[0].Address.Zip != nil {
		t.Errorf("missing zip = %#v, want nil", document.People[0].Address.Zip)
	}
	if document.People[1].Address.Zip == nil || *document.People[1].Address.Zip != "" {
		t.Errorf("empty zip = %#v, want pointer to empty string", document.People[1].Address.Zip)
	}
}

func TestWriteReturnsWriterError(t *testing.T) {
	wantErr := errors.New("write failed")
	err := xmlwriter.Write(errorWriter{err: wantErr}, []model.Person{{FirstName: "Ada", LastName: "Lovelace"}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("errors.Is(%v) = false; error = %v", wantErr, err)
	}
}

func decode(t *testing.T, data []byte) testPeopleXML {
	t.Helper()
	var document testPeopleXML
	if err := xml.Unmarshal(data, &document); err != nil {
		t.Fatalf("xml.Unmarshal() error = %v\nXML:\n%s", err, data)
	}
	return document
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}
