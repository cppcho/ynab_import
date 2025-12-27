package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type SmbcCard2 struct{}

func (p SmbcCard2) Name() string {
	return "smbc_card2"
}

func (p SmbcCard2) Parse(records [][]string) ([]YnabRecord, error) {
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

	parsed := make([]YnabRecord, 0)
	for _, row := range records {
		// if row[0] is in date format
		if _, err := time.Parse("2006/1/2", row[0]); err == nil {
			date, err := convertDate("2006/1/2", "2006-01-02", row[0])
			if err != nil {
				// Skip invalid row with warning (like epos.go)
				fmt.Fprintf(os.Stderr, "Warning: skipping row with invalid date: %v\n", err)
				continue
			}
			parsed = append(parsed, YnabRecord{
				date:   date,
				amount: flipSign(row[5]),
				payee:  row[1],
			})
		}
	}
	return parsed, nil
}
