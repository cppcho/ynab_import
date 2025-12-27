package main

import (
	"fmt"
	"os"
)

type View struct{}

func (p View) Name() string {
	return "view"
}

func (p View) Parse(records [][]string) ([]YnabRecord, error) {
	if len(records) <= 6 {
		return nil, nil
	}
	if !(records[0][0] == "会員番号" && records[4][0] == "ご利用年月日") {
		return nil, nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[6:] {
		date, err := convertDate("2006/01/02", "2006-01-02", row[0])
		if err != nil {
			// Skip invalid row with warning (like epos.go)
			fmt.Fprintf(os.Stderr, "Warning: skipping row with invalid date: %v\n", err)
			continue
		}
		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: flipSign(row[4]),
			payee:  row[1],
		})
	}
	return parsed, nil
}
