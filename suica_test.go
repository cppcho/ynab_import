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
	// Suica parser only handles MHTML, not CSV
	parser := Suica{}
	result, err := parser.Parse([][]string{{"dummy", "data"}})

	if err != nil {
		t.Errorf("Parse() should not error on CSV input: %v", err)
	}

	if result != nil {
		t.Error("Parse() should return nil for CSV input (Suica uses MHTML)")
	}
}

func TestSuica_ParseMHTML_NonSuicaFile(t *testing.T) {
	parser := Suica{}

	// Using a CSV file which is not MHTML
	result, err := parser.ParseMHTML("testdata/parsers/smbc_valid.csv")

	if err != nil {
		t.Errorf("ParseMHTML() should not error on non-MHTML file: %v", err)
	}

	if result != nil {
		t.Error("ParseMHTML() should return nil for non-Suica file")
	}
}

func TestSuica_ParseMHTML_FileNotFound(t *testing.T) {
	parser := Suica{}

	result, err := parser.ParseMHTML("nonexistent.mhtml")

	if err == nil {
		t.Error("ParseMHTML() should error on nonexistent file")
	}

	if result != nil {
		t.Error("ParseMHTML() should return nil on error")
	}
}

func TestSuica_ParseMHTML_Valid(t *testing.T) {
	parser := Suica{}

	result, err := parser.ParseMHTML("testdata/parsers/suica_valid.mhtml")
	if err != nil {
		t.Fatalf("ParseMHTML() error: %v", err)
	}

	if result == nil {
		t.Fatal("ParseMHTML() returned nil result")
	}

	// The test file has 4 data rows:
	// 1. 03/19 入 京王新宿 -> 京王橋本 -409 (transit, included)
	// 2. 03/18 物販 -248 (shopping, included)
	// 3. 03/18 ｵｰﾄ +10,000 (auto-charge, skipped)
	// 4. 03/17 入 都営六本木 -> 都　新宿 -220 (transit, included)

	if len(result.ValidRecords) != 3 {
		t.Fatalf("expected 3 valid records, got %d", len(result.ValidRecords))
	}

	// Check first record (transit)
	r := result.ValidRecords[0]
	if r.date != "2026-03-19" {
		t.Errorf("record 0 date = %q, want %q", r.date, "2026-03-19")
	}
	if r.payee != "交通" {
		t.Errorf("record 0 payee = %q, want %q", r.payee, "交通")
	}
	if r.memo != "京王新宿 -> 京王橋本" {
		t.Errorf("record 0 memo = %q, want %q", r.memo, "京王新宿 -> 京王橋本")
	}
	if r.amount != "-409" {
		t.Errorf("record 0 amount = %q, want %q", r.amount, "-409")
	}

	// Check second record (shopping)
	r = result.ValidRecords[1]
	if r.date != "2026-03-18" {
		t.Errorf("record 1 date = %q, want %q", r.date, "2026-03-18")
	}
	if r.payee != "物販" {
		t.Errorf("record 1 payee = %q, want %q", r.payee, "物販")
	}
	if r.memo != "" {
		t.Errorf("record 1 memo = %q, want empty", r.memo)
	}
	if r.amount != "-248" {
		t.Errorf("record 1 amount = %q, want %q", r.amount, "-248")
	}

	// Check third record (transit with spaces in station name)
	r = result.ValidRecords[2]
	if r.date != "2026-03-17" {
		t.Errorf("record 2 date = %q, want %q", r.date, "2026-03-17")
	}
	if r.payee != "交通" {
		t.Errorf("record 2 payee = %q, want %q", r.payee, "交通")
	}
	if r.memo != "都営六本木 -> 都　新宿" {
		t.Errorf("record 2 memo = %q, want %q", r.memo, "都営六本木 -> 都　新宿")
	}
	if r.amount != "-220" {
		t.Errorf("record 2 amount = %q, want %q", r.amount, "-220")
	}
}

func TestExtractYearFromMHTML(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			"valid date header",
			"From: test\nDate: Sat, 21 Mar 2026 10:23:31 +0900\nMIME-Version: 1.0",
			2026,
		},
		{
			"different year",
			"Date: Mon, 01 Jan 2025 00:00:00 +0900\n",
			2025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year := extractYearFromMHTML(tt.content)
			if year != tt.expected {
				t.Errorf("extractYearFromMHTML() = %d, want %d", year, tt.expected)
			}
		})
	}
}

func TestSuica_ParseMHTMLTransactions_TicketBusAndAdjustedFare(t *testing.T) {
	row := func(cells ...string) string {
		html := "<tr>"
		for _, c := range cells {
			html += "<td>" + c + "</td>"
		}
		return html + "</tr>"
	}
	html := row("", "09/03", "購", "京王橋本", "", "", "\\7,338", "-410") +
		row("", "08/06", "＊入", "地青山一", "出", "地　渋谷", "\\8,518", "-108") +
		row("", "10/11", "ﾊﾞｽ等", "神姫バス", "", "", "\\15,100", "-230")

	result, err := Suica{}.parseMHTMLTransactions(html, 2026)
	if err != nil {
		t.Fatalf("parseMHTMLTransactions() error: %v", err)
	}
	if len(result.SkippedRows) != 0 {
		t.Errorf("expected no skipped rows, got %+v", result.SkippedRows)
	}

	want := []YnabRecord{
		{date: "2026-09-03", payee: "交通", memo: "京王橋本", amount: "-410"},
		{date: "2026-08-06", payee: "交通", memo: "地青山一 -> 地　渋谷", amount: "-108"},
		{date: "2026-10-11", payee: "交通", memo: "神姫バス", amount: "-230"},
	}
	if len(result.ValidRecords) != len(want) {
		t.Fatalf("got %d records, want %d", len(result.ValidRecords), len(want))
	}
	for i, w := range want {
		if result.ValidRecords[i] != w {
			t.Errorf("record %d = %+v, want %+v", i, result.ValidRecords[i], w)
		}
	}
}
