package main

import (
	"testing"
)

func TestEpos_Name(t *testing.T) {
	parser := Epos{}
	if parser.Name() != "epos" {
		t.Errorf("Name() = %q, want %q", parser.Name(), "epos")
	}
}

func TestEpos_Parse_ValidCSV(t *testing.T) {
	records, err := readCsvToRawRecords("testdata/parsers/epos_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Epos{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Parse() returned nil for valid Epos CSV")
	}

	// Should have 3 data rows
	if len(result) != 3 {
		t.Errorf("Parse() returned %d records, want 3", len(result))
	}

	// Verify first record
	if len(result) > 0 {
		if result[0].date != "2025-12-24" {
			t.Errorf("Record[0].date = %q, want %q", result[0].date, "2025-12-24")
		}
		// Amount should be flipped (negative)
		if result[0].amount != "-1155" {
			t.Errorf("Record[0].amount = %q, want %q (flipSign applied)", result[0].amount, "-1155")
		}
	}
}

func TestEpos_Parse_WrongHeaders(t *testing.T) {
	// Use SMBC CSV (different headers)
	records, err := readCsvToRawRecords("testdata/parsers/smbc_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	parser := Epos{}
	result, err := parser.Parse(records)
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}

	if result != nil {
		t.Error("Parse() should return nil for non-Epos CSV")
	}
}

func TestEpos_Parse_DateConversion(t *testing.T) {
	// Epos uses "2006年01月02日" Japanese format
	parser := Epos{}

	mockRecords := [][]string{
		{"種別（ショッピング、キャッシング、その他）", "ご利用年月日", "ご利用場所", "ご利用内容", "ご利用金額", "お支払金額（キャッシングでは利息を含みます）", "支払区分"},
		{"ショッピング", "2025年01月05日", "Test Shop", "−", "1000", "1000", "1回払い"},
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

func TestEpos_Parse_EmptyRowSkipping(t *testing.T) {
	// Epos skips rows where row[1] == "" or row[6] == ""
	parser := Epos{}

	mockRecords := [][]string{
		{"種別（ショッピング、キャッシング、その他）", "ご利用年月日", "ご利用場所", "ご利用内容", "ご利用金額", "お支払金額（キャッシングでは利息を含みます）", "支払区分"},
		{"ショッピング", "2025年01月05日", "Shop 1", "−", "1000", "1000", "1回払い"}, // Valid
		{"ショッピング", "", "Shop 2", "−", "2000", "2000", "1回払い"},            // Empty row[1] - should skip
		{"ショッピング", "2025年01月06日", "Shop 3", "−", "3000", "3000", ""},     // Empty row[6] - should skip
		{"ショッピング", "2025年01月07日", "Shop 4", "−", "4000", "4000", "1回払い"}, // Valid
	}

	result, err := parser.Parse(mockRecords)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Parse() returned nil")
	}

	// Should have only 2 valid records (skipped 2)
	if len(result) != 2 {
		t.Errorf("Parse() returned %d records, want 2 (2 rows should be skipped)", len(result))
	}
}

func TestEpos_Parse_FullWidthHyphenSkipping(t *testing.T) {
	// Epos should skip rows with full-width hyphen (－) in date field
	// These appear in special rows like annual fees ("その他" category)
	parser := Epos{}

	mockRecords := [][]string{
		{"種別（ショッピング、キャッシング、その他）", "ご利用年月日", "ご利用場所", "ご利用内容", "ご利用金額", "お支払金額（キャッシングでは利息を含みます）", "支払区分"},
		{"ショッピング", "2025年01月05日", "Shop 1", "−", "1000", "1000", "1回払い"}, // Valid
		{"ショッピング", "2025年01月06日", "Shop 2", "−", "2000", "2000", "1回払い"}, // Valid
		{"その他", "－", "－", "年会費", "－", "20000", "－"},                      // Full-width hyphen in date - should skip
		{"ショッピング", "2025年01月07日", "Shop 3", "−", "3000", "3000", "1回払い"}, // Valid
	}

	result, err := parser.Parse(mockRecords)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Parse() returned nil")
	}

	// Should have only 3 valid records (skipped the annual fee row)
	if len(result) != 3 {
		t.Errorf("Parse() returned %d records, want 3 (annual fee row should be skipped)", len(result))
	}

	// Verify the valid records are correct
	if len(result) == 3 {
		if result[0].payee != "Shop 1" {
			t.Errorf("Record[0].payee = %q, want %q", result[0].payee, "Shop 1")
		}
		if result[1].payee != "Shop 2" {
			t.Errorf("Record[1].payee = %q, want %q", result[1].payee, "Shop 2")
		}
		if result[2].payee != "Shop 3" {
			t.Errorf("Record[2].payee = %q, want %q", result[2].payee, "Shop 3")
		}
	}
}
