package main

import (
	"fmt"
	"os"
	"reflect"
)

type Sbi struct{}

func (p Sbi) Name() string {
	return "sbi"
}

func (p Sbi) Parse(records [][]string) ([]YnabRecord, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	if !reflect.DeepEqual(records[0], []string{"日付", "内容", "出金金額(円)", "入金金額(円)", "残高(円)", "メモ"}) {
		return nil, nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[1:] {
		amount := row[3]
		if row[2] != "" {
			amount = flipSign(row[2])
		}
		date, err := convertDate("2006/01/02", "2006-01-02", row[0])
		if err != nil {
			// Skip invalid row with warning (like epos.go)
			fmt.Fprintf(os.Stderr, "Warning: skipping row with invalid date: %v\n", err)
			continue
		}
		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: amount,
			memo:   row[1],
		})
	}
	return parsed, nil
}
