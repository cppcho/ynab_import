package main

import (
	"testing"
)

func TestPayPay_Name(t *testing.T) {
	parser := PayPay{}
	if parser.Name() != "paypay" {
		t.Errorf("Name() = %q, want %q", parser.Name(), "paypay")
	}
}

func TestPayPay_Parse_ValidCSV(t *testing.T) {
	records, err := readCsvToRawRecords("testdata/parsers/paypay_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := PayPay{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Parse() returned nil for valid PayPay CSV")
	}

	// Should have 3 data rows (excluding header)
	if len(result.ValidRecords) != 3 {
		t.Errorf("Parse() returned %d records, want 3", len(result.ValidRecords))
	}

	// Verify first record (withdrawal: 出金金額（円）column has value, should be negative)
	if len(result.ValidRecords) > 0 {
		if result.ValidRecords[0].date != "2025-12-27" {
			t.Errorf("Record[0].date = %q, want %q", result.ValidRecords[0].date, "2025-12-27")
		}
		if result.ValidRecords[0].amount != "-1800" { // Withdrawal flipped to negative, comma removed
			t.Errorf("Record[0].amount = %q, want %q", result.ValidRecords[0].amount, "-1800")
		}
		if result.ValidRecords[0].payee != "テストストア" {
			t.Errorf("Record[0].payee = %q, want %q", result.ValidRecords[0].payee, "テストストア")
		}
	}

	// Verify second record (deposit: 入金金額（円）column has value, should be positive)
	if len(result.ValidRecords) > 1 {
		if result.ValidRecords[1].date != "2025-12-18" {
			t.Errorf("Record[1].date = %q, want %q", result.ValidRecords[1].date, "2025-12-18")
		}
		if result.ValidRecords[1].amount != "1,600" { // Deposit kept positive
			t.Errorf("Record[1].amount = %q, want %q", result.ValidRecords[1].amount, "1,600")
		}
		if result.ValidRecords[1].payee != "山田太郎" {
			t.Errorf("Record[1].payee = %q, want %q", result.ValidRecords[1].payee, "山田太郎")
		}
	}

	// Verify third record (withdrawal without commas)
	if len(result.ValidRecords) > 2 {
		if result.ValidRecords[2].amount != "-680" {
			t.Errorf("Record[2].amount = %q, want %q", result.ValidRecords[2].amount, "-680")
		}
	}
}

func TestPayPay_Parse_WrongHeaders(t *testing.T) {
	// Use a different parser's CSV
	records, err := readCsvToRawRecords("testdata/parsers/rakuten_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := PayPay{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}

	if result != nil {
		t.Error("Parse() should return nil for non-PayPay CSV")
	}
}

func TestPayPay_Parse_EmptyRecords(t *testing.T) {
	parser := PayPay{}
	result, err := parser.Parse([][]string{})

	if err != nil {
		t.Errorf("Parse() unexpected error for empty input: %v", err)
	}
	if result != nil {
		t.Error("Parse() should return nil for empty records")
	}
}

func TestPayPay_Parse_WithBOM(t *testing.T) {
	// Test CSV with UTF-8 BOM in first column (real-world PayPay export format)
	parser := PayPay{}

	mockRecords := [][]string{
		{"\ufeff取引日", "出金金額（円）", "入金金額（円）", "海外出金金額", "通貨", "変換レート（円）", "利用国", "取引内容", "取引先", "取引方法", "支払い区分", "利用者", "取引番号"},
		{"2025/1/5 12:00:00", "1000", "-", "-", "-", "-", "-", "支払い", "Test Store", "PayPay残高", "-", "-", "12345"},
	}

	result, err := parser.Parse(mockRecords)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Parse() returned nil for valid PayPay CSV with BOM")
	}

	if len(result.ValidRecords) != 1 {
		t.Errorf("Parse() returned %d records, want 1", len(result.ValidRecords))
	}

	if len(result.ValidRecords) > 0 {
		if result.ValidRecords[0].date != "2025-01-05" {
			t.Errorf("Date = %q, want %q", result.ValidRecords[0].date, "2025-01-05")
		}
		if result.ValidRecords[0].amount != "-1000" {
			t.Errorf("Amount = %q, want %q", result.ValidRecords[0].amount, "-1000")
		}
	}
}

func TestPayPay_Parse_InvalidDate(t *testing.T) {
	parser := PayPay{}

	mockRecords := [][]string{
		{"取引日", "出金金額（円）", "入金金額（円）", "海外出金金額", "通貨", "変換レート（円）", "利用国", "取引内容", "取引先", "取引方法", "支払い区分", "利用者", "取引番号"},
		{"2025/1/5 12:00:00", "1000", "-", "-", "-", "-", "-", "支払い", "Valid", "PayPay残高", "-", "-", "12345"},
		{"invalid-date 12:00:00", "2000", "-", "-", "-", "-", "-", "支払い", "Invalid", "PayPay残高", "-", "-", "12346"}, // Should skip
		{"2025/1/6 13:00:00", "3000", "-", "-", "-", "-", "-", "支払い", "Valid", "PayPay残高", "-", "-", "12347"},
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

func TestPayPay_Parse_DateTimeHandling(t *testing.T) {
	// PayPay uses "2006/1/2 15:04:05" format, should extract date part and convert to "2006-01-02"
	parser := PayPay{}

	mockRecords := [][]string{
		{"取引日", "出金金額（円）", "入金金額（円）", "海外出金金額", "通貨", "変換レート（円）", "利用国", "取引内容", "取引先", "取引方法", "支払い区分", "利用者", "取引番号"},
		{"2025/1/5 12:34:56", "1000", "-", "-", "-", "-", "-", "支払い", "Test", "PayPay残高", "-", "-", "12345"},
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

func TestPayPay_Parse_AmountHandling(t *testing.T) {
	parser := PayPay{}

	tests := []struct {
		name           string
		withdrawal     string // 出金金額（円）
		deposit        string // 入金金額（円）
		expectedAmount string
	}{
		{"deposit only", "-", "5000", "5000"},
		{"withdrawal only", "3000", "-", "-3000"},
		{"withdrawal with comma", "1,234", "-", "-1234"},
		{"deposit with comma", "-", "5,678", "5,678"},
		{"both (withdrawal preferred)", "1000", "2000", "-1000"}, // withdrawal takes precedence
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRecords := [][]string{
				{"取引日", "出金金額（円）", "入金金額（円）", "海外出金金額", "通貨", "変換レート（円）", "利用国", "取引内容", "取引先", "取引方法", "支払い区分", "利用者", "取引番号"},
				{"2025/1/1 12:00:00", tt.withdrawal, tt.deposit, "-", "-", "-", "-", "Test", "Test", "PayPay残高", "-", "-", "12345"},
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
