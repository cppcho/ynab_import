package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCsvToRawRecords_UTF8(t *testing.T) {
	records, err := readCsvToRawRecords("testdata/csv/utf8_simple.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	if len(records) != 4 {
		t.Errorf("readCsvToRawRecords() got %d records, want 4", len(records))
	}

	// Check header
	expectedHeader := []string{"日付", "金額", "内容"}
	if len(records) > 0 {
		for i, h := range expectedHeader {
			if records[0][i] != h {
				t.Errorf("Header[%d] = %q, want %q", i, records[0][i], h)
			}
		}
	}
}

func TestReadCsvToRawRecords_ShiftJIS(t *testing.T) {
	// Note: This test uses one of the real Shift_JIS samples for accurate testing
	// The small synthetic file may not have enough data for reliable encoding detection
	records, err := readCsvToRawRecords("testdata/parsers/epos_valid.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	if len(records) < 1 {
		t.Fatal("readCsvToRawRecords() returned no records")
	}

	// Verify Shift_JIS was properly decoded to UTF-8
	// If encoding detection failed, Japanese characters would be garbled
	// Epos header starts with: 種別（ショッピング、キャッシング、その他）
	if len(records[0]) > 0 && !containsJapanese(records[0][0]) {
		t.Errorf("Shift_JIS decoding may have failed: got %q", records[0][0])
	}
}

// Helper function to check if string contains Japanese characters
func containsJapanese(s string) bool {
	for _, r := range s {
		if (r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF) || // Katakana
			(r >= 0x4E00 && r <= 0x9FAF) { // Kanji
			return true
		}
	}
	return false
}

func TestReadCsvToRawRecords_Empty(t *testing.T) {
	records, err := readCsvToRawRecords("testdata/csv/empty.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	if len(records) != 0 {
		t.Errorf("readCsvToRawRecords() got %d records, want 0", len(records))
	}
}

func TestReadCsvToRawRecords_Malformed(t *testing.T) {
	// With LazyQuotes = true, malformed CSVs should still be read
	records, err := readCsvToRawRecords("testdata/csv/malformed.csv")
	if err != nil {
		t.Fatalf("readCsvToRawRecords() unexpected error = %v", err)
	}

	// Should have at least the header row
	if len(records) < 1 {
		t.Errorf("readCsvToRawRecords() got %d records, want at least 1", len(records))
	}
}

func TestReadCsvToRawRecords_FileNotFound(t *testing.T) {
	_, err := readCsvToRawRecords("testdata/csv/does_not_exist.csv")
	if err == nil {
		t.Error("readCsvToRawRecords() expected error for non-existent file, got nil")
	}
}

