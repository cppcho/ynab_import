package main

import (
	"testing"
)

func TestSmbc_Name(t *testing.T) {
	parser := Smbc{}
	if parser.Name() != "smbc" {
		t.Errorf("Name() = %q, want %q", parser.Name(), "smbc")
	}
}

func TestSmbc_Parse_ValidCSV(t *testing.T) {
	records, err := readCsvToRawRecords("testdata/parsers/smbc_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Smbc{}
	result := parser.Parse(records)

	if result == nil {
		t.Fatal("Parse() returned nil for valid SMBC CSV")
	}

	// Should have 3 data rows (excluding header)
	if len(result) != 3 {
		t.Errorf("Parse() returned %d records, want 3", len(result))
	}

	// Verify first record (deposit: お預入れ column has value)
	if len(result) > 0 {
		if result[0].date != "2025-12-26" {
			t.Errorf("Record[0].date = %q, want %q", result[0].date, "2025-12-26")
		}
		if result[0].amount != "31113" { // お預入れ value, not flipped
			t.Errorf("Record[0].amount = %q, want %q", result[0].amount, "31113")
		}
	}

	// Verify second record (withdrawal: お引出し column has value, should be flipped)
	if len(result) > 1 {
		// Original value is 23000 in お引出し, should be flipped to -23000
		if result[1].amount != "-23000" {
			t.Errorf("Record[1].amount = %q, want %q (flipSign applied)", result[1].amount, "-23000")
		}
	}
}

func TestSmbc_Parse_WrongHeaders(t *testing.T) {
	// Use a different parser's CSV
	records, err := readCsvToRawRecords("testdata/parsers/rakuten_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Smbc{}
	result := parser.Parse(records)

	if result != nil {
		t.Error("Parse() should return nil for non-SMBC CSV")
	}
}

func TestSmbc_Parse_EmptyRecords(t *testing.T) {
	// This test documents that empty input causes panic (current behavior)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Parse() should panic for empty records, but didn't")
		}
	}()

	parser := Smbc{}
	parser.Parse([][]string{}) // Will panic
}

func TestSmbc_Parse_DateConversion(t *testing.T) {
	// SMBC uses "2006/1/2" format, should convert to "2006-01-02"
	parser := Smbc{}

	mockRecords := [][]string{
		{"年月日", "お引出し", "お預入れ", "お取り扱い内容", "残高", "メモ", "ラベル"},
		{"2025/1/5", "", "1000", "Test", "10000", "", ""},
	}

	result := parser.Parse(mockRecords)
	if result == nil {
		t.Fatal("Parse() returned nil")
	}

	if len(result) > 0 {
		if result[0].date != "2025-01-05" {
			t.Errorf("Date conversion failed: got %q, want %q", result[0].date, "2025-01-05")
		}
	}
}

func TestSmbc_Parse_AmountHandling(t *testing.T) {
	parser := Smbc{}

	tests := []struct {
		name           string
		withdrawal     string // お引出し
		deposit        string // お預入れ
		expectedAmount string
	}{
		{"deposit only", "", "5000", "5000"},
		{"withdrawal only", "3000", "", "-3000"},
		{"both (withdrawal preferred)", "1000", "2000", "-1000"}, // withdrawal takes precedence
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRecords := [][]string{
				{"年月日", "お引出し", "お預入れ", "お取り扱い内容", "残高", "メモ", "ラベル"},
				{"2025/1/1", tt.withdrawal, tt.deposit, "Test", "10000", "", ""},
			}

			result := parser.Parse(mockRecords)
			if result == nil || len(result) == 0 {
				t.Fatal("Parse() returned nil or empty")
			}

			if result[0].amount != tt.expectedAmount {
				t.Errorf("Amount = %q, want %q", result[0].amount, tt.expectedAmount)
			}
		})
	}
}
