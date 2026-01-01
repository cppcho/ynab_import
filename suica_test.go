package main

import (
	"testing"
)

func TestSuica_Name(t *testing.T) {
	parser := Suica{}
	if parser.Name() != "suica" {
		t.Errorf("Name() = %q, want %q", parser.Name(), "suica")
	}
}

func TestSuica_Parse_ReturnsNil(t *testing.T) {
	// Suica parser only handles PDFs, not CSV
	parser := Suica{}
	result, err := parser.Parse([][]string{{"dummy", "data"}})

	if err != nil {
		t.Errorf("Parse() should not error on CSV input: %v", err)
	}

	if result != nil {
		t.Error("Parse() should return nil for CSV input (Suica uses PDF)")
	}
}

func TestSuica_ParsePDF_WrongFormat(t *testing.T) {
	// Create a temporary non-Suica PDF or use a different parser's test file
	parser := Suica{}

	// Using a CSV file which will fail PDF extraction
	result, err := parser.ParsePDF("testdata/parsers/smbc_valid.csv")

	if err == nil {
		t.Error("ParsePDF() should error on non-PDF file")
	}

	if result != nil {
		t.Error("ParsePDF() should return nil on error")
	}
}

func TestExtractYearFromFilename(t *testing.T) {
	tests := []struct {
		filename     string
		expectedYear int
	}{
		{"JE000000000000000_20251028_20260101110125.pdf", 2025},
		{"test_20230515_something.pdf", 2023},
		{"nodate.pdf", 0},
		{"invalid_12345678_test.pdf", 1234},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			year := extractYearFromFilename(tt.filename)
			if year != tt.expectedYear {
				t.Errorf("extractYearFromFilename(%q) = %d, want %d", tt.filename, year, tt.expectedYear)
			}
		})
	}
}
