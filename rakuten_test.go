package main

import (
	"testing"
)

func TestRakuten_Name(t *testing.T) {
	parser := Rakuten{}
	if parser.Name() != "rakuten" {
		t.Errorf("Name() = %q, want %q", parser.Name(), "rakuten")
	}
}

func TestRakuten_Parse_ValidCSV(t *testing.T) {
	records, err := readCsvToRawRecords("testdata/parsers/rakuten_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Rakuten{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Parse() returned nil for valid Rakuten CSV")
	}

	// Should have 3 data rows
	if len(result) != 3 {
		t.Errorf("Parse() returned %d records, want 3", len(result))
	}

	// Verify first record
	if len(result) > 0 {
		if result[0].date != "2025-11-27" {
			t.Errorf("Record[0].date = %q, want %q", result[0].date, "2025-11-27")
		}
		if result[0].amount != "-100000" {
			t.Errorf("Record[0].amount = %q, want %q", result[0].amount, "-100000")
		}
	}
}

func TestRakuten_Parse_WrongHeaders(t *testing.T) {
	// Use SMBC CSV (different headers)
	records, err := readCsvToRawRecords("testdata/parsers/smbc_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Rakuten{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}

	if result != nil {
		t.Error("Parse() should return nil for non-Rakuten CSV")
	}
}

func TestRakuten_Parse_DateConversion(t *testing.T) {
	// Rakuten uses "20060102" compact format, should convert to "2006-01-02"
	parser := Rakuten{}

	mockRecords := [][]string{
		{"取引日", "入出金(円)", "取引後残高(円)", "入出金内容"},
		{"20250105", "-1000", "50000", "Test Payment"},
	}

	result, err := parser.Parse(mockRecords)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Parse() returned nil")
	}

	if len(result) > 0 {
		if result[0].date != "2025-01-05" {
			t.Errorf("Date conversion failed: got %q, want %q", result[0].date, "2025-01-05")
		}
	}
}

func TestRakuten_Parse_AmountHandling(t *testing.T) {
	parser := Rakuten{}

	tests := []struct {
		name           string
		amount         string
		expectedAmount string
	}{
		{"negative amount", "-5000", "-5000"},
		{"positive amount", "3000", "3000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRecords := [][]string{
				{"取引日", "入出金(円)", "取引後残高(円)", "入出金内容"},
				{"20250101", tt.amount, "100000", "Test"},
			}

			result, err := parser.Parse(mockRecords)
			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}
			if len(result) == 0 {
				t.Fatal("Parse() returned nil or empty")
			}

			if result[0].amount != tt.expectedAmount {
				t.Errorf("Amount = %q, want %q", result[0].amount, tt.expectedAmount)
			}
		})
	}
}
