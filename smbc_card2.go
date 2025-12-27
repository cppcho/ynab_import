package main

import (
	"strings"
	"time"
)

type SmbcCard2 struct{}

func (p SmbcCard2) Name() string {
	return "smbc_card2"
}

func (p SmbcCard2) Parse(records [][]string) (*ParseResult, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	if !strings.HasSuffix(records[0][0], "様") {
		return nil, nil
	}
	if !strings.HasSuffix(records[0][1], "****") {
		return nil, nil
	}

	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	for i, row := range records {
		// if row[0] is in date format
		if _, err := time.Parse("2006/1/2", row[0]); err == nil {
			date, err := convertDate("2006/1/2", "2006-01-02", row[0])
			if err != nil {
				skippedRows = append(skippedRows, SkippedRow{
					RowNumber: i + 1, // +1 for 0-index (no header skip)
					RawData:   row,
					Reason:    err.Error(),
				})
				continue
			}
			validRecords = append(validRecords, YnabRecord{
				date:   date,
				amount: flipSign(row[5]),
				payee:  row[1],
			})
		}
	}

	return &ParseResult{
		ValidRecords: validRecords,
		SkippedRows:  skippedRows,
	}, nil
}
