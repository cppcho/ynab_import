package main

import (
	"fmt"
	"os"
	"reflect"
)

type Rakuten struct{}

func (p Rakuten) Name() string {
	return "rakuten"
}

func (p Rakuten) Parse(records [][]string) ([]YnabRecord, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	if !reflect.DeepEqual(records[0], []string{"取引日", "入出金(円)", "取引後残高(円)", "入出金内容"}) {
		return nil, nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[1:] {
		date, err := convertDate("20060102", "2006-01-02", row[0])
		if err != nil {
			// Skip invalid row with warning (like epos.go)
			fmt.Fprintf(os.Stderr, "Warning: skipping row with invalid date: %v\n", err)
			continue
		}
		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: row[1],
			payee:  row[3],
		})
	}
	return parsed, nil
}
