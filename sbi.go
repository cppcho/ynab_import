package main

import (
	"reflect"
)

type Sbi struct{}

func (p Sbi) Name() string {
	return "sbi"
}

func (p Sbi) Parse(records [][]string) (*ParseResult, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	if !reflect.DeepEqual(records[0], []string{"日付", "内容", "出金金額(円)", "入金金額(円)", "残高(円)", "メモ"}) {
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
