package main

import (
	"fmt"
	"os"
)

type RakutenCard struct{}

func (p RakutenCard) Name() string {
	return "rakuten_card"
}

func (p RakutenCard) Parse(records [][]string) ([]YnabRecord, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	if len(records[0]) != 10 || records[0][9] != "新規サイン" {
		return nil, nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[1:] {
		date, err := convertDate("2006/01/02", "2006-01-02", row[0])
		if err != nil {
			// Skip invalid row with warning (like epos.go)
			fmt.Fprintf(os.Stderr, "Warning: skipping row with invalid date: %v\n", err)
			continue
		}
		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: flipSign(row[6]),
			payee:  row[1],
		})
	}
	return parsed, nil
}
