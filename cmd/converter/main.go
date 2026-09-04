package main

import (
	"fmt"
	"os"

	"github.com/jockeoh/legacy-text-to-xml/internal/model"
	"github.com/jockeoh/legacy-text-to-xml/internal/parser"
	"github.com/jockeoh/legacy-text-to-xml/internal/xmlwriter"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: converter <input-file> <output-file>")
		os.Exit(2)
	}

	if err := convert(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintf(os.Stderr, "converter: %v\n", err)
		os.Exit(1)
	}
}

func convert(inputPath, outputPath string) error {
	people, err := parseFile(inputPath)
	if err != nil {
		return fmt.Errorf("parse %q: %w", inputPath, err)
	}
	if err := writeFile(outputPath, people); err != nil {
		return fmt.Errorf("write %q: %w", outputPath, err)
	}
	return nil
}

func parseFile(path string) (_ []model.Person, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open input: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close input: %w", closeErr)
		}
	}()

	return parser.Parse(file)
}

func writeFile(path string, people []model.Person) (err error) {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close output: %w", closeErr)
		}
	}()

	return xmlwriter.Write(file, people)
}
