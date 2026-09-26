package main

import (
	"reflect"
	"strings"
)

type Shinsei struct{}

func (p Shinsei) Name() string {
	return "shinsei"
}

func (p Shinsei) Parse(records [][]string) (*ParseResult, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	// Newer exports add a trailing メモ column, and may carry a BOM and
	// literal quotes on the first field.
	header := make([]string, len(records[0]))
	for i, h := range records[0] {
		header[i] = strings.Trim(strings.TrimPrefix(h, "\ufeff"), "\"")
	}
	baseHeaders := []string{"取引日", "摘要", "出金金額", "入金金額", "残高"}
	if !reflect.DeepEqual(header, baseHeaders) &&
		!reflect.DeepEqual(header, append(baseHeaders, "メモ")) {
		return nil, nil
	}

	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	for i, row := range records[1:] {
		amount := row[3]
		if row[2] != "" {
			amount = flipSign(row[2])
		}
		date, err := convertDate("2006/01/02", "2006-01-02", row[0])
		if err != nil {
			skippedRows = append(skippedRows, SkippedRow{
				RowNumber: i + 2, // +2 for header and 0-index
				RawData:   row,
				Reason:    err.Error(),
			})
			continue
		}
		validRecords = append(validRecords, YnabRecord{
			date:   date,
			amount: amount,
			memo:   row[1],
		})
	}

	return &ParseResult{
		ValidRecords: validRecords,
		SkippedRows:  skippedRows,
	}, nil
}
