// Package model contains the source-independent representation used between
// the legacy parser and the XML writer.
package model

type Person struct {
	FirstName string
	LastName  string
	Address   *Address
	Phone     *Phone
	Family    []FamilyMember
}

type FamilyMember struct {
	Name    string
	Born    string
	Address *Address
	Phone   *Phone
}

type Address struct {
	Street string
	City   string
	Zip    *string
}

type Phone struct {
	Mobile   string
	Landline string
}
