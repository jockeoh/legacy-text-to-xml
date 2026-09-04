// Package xmlwriter serializes the internal model to the required XML format.
package xmlwriter

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/jockeoh/legacy-text-to-xml/internal/model"
)

type peopleXML struct {
	XMLName xml.Name    `xml:"people"`
	People  []personXML `xml:"person"`
}

type personXML struct {
	FirstName string      `xml:"firstname"`
	LastName  string      `xml:"lastname"`
	Address   *addressXML `xml:"address,omitempty"`
	Phone     *phoneXML   `xml:"phone,omitempty"`
	Family    []familyXML `xml:"family"`
}

type familyXML struct {
	Name    string      `xml:"name"`
	Born    string      `xml:"born"`
	Address *addressXML `xml:"address,omitempty"`
	Phone   *phoneXML   `xml:"phone,omitempty"`
}

type addressXML struct {
	Street string  `xml:"street"`
	City   string  `xml:"city"`
	Zip    *string `xml:"zip,omitempty"`
}

type phoneXML struct {
	Mobile   string `xml:"mobile"`
	Landline string `xml:"landline"`
}

// Write writes indented UTF-8 XML to w. The order of fields in the XML structs
// defines the canonical element order.
func Write(w io.Writer, people []model.Person) error {
	document := peopleXML{People: make([]personXML, 0, len(people))}
	for _, person := range people {
		document.People = append(document.People, toPersonXML(person))
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "    ")
	if err := encoder.Encode(document); err != nil {
		return fmt.Errorf("encode XML: %w", err)
	}
	if _, err := io.WriteString(w, "\n"); err != nil {
		return fmt.Errorf("write XML newline: %w", err)
	}
	return nil
}

func toPersonXML(person model.Person) personXML {
	result := personXML{
		FirstName: person.FirstName,
		LastName:  person.LastName,
		Address:   toAddressXML(person.Address),
		Phone:     toPhoneXML(person.Phone),
		Family:    make([]familyXML, 0, len(person.Family)),
	}
	for _, family := range person.Family {
		result.Family = append(result.Family, familyXML{
			Name:    family.Name,
			Born:    family.Born,
			Address: toAddressXML(family.Address),
			Phone:   toPhoneXML(family.Phone),
		})
	}
	return result
}

func toAddressXML(address *model.Address) *addressXML {
	if address == nil {
		return nil
	}
	return &addressXML{
		Street: address.Street,
		City:   address.City,
		Zip:    address.Zip,
	}
}

func toPhoneXML(phone *model.Phone) *phoneXML {
	if phone == nil {
		return nil
	}
	return &phoneXML{Mobile: phone.Mobile, Landline: phone.Landline}
}
