package main

import (
	"reflect"
)

type Rakuten struct{}

func (p Rakuten) Name() string {
	return "rakuten"
}

func (p Rakuten) Parse(records [][]string) (*ParseResult, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	if !reflect.DeepEqual(records[0], []string{"取引日", "入出金(円)", "取引後残高(円)", "入出金内容"}) {
		return nil, nil
	}

	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	for i, row := range records[1:] {
		date, err := convertDate("20060102", "2006-01-02", row[0])
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
			amount: row[1],
			payee:  row[3],
		})
	}

	return &ParseResult{
		ValidRecords: validRecords,
		SkippedRows:  skippedRows,
	}, nil
}
