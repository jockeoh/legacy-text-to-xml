package parser_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jockeoh/legacy-text-to-xml/internal/model"
	"github.com/jockeoh/legacy-text-to-xml/internal/parser"
)

func TestParseSample(t *testing.T) {
	file, err := os.Open("../../testdata/sample-input.txt")
	if err != nil {
		t.Fatalf("open sample input: %v", err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close sample input: %v", err)
		}
	})

	people, err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(people) != 2 {
		t.Fatalf("len(people) = %d, want 2", len(people))
	}

	elof := people[0]
	if elof.FirstName != "Elof" || elof.LastName != "Sundin" {
		t.Errorf("first person = %#v, want Elof Sundin", elof)
	}
	if elof.Address == nil || elof.Address.Street != "S:t Johannesgatan 16" || value(elof.Address.Zip) != "75330" {
		t.Errorf("Elof address = %#v, want sample address", elof.Address)
	}
	if elof.Phone == nil || elof.Phone.Mobile != "073-101801" {
		t.Errorf("Elof phone = %#v, want sample phone", elof.Phone)
	}
	if len(elof.Family) != 2 {
		t.Fatalf("len(Elof.Family) = %d, want 2", len(elof.Family))
	}
	if elof.Family[0].Name != "Hans" || elof.Family[0].Address == nil || elof.Family[0].Phone != nil {
		t.Errorf("first family member = %#v, want Hans with address", elof.Family[0])
	}
	if elof.Family[1].Name != "Anna" || elof.Family[1].Phone == nil || elof.Family[1].Address != nil {
		t.Errorf("second family member = %#v, want Anna with phone", elof.Family[1])
	}

	boris := people[1]
	if boris.FirstName != "Boris" || boris.Address == nil {
		t.Fatalf("second person = %#v, want Boris with address", boris)
	}
	if boris.Address.Street != "10 Downing Street" || boris.Address.City != "London" || boris.Address.Zip != nil {
		t.Errorf("Boris address = %#v, want address without zip", boris.Address)
	}
}

