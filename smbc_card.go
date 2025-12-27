package main

import (
	"fmt"
	"os"
	"strings"
)

type SmbcCard struct{}

func (p SmbcCard) Name() string {
	return "smbc_card"
}

func (p SmbcCard) Parse(records [][]string) ([]YnabRecord, error) {
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

	parsed := make([]YnabRecord, 0)
	for _, row := range records {
		date, err := convertDate("2006/1/2", "2006-01-02", row[0])
		if err != nil {
			// Skip invalid row with warning (like epos.go)
			fmt.Fprintf(os.Stderr, "Warning: skipping row with invalid date: %v\n", err)
			continue
		}

		// Use column 7 for amount, or column 6 if column 7 is empty (international transactions)
		amount := row[7]
		if amount == "" && len(row) > 6 {
			amount = row[6]
		}

		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: flipSign(amount),
			payee:  row[1],
		})
	}
	return parsed, nil
}
