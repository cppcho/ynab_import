package main

import (
	"reflect"
)

type Epos struct{}

func (p Epos) Name() string {
	return "epos"
}

func (p Epos) Parse(records [][]string) (*ParseResult, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	if !reflect.DeepEqual(records[0], []string{"種別（ショッピング、キャッシング、その他）", "ご利用年月日", "ご利用場所", "ご利用内容", "ご利用金額", "お支払金額（キャッシングでは利息を含みます）", "支払区分"}) {
		return nil, nil
	}

	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	for i, row := range records[1:] {
		if row[1] == "" || row[6] == "" {
			skippedRows = append(skippedRows, SkippedRow{
				RowNumber: i + 2, // +2 for header and 0-index
				RawData:   row,
				Reason:    "missing required fields (date or payment type)",
			})
			continue
		}
		date, err := convertDate("2006年01月02日", "2006-01-02", row[1])
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
			amount: flipSign(row[5]),
			payee:  row[2],
		})
	}

	return &ParseResult{
		ValidRecords: validRecords,
		SkippedRows:  skippedRows,
	}, nil
}
