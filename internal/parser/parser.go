// Package parser reads the line-based legacy format into the internal model.
package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jockeoh/legacy-text-to-xml/internal/model"
)

const maxLineSize = 1024 * 1024

var (
	ErrInvalidHierarchy  = errors.New("invalid record hierarchy")
	ErrUnknownRecord     = errors.New("unknown record type")
	ErrInvalidFieldCount = errors.New("invalid field count")
	ErrDuplicateRecord   = errors.New("duplicate record")
)

// ParseError identifies the input line and record that could not be parsed.
// Err retains the error category so callers can inspect it with errors.Is.
type ParseError struct {
	Line   int
	Record string
	Err    error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d, record %q: %v", e.Line, e.Record, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// Parse reads legacy records from r. Blank lines are ignored but still count
// towards line numbers in errors.
func Parse(r io.Reader) ([]model.Person, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)

	var people []model.Person
	currentPerson := -1
	currentFamily := -1
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "|")
		recordType := parts[0]

		switch recordType {
		case "P":
			if err := requireFieldCount(lineNumber, line, parts, 3); err != nil {
				return nil, err
			}
			people = append(people, model.Person{
				FirstName: parts[1],
				LastName:  parts[2],
			})
			currentPerson = len(people) - 1
			currentFamily = -1

		case "F":
			if err := requirePerson(lineNumber, line, currentPerson); err != nil {
				return nil, err
			}
			if err := requireFieldCount(lineNumber, line, parts, 3); err != nil {
				return nil, err
			}
			person := &people[currentPerson]
			person.Family = append(person.Family, model.FamilyMember{
				Name: parts[1],
				Born: parts[2],
			})
			currentFamily = len(person.Family) - 1

		case "T":
			if err := requirePerson(lineNumber, line, currentPerson); err != nil {
				return nil, err
			}
			if err := requireFieldCount(lineNumber, line, parts, 3); err != nil {
				return nil, err
			}
			phone := &model.Phone{Mobile: parts[1], Landline: parts[2]}
			person := &people[currentPerson]
			if currentFamily >= 0 {
				family := &person.Family[currentFamily]
				if family.Phone != nil {
					return nil, parseError(lineNumber, line, ErrDuplicateRecord, "family member already has a phone")
				}
				family.Phone = phone
				continue
			}
			if person.Phone != nil {
				return nil, parseError(lineNumber, line, ErrDuplicateRecord, "person already has a phone")
			}
			person.Phone = phone

		case "A":
			if err := requirePerson(lineNumber, line, currentPerson); err != nil {
				return nil, err
			}
			if len(parts) != 3 && len(parts) != 4 {
				return nil, parseError(lineNumber, line, ErrInvalidFieldCount, fmt.Sprintf("A expects 3 or 4 fields, got %d", len(parts)))
			}
			address := &model.Address{Street: parts[1], City: parts[2]}
			if len(parts) == 4 {
				zip := parts[3]
				address.Zip = &zip
			}
			person := &people[currentPerson]
			if currentFamily >= 0 {
				family := &person.Family[currentFamily]
				if family.Address != nil {
					return nil, parseError(lineNumber, line, ErrDuplicateRecord, "family member already has an address")
				}
				family.Address = address
				continue
			}
			if person.Address != nil {
				return nil, parseError(lineNumber, line, ErrDuplicateRecord, "person already has an address")
			}
			person.Address = address

		default:
			return nil, parseError(lineNumber, line, ErrUnknownRecord, fmt.Sprintf("record type %q", recordType))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read input at line %d: %w", lineNumber+1, err)
	}
	return people, nil
}

func requirePerson(lineNumber int, record string, currentPerson int) error {
	if currentPerson < 0 {
		return parseError(lineNumber, record, ErrInvalidHierarchy, "record requires a preceding P record")
	}
	return nil
}

func requireFieldCount(lineNumber int, record string, parts []string, want int) error {
	if len(parts) != want {
		return parseError(lineNumber, record, ErrInvalidFieldCount, fmt.Sprintf("%s expects %d fields, got %d", parts[0], want, len(parts)))
	}
	return nil
}

func parseError(lineNumber int, record string, category error, detail string) error {
	return &ParseError{
		Line:   lineNumber,
		Record: record,
		Err:    fmt.Errorf("%w: %s", category, detail),
	}
}
