package main

import (
	"fmt"
	"os"
)

type Saison struct{}

func (p Saison) Name() string {
	return "saison"
}

func (p Saison) Parse(records [][]string) ([]YnabRecord, error) {
	if len(records) <= 4 {
		return nil, nil
	}
	if !(records[0][0] == "カード名称" && records[3][0] == "利用日") {
		return nil, nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[4:] {
		date, err := convertDate("2006/01/02", "2006-01-02", row[0])
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
	return parsed, nil
}
