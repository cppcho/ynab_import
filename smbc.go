package main

import (
	"fmt"
	"os"
	"reflect"
)

type Smbc struct{}

func (p Smbc) Name() string {
	return "smbc"
}

func (p Smbc) Parse(records [][]string) ([]YnabRecord, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	if !reflect.DeepEqual(records[0], []string{"年月日", "お引出し", "お預入れ", "お取り扱い内容", "残高", "メモ", "ラベル"}) {
		return nil, nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[1:] {
		amount := row[2]
		if row[1] != "" {
			amount = flipSign(row[1])
		}
		date, err := convertDate("2006/1/2", "2006-01-02", row[0])
		if err != nil {
			// Skip invalid row with warning (like epos.go)
			fmt.Fprintf(os.Stderr, "Warning: skipping row with invalid date: %v\n", err)
			continue
		}
		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: amount,
			payee:  row[3],
		})
	}
	return parsed, nil
}
