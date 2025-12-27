package main

import (
	"testing"
)

func TestSmbcCard_Name(t *testing.T) {
	parser := SmbcCard{}
	if parser.Name() != "smbc_card" {
		t.Errorf("Name() = %q, want %q", parser.Name(), "smbc_card")
	}
}

func TestSmbcCard_Parse_ValidCSV(t *testing.T) {
	records, err := readCsvToRawRecords("testdata/parsers/smbc_card_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := SmbcCard{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Parse() returned nil for valid SMBC Card CSV")
	}

	// Should have 3 data rows
	if len(result.ValidRecords) != 3 {
		t.Errorf("Parse() returned %d records, want 3", len(result.ValidRecords))
	}

	// Verify first record
	if len(result.ValidRecords) > 0 {
		if result.ValidRecords[0].date != "2025-12-23" {
			t.Errorf("Record[0].date = %q, want %q", result.ValidRecords[0].date, "2025-12-23")
		}
		// Amount should be flipped (negative)
		if result.ValidRecords[0].amount != "-2230" {
			t.Errorf("Record[0].amount = %q, want %q (flipSign applied)", result.ValidRecords[0].amount, "-2230")
		}
	}
}

func TestSmbcCard_Parse_WrongHeaders(t *testing.T) {
	// Use SMBC CSV (different format)
	records, err := readCsvToRawRecords("testdata/parsers/smbc_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := SmbcCard{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}

	if result != nil {
		t.Error("Parse() should return nil for non-SMBC Card CSV")
	}
}

func TestSmbcCard_Parse_DateConversion(t *testing.T) {
	// SMBC Card uses "2006/1/2" format
	parser := SmbcCard{}

	mockRecords := [][]string{
		{"2025/1/5", "Test Shop", "ご本人", "1回払い", "", "'26/01", "1000", "1000", "", "", "", "", ""},
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

func TestSmbcCard_Parse_Validation(t *testing.T) {
	parser := SmbcCard{}

	tests := []struct {
		name       string
		col2       string // records[0][2]
		col5       string // records[0][5]
		shouldPass bool
	}{
		{"ご本人 with quote prefix", "ご本人", "'26/01", true},
		{"ご家族 with quote prefix", "ご家族", "'26/01", true},
		{"invalid col2", "その他", "'26/01", false},
		{"no quote prefix", "ご本人", "26/01", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRecords := [][]string{
				{"2025/1/1", "Test", tt.col2, "1回払い", "", tt.col5, "1000", "1000", "", "", "", "", ""},
			}

			result, err := parser.Parse(mockRecords)
			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}
			if tt.shouldPass && result == nil {
				t.Error("Parse() returned nil, expected valid result")
			}
			if !tt.shouldPass && result != nil {
				t.Error("Parse() returned result, expected nil")
			}
		})
	}
}

func TestSmbcCard_Parse_EmptyColumn7Fallback(t *testing.T) {
	parser := SmbcCard{}

	tests := []struct {
		name           string
		col6           string // Amount in column 6
		col7           string // Amount in column 7 (may be empty)
		expectedAmount string
	}{
		{
			name:           "normal transaction with col7",
			col6:           "2230",
			col7:           "2230",
			expectedAmount: "-2230", // flipSign applied
		},
		{
			name:           "international transaction - empty col7",
			col6:           "3198",
			col7:           "",
			expectedAmount: "-3198", // Should fall back to col6
		},
		{
			name:           "both columns have values",
			col6:           "1000",
			col7:           "1000",
			expectedAmount: "-1000", // Prefers col7
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRecords := [][]string{
				{"2025/12/5", "Test Merchant", "ご本人", "1回払い", "", "'26/01", tt.col6, tt.col7, "", "", "", "", ""},
			}

			result, err := parser.Parse(mockRecords)
			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("Parse() returned nil for valid record")
			}

			if len(result.ValidRecords) != 1 {
				t.Fatalf("Parse() returned %d records, want 1", len(result.ValidRecords))
			}

			if result.ValidRecords[0].amount != tt.expectedAmount {
				t.Errorf("amount = %q, want %q", result.ValidRecords[0].amount, tt.expectedAmount)
			}
		})
	}
}
