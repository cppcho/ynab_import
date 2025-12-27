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
	result, err := parser.Parse(records)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Parse() returned nil for valid SMBC CSV")
	}

	// Should have 3 data rows (excluding header)
	if len(result.ValidRecords) != 3 {
		t.Errorf("Parse() returned %d records, want 3", len(result.ValidRecords))
	}

	// Verify first record (deposit: お預入れ column has value)
	if len(result.ValidRecords) > 0 {
		if result.ValidRecords[0].date != "2025-12-26" {
			t.Errorf("Record[0].date = %q, want %q", result.ValidRecords[0].date, "2025-12-26")
		}
		if result.ValidRecords[0].amount != "31113" { // お預入れ value, not flipped
			t.Errorf("Record[0].amount = %q, want %q", result.ValidRecords[0].amount, "31113")
		}
	}

	// Verify second record (withdrawal: お引出し column has value, should be flipped)
	if len(result.ValidRecords) > 1 {
		// Original value is 23000 in お引出し, should be flipped to -23000
		if result.ValidRecords[1].amount != "-23000" {
			t.Errorf("Record[1].amount = %q, want %q (flipSign applied)", result.ValidRecords[1].amount, "-23000")
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
	result, err := parser.Parse(records)
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}

	if result != nil {
		t.Error("Parse() should return nil for non-SMBC CSV")
	}
}

func TestSmbc_Parse_EmptyRecords(t *testing.T) {
	parser := Smbc{}
	result, err := parser.Parse([][]string{})

	if err != nil {
		t.Errorf("Parse() unexpected error for empty input: %v", err)
	}
	if result != nil {
		t.Error("Parse() should return nil for empty records")
	}
}

func TestSmbc_Parse_InvalidDate(t *testing.T) {
	parser := Smbc{}

	mockRecords := [][]string{
		{"年月日", "お引出し", "お預入れ", "お取り扱い内容", "残高", "メモ", "ラベル"},
		{"2025/1/5", "", "1000", "Valid", "10000", "", ""},
		{"invalid-date", "", "2000", "Invalid", "12000", "", ""}, // Should skip
		{"2025/1/6", "", "3000", "Valid", "15000", "", ""},
	}

	result, err := parser.Parse(mockRecords)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	// Should have 2 valid records (1 skipped)
	if len(result.ValidRecords) != 2 {
		t.Errorf("Parse() returned %d valid records, want 2", len(result.ValidRecords))
	}

	// Should have 1 skipped row
	if len(result.SkippedRows) != 1 {
		t.Errorf("Parse() returned %d skipped rows, want 1", len(result.SkippedRows))
	}

	// Verify skipped row details
	if len(result.SkippedRows) > 0 {
		if result.SkippedRows[0].RowNumber != 3 {
			t.Errorf("SkippedRow[0].RowNumber = %d, want 3", result.SkippedRows[0].RowNumber)
		}
	}
}

func TestSmbc_Parse_DateConversion(t *testing.T) {
	// SMBC uses "2006/1/2" format, should convert to "2006-01-02"
	parser := Smbc{}

	mockRecords := [][]string{
		{"年月日", "お引出し", "お預入れ", "お取り扱い内容", "残高", "メモ", "ラベル"},
		{"2025/1/5", "", "1000", "Test", "10000", "", ""},
	}

	result, err := parser.Parse(mockRecords)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Parse() returned nil")
	}

	if len(result.ValidRecords) > 0 {
		if result.ValidRecords[0].date != "2025-01-05" {
			t.Errorf("Date conversion failed: got %q, want %q", result.ValidRecords[0].date, "2025-01-05")
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

			result, err := parser.Parse(mockRecords)
			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}
			if result == nil || len(result.ValidRecords) == 0 {
				t.Fatal("Parse() returned nil or empty")
			}

			if result.ValidRecords[0].amount != tt.expectedAmount {
				t.Errorf("Amount = %q, want %q", result.ValidRecords[0].amount, tt.expectedAmount)
			}
		})
	}
}