func TestParseValidRecords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, []model.Person)
	}{
		{
			name:  "multiple people",
			input: "P|Ada|Lovelace\nP|Alan|Turing\n",
			check: func(t *testing.T, people []model.Person) {
				if len(people) != 2 || people[1].FirstName != "Alan" {
					t.Fatalf("people = %#v, want Ada then Alan", people)
				}
			},
		},
		{
			name:  "person with only name",
			input: "P|Grace|Hopper\n",
			check: func(t *testing.T, people []model.Person) {
				person := onlyPerson(t, people)
				if person.Address != nil || person.Phone != nil || len(person.Family) != 0 {
					t.Fatalf("person = %#v, want only name", person)
				}
			},
		},
		{
			name:  "person with phone",
			input: "P|Grace|Hopper\nT|mobile|landline\n",
			check: func(t *testing.T, people []model.Person) {
				person := onlyPerson(t, people)
				if person.Phone == nil || person.Phone.Mobile != "mobile" || person.Phone.Landline != "landline" {
					t.Fatalf("phone = %#v, want preserved fields", person.Phone)
				}
			},
		},
		{
			name:  "person with address",
			input: "P|Grace|Hopper\nA|street|city|zip\n",
			check: func(t *testing.T, people []model.Person) {
				person := onlyPerson(t, people)
				if person.Address == nil || value(person.Address.Zip) != "zip" {
					t.Fatalf("address = %#v, want address with zip", person.Address)
				}
			},
		},
		{
			name:  "family with phone",
			input: "P|Grace|Hopper\nF|Child|2000\nT|mobile|landline\n",
			check: func(t *testing.T, people []model.Person) {
				person := onlyPerson(t, people)
				if person.Phone != nil || len(person.Family) != 1 || person.Family[0].Phone == nil {
					t.Fatalf("person = %#v, want phone on family", person)
				}
			},
		},
		{
			name:  "family with address",
			input: "P|Grace|Hopper\nF|Child|2000\nA|street|city|zip\n",
			check: func(t *testing.T, people []model.Person) {
				person := onlyPerson(t, people)
				if person.Address != nil || len(person.Family) != 1 || person.Family[0].Address == nil {
					t.Fatalf("person = %#v, want address on family", person)
				}
			},
		},
		{
			name:  "family order and current family replacement",
			input: "P|Grace|Hopper\nF|First|2000\nA|one|city\nF|Second|2001\nT|two|phone\n",
			check: func(t *testing.T, people []model.Person) {
				family := onlyPerson(t, people).Family
				if len(family) != 2 || family[0].Name != "First" || family[1].Name != "Second" {
					t.Fatalf("family = %#v, want input order", family)
				}
				if family[0].Address == nil || family[0].Phone != nil || family[1].Address != nil || family[1].Phone == nil {
					t.Fatalf("family ownership incorrect: %#v", family)
				}
			},
		},
		{
			name:  "blank lines ignored",
			input: " \t\n\nP|Grace|Hopper\n\r\nT|mobile|landline\n",
			check: func(t *testing.T, people []model.Person) {
				if onlyPerson(t, people).Phone == nil {
					t.Fatal("phone missing after blank lines")
				}
			},
		},
		{
			name:  "payload whitespace and empty fields preserved",
			input: "P| Grace |\nT|| landline \n",
			check: func(t *testing.T, people []model.Person) {
				person := onlyPerson(t, people)
				if person.FirstName != " Grace " || person.LastName != "" {
					t.Fatalf("name fields = %q, %q, want whitespace and empty value", person.FirstName, person.LastName)
				}
				if person.Phone == nil || person.Phone.Mobile != "" || person.Phone.Landline != " landline " {
					t.Fatalf("phone = %#v, want exact payload", person.Phone)
				}
			},
		},
		{
			name:  "missing and explicit empty zip differ",
			input: "P|No|Zip\nA|street|city\nP|Empty|Zip\nA|street|city|\n",
			check: func(t *testing.T, people []model.Person) {
				if people[0].Address.Zip != nil {
					t.Fatalf("missing zip = %#v, want nil", people[0].Address.Zip)
				}
				if people[1].Address.Zip == nil || *people[1].Address.Zip != "" {
					t.Fatalf("explicit empty zip = %#v, want pointer to empty string", people[1].Address.Zip)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			people, err := parser.Parse(strings.NewReader(test.input))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			test.check(t, people)
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		category error
		line     int
		record   string
	}{
		{name: "phone before person", input: "T|mobile|landline\n", category: parser.ErrInvalidHierarchy, line: 1, record: "T|mobile|landline"},
		{name: "address before person", input: "A|street|city|zip\n", category: parser.ErrInvalidHierarchy, line: 1, record: "A|street|city|zip"},
		{name: "family before person", input: "F|Child|2000\n", category: parser.ErrInvalidHierarchy, line: 1, record: "F|Child|2000"},
		{name: "unknown record", input: "P|Ada|Lovelace\nX|value\n", category: parser.ErrUnknownRecord, line: 2, record: "X|value"},
		{name: "person field count", input: "P|Ada\n", category: parser.ErrInvalidFieldCount, line: 1, record: "P|Ada"},
		{name: "phone field count", input: "P|Ada|Lovelace\nT|one|two|three\n", category: parser.ErrInvalidFieldCount, line: 2, record: "T|one|two|three"},
		{name: "family field count", input: "P|Ada|Lovelace\nF|Child\n", category: parser.ErrInvalidFieldCount, line: 2, record: "F|Child"},
		{name: "address too few fields", input: "P|Ada|Lovelace\nA|street\n", category: parser.ErrInvalidFieldCount, line: 2, record: "A|street"},
		{name: "address too many fields", input: "P|Ada|Lovelace\nA|street|city|zip|extra\n", category: parser.ErrInvalidFieldCount, line: 2, record: "A|street|city|zip|extra"},
		{name: "duplicate person phone", input: "P|Ada|Lovelace\nT|one|one\nT|two|two\n", category: parser.ErrDuplicateRecord, line: 3, record: "T|two|two"},
		{name: "duplicate person address", input: "P|Ada|Lovelace\nA|one|one\nA|two|two\n", category: parser.ErrDuplicateRecord, line: 3, record: "A|two|two"},
		{name: "duplicate family phone", input: "P|Ada|Lovelace\nF|Child|2000\nT|one|one\nT|two|two\n", category: parser.ErrDuplicateRecord, line: 4, record: "T|two|two"},
		{name: "duplicate family address", input: "P|Ada|Lovelace\nF|Child|2000\nA|one|one\nA|two|two\n", category: parser.ErrDuplicateRecord, line: 4, record: "A|two|two"},
		{name: "blank lines count", input: "\n  \nP|Ada\n", category: parser.ErrInvalidFieldCount, line: 3, record: "P|Ada"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parser.Parse(strings.NewReader(test.input))
			if err == nil {
				t.Fatal("Parse() error = nil, want error")
			}
			if !errors.Is(err, test.category) {
				t.Errorf("errors.Is(%v) = false, want true; error = %v", test.category, err)
			}
			var parseErr *parser.ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("errors.As(*ParseError) = false; error = %v", err)
			}
			if parseErr.Line != test.line {
				t.Errorf("ParseError.Line = %d, want %d", parseErr.Line, test.line)
			}
			if parseErr.Record != test.record {
				t.Errorf("ParseError.Record = %q, want %q", parseErr.Record, test.record)
			}
		})
	}
}

func TestParseAcceptsLineAboveScannerDefaultLimit(t *testing.T) {
	firstName := strings.Repeat("a", 70*1024)
	people, err := parser.Parse(strings.NewReader("P|" + firstName + "|Last\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := onlyPerson(t, people).FirstName; got != firstName {
		t.Fatalf("len(firstname) = %d, want %d", len(got), len(firstName))
	}
}

func TestParseReportsLineForRecordAboveConfiguredLimit(t *testing.T) {
	oversized := strings.Repeat("a", 1024*1024)
	_, err := parser.Parse(strings.NewReader("P|" + oversized + "|Last\n"))
	if err == nil {
		t.Fatal("Parse() error = nil, want scanner size error")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("error = %q, want line 1", err)
	}
}

func onlyPerson(t *testing.T, people []model.Person) model.Person {
	t.Helper()
	if len(people) != 1 {
		t.Fatalf("len(people) = %d, want 1", len(people))
	}
	return people[0]
}

func value(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
