package main

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertSampleFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "output.xml")
	if err := convert("../../testdata/sample-input.txt", outputPath); err != nil {
		t.Fatalf("convert() error = %v", err)
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var document struct {
		People []struct{} `xml:"person"`
	}
	if err := xml.Unmarshal(output, &document); err != nil {
		t.Fatalf("xml.Unmarshal(output) error = %v\nXML:\n%s", err, output)
	}
	if len(document.People) != 2 {
		t.Fatalf("len(people) = %d, want 2", len(document.People))
	}
}
