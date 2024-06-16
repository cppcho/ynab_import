package main

import (
	"strings"
	"time"
)

type SmbcCard2 struct{}

func (p SmbcCard2) Name() string {
	return "smbc_card2"
}

func (p SmbcCard2) Parse(records [][]string) []YnabRecord {
	if !strings.HasSuffix(records[0][0], "様") {
		return nil
	}
	if !strings.HasSuffix(records[0][1], "****") {
		return nil
	}

	parsed := make([]YnabRecord, 0)
	for _, row := range records {
		// if row[0] is in date format
		if _, err := time.Parse("2006/1/2", row[0]); err == nil {
			parsed = append(parsed, YnabRecord{
				date:   convertDate("2006/1/2", "2006-01-02", row[0]),
				amount: flipSign(row[5]),
				payee:  row[1],
			})
		}
	}
	return parsed
}