func TestWriteRecordsToCsv_Basic(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.csv")

	records := []YnabRecord{
		{date: "2024-01-15", payee: "Store A", memo: "Purchase", amount: "1000"},
		{date: "2024-01-16", payee: "Store B", memo: "Payment", amount: "-500"},
	}

	err := writeRecordsToCsv(records, outputPath)
	if err != nil {
		t.Fatalf("writeRecordsToCsv() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("writeRecordsToCsv() did not create file at %s", outputPath)
	}

	// Read back and verify
	readRecords, err := readCsvToRawRecords(outputPath)
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	if len(readRecords) != 3 { // header + 2 data rows
		t.Errorf("got %d rows, want 3", len(readRecords))
	}

	// Verify header
	expectedHeader := []string{"Date", "Payee", "Memo", "Amount"}
	for i, h := range expectedHeader {
		if readRecords[0][i] != h {
			t.Errorf("Header[%d] = %q, want %q", i, readRecords[0][i], h)
		}
	}
}

func TestWriteRecordsToCsv_DoubleFlipSignBug(t *testing.T) {
	// BUG DOCUMENTATION: This test documents the double flipSign bug in csv.go:32
	// The code calls flipSign(flipSign(record.amount)), which is a no-op
	// Expected behavior: amount should be written as-is (no flipping)
	// Current behavior: amount is flipped twice, resulting in original value (accidentally correct)

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.csv")

	records := []YnabRecord{
		{date: "2024-01-15", payee: "Test", memo: "Memo", amount: "1000"},
		{date: "2024-01-16", payee: "Test", memo: "Memo", amount: "-500"},
	}

	err := writeRecordsToCsv(records, outputPath)
	if err != nil {
		t.Fatalf("writeRecordsToCsv() error = %v", err)
	}

	// Read back
	readRecords, err := readCsvToRawRecords(outputPath)
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	// BUG: Double flipSign is a no-op, so amounts should equal original values
	// Row 1 (index 1, after header): amount should be "1000" (not "-1000")
	if readRecords[1][3] != "1000" {
		t.Errorf("Double flipSign bug: got %q, want %q (proving no-op)", readRecords[1][3], "1000")
	}

	// Row 2 (index 2): amount should be "-500" (not "500")
	if readRecords[2][3] != "-500" {
		t.Errorf("Double flipSign bug: got %q, want %q (proving no-op)", readRecords[2][3], "-500")
	}
}

func TestWriteRecordsToCsv_SkipsEmptyDateOrAmount(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.csv")

	records := []YnabRecord{
		{date: "2024-01-15", payee: "Valid", memo: "Valid", amount: "1000"},
		{date: "", payee: "No Date", memo: "Should Skip", amount: "500"},          // Empty date - should skip
		{date: "2024-01-16", payee: "No Amount", memo: "Should Skip", amount: ""}, // Empty amount - should skip
		{date: "2024-01-17", payee: "Valid", memo: "Valid", amount: "2000"},
	}

	err := writeRecordsToCsv(records, outputPath)
	if err != nil {
		t.Fatalf("writeRecordsToCsv() error = %v", err)
	}

	// Read back
	readRecords, err := readCsvToRawRecords(outputPath)
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	// Should have header + 2 valid rows (skipped 2 invalid)
	if len(readRecords) != 3 {
		t.Errorf("got %d rows, want 3 (header + 2 valid rows)", len(readRecords))
	}
}

func TestWriteRecordsToCsv_EmptyRecordsList(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.csv")

	records := []YnabRecord{}

	err := writeRecordsToCsv(records, outputPath)
	if err != nil {
		t.Fatalf("writeRecordsToCsv() error = %v", err)
	}

	// Read back
	readRecords, err := readCsvToRawRecords(outputPath)
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	// Should have only header
	if len(readRecords) != 1 {
		t.Errorf("got %d rows, want 1 (header only)", len(readRecords))
	}
}

func TestWriteRecordsToCsv_SpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.csv")

	records := []YnabRecord{
		{date: "2024-01-15", payee: "Store \"Quotes\"", memo: "Comma, in memo", amount: "1000"},
		{date: "2024-01-16", payee: "New\nLine", memo: "Tab\there", amount: "2000"},
	}

	err := writeRecordsToCsv(records, outputPath)
	if err != nil {
		t.Fatalf("writeRecordsToCsv() error = %v", err)
	}

	// Read back
	readRecords, err := readCsvToRawRecords(outputPath)
	if err != nil {
		t.Fatalf("readCsvToRawRecords() error = %v", err)
	}

	// Verify special characters are preserved
	if readRecords[1][1] != "Store \"Quotes\"" {
		t.Errorf("Quotes not preserved: got %q", readRecords[1][1])
	}
	if readRecords[1][2] != "Comma, in memo" {
		t.Errorf("Comma not preserved: got %q", readRecords[1][2])
	}
}

func TestWriteRecordsToCsv_InvalidPath(t *testing.T) {
	// Try to write to a directory that doesn't exist (and we can't create)
	invalidPath := "/nonexistent/directory/that/should/not/exist/output.csv"

	records := []YnabRecord{
		{date: "2024-01-15", payee: "Test", memo: "Test", amount: "1000"},
	}

	err := writeRecordsToCsv(records, invalidPath)
	if err == nil {
		t.Error("writeRecordsToCsv() expected error for invalid path, got nil")
	}
}

func TestPrintCsv(t *testing.T) {
	// Note: printCsv writes to stdout, which is hard to test without capturing output
	// This is a basic test to ensure it doesn't panic
	records := [][]string{
		{"Header1", "Header2", "Header3"},
		{"Data1", "Data2", "Data3"},
	}

	// Should not panic
	printCsv(records, "dummy_path")
}
