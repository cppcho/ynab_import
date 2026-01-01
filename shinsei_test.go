package main

import (
	"testing"
)

func TestShinsei_Name(t *testing.T) {
	parser := Shinsei{}
	if parser.Name() != "shinsei" {
		t.Errorf("Name() = %q, want %q", parser.Name(), "shinsei")
	}
}

func TestShinsei_Parse_ValidCSV(t *testing.T) {
	records, err := readCsvToRawRecords("testdata/parsers/shinsei_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Shinsei{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Parse() returned nil for valid Shinsei CSV")
	}

	// Should have 10 data rows (from the sample CSV)
	if len(result.ValidRecords) != 10 {
		t.Errorf("Parse() returned %d records, want 10", len(result.ValidRecords))
	}

	// Verify first record (withdrawal - 地方税)
	if len(result.ValidRecords) > 0 {
		if result.ValidRecords[0].date != "2026-01-01" {
			t.Errorf("Record[0].date = %q, want %q", result.ValidRecords[0].date, "2026-01-01")
		}
		// Withdrawal amount should be flipped
		if result.ValidRecords[0].amount != "-3" {
			t.Errorf("Record[0].amount = %q, want %q (flipSign applied)", result.ValidRecords[0].amount, "-3")
		}
		if result.ValidRecords[0].memo != "地方税" {
			t.Errorf("Record[0].memo = %q, want %q", result.ValidRecords[0].memo, "地方税")
		}
	}

	// Verify third record (deposit - 税引前利息)
	if len(result.ValidRecords) > 2 {
		if result.ValidRecords[2].date != "2026-01-01" {
			t.Errorf("Record[2].date = %q, want %q", result.ValidRecords[2].date, "2026-01-01")
		}
		// Deposit amount should not be flipped
		if result.ValidRecords[2].amount != "76" {
			t.Errorf("Record[2].amount = %q, want %q", result.ValidRecords[2].amount, "76")
		}
		if result.ValidRecords[2].memo != "税引前利息" {
			t.Errorf("Record[2].memo = %q, want %q", result.ValidRecords[2].memo, "税引前利息")
		}
	}
}

func TestShinsei_Parse_WrongHeaders(t *testing.T) {
	// Use SMBC CSV (different headers)
	records, err := readCsvToRawRecords("testdata/parsers/smbc_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Shinsei{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}

	if result != nil {
		t.Error("Parse() should return nil for non-Shinsei CSV")
	}
}

func TestShinsei_Parse_DateConversion(t *testing.T) {
	// Shinsei uses "2006/01/02" format, should convert to "2006-01-02"
	parser := Shinsei{}

	mockRecords := [][]string{
		{"取引日", "摘要", "出金金額", "入金金額", "残高"},
		{"2025/01/05", "Test", "", "1000", "100000"},
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

func TestShinsei_Parse_AmountHandling(t *testing.T) {
	parser := Shinsei{}

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
				{"取引日", "摘要", "出金金額", "入金金額", "残高"},
				{"2025/01/01", "Test", tt.withdrawal, tt.deposit, "100000"},
			}

			result, err := parser.Parse(mockRecords)
			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}
			if len(result.ValidRecords) == 0 {
				t.Fatal("Parse() returned nil or empty")
			}

			if result.ValidRecords[0].amount != tt.expectedAmount {
				t.Errorf("Amount = %q, want %q", result.ValidRecords[0].amount, tt.expectedAmount)
			}
		})
	}
}
