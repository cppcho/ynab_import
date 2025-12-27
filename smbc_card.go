package main

import (
	"strings"
)

type SmbcCard struct{}

func (p SmbcCard) Name() string {
	return "smbc_card"
}

func (p SmbcCard) Parse(records [][]string) (*ParseResult, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	if records[0][2] != "ご本人" && records[0][2] != "ご家族" {
		return nil, nil
	}

	if !strings.HasPrefix(records[0][5], "'") {
		return nil, nil
	}

	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	for i, row := range records {
		date, err := convertDate("2006/1/2", "2006-01-02", row[0])
		if err != nil {
			skippedRows = append(skippedRows, SkippedRow{
				RowNumber: i + 1, // +1 for 0-index (no header skip)
				RawData:   row,
				Reason:    err.Error(),
			})
			continue
		}

		// Use column 7 for amount, or column 6 if column 7 is empty (international transactions)
		amount := row[7]
		if amount == "" && len(row) > 6 {
			amount = row[6]
		}

		validRecords = append(validRecords, YnabRecord{
			date:   date,
			amount: flipSign(amount),
			payee:  row[1],
		})
	}

	return &ParseResult{
		ValidRecords: validRecords,
		SkippedRows:  skippedRows,
	}, nil
}
