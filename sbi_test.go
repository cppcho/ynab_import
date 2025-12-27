package main

import (
	"testing"
)

func TestSbi_Name(t *testing.T) {
	parser := Sbi{}
	if parser.Name() != "sbi" {
		t.Errorf("Name() = %q, want %q", parser.Name(), "sbi")
	}
}

func TestSbi_Parse_ValidCSV(t *testing.T) {
	records, err := readCsvToRawRecords("testdata/parsers/sbi_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Sbi{}
	result := parser.Parse(records)

	if result == nil {
		t.Fatal("Parse() returned nil for valid SBI CSV")
	}

	// Should have 3 data rows
	if len(result) != 3 {
		t.Errorf("Parse() returned %d records, want 3", len(result))
	}

	// Verify first record (withdrawal)
	if len(result) > 0 {
		if result[0].date != "2025-12-26" {
			t.Errorf("Record[0].date = %q, want %q", result[0].date, "2025-12-26")
		}
		// Withdrawal amount should be flipped
		if result[0].amount != "-91688" {
			t.Errorf("Record[0].amount = %q, want %q (flipSign applied)", result[0].amount, "-91688")
		}
	}

	// Verify second record (deposit)
	if len(result) > 1 {
		// Deposit amount should not be flipped
		if result[1].amount != "50000" {
			t.Errorf("Record[1].amount = %q, want %q", result[1].amount, "50000")
		}
	}
}

func TestSbi_Parse_WrongHeaders(t *testing.T) {
	// Use SMBC CSV (different headers)
	records, err := readCsvToRawRecords("testdata/parsers/smbc_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Sbi{}
	result := parser.Parse(records)

	if result != nil {
		t.Error("Parse() should return nil for non-SBI CSV")
	}
}

func TestSbi_Parse_DateConversion(t *testing.T) {
	// SBI uses "2006/01/02" format, should convert to "2006-01-02"
	parser := Sbi{}

	mockRecords := [][]string{
		{"日付", "内容", "出金金額(円)", "入金金額(円)", "残高(円)", "メモ"},
		{"2025/01/05", "Test", "", "1000", "100000", "-"},
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

func TestSbi_Parse_AmountHandling(t *testing.T) {
	parser := Sbi{}

	tests := []struct {
		name           string
		withdrawal     string // 出金金額
		deposit        string // 入金金額
		expectedAmount string
	}{
		{"deposit only", "", "5000", "5000"},
		{"withdrawal only", "3000", "", "-3000"},
		{"both (withdrawal preferred)", "1000", "2000", "-1000"}, // withdrawal takes precedence
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRecords := [][]string{
				{"日付", "内容", "出金金額(円)", "入金金額(円)", "残高(円)", "メモ"},
				{"2025/01/01", "Test", tt.withdrawal, tt.deposit, "100000", "-"},
			}

			result := parser.Parse(mockRecords)
			if len(result) == 0 {
				t.Fatal("Parse() returned nil or empty")
			}

			if result[0].amount != tt.expectedAmount {
				t.Errorf("Amount = %q, want %q", result[0].amount, tt.expectedAmount)
			}
		})
	}
}
