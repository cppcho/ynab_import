package main

import (
	"strings"
)

type SmbcCard struct{}

func (p SmbcCard) Name() string {
	return "smbc_card"
}

func (p SmbcCard) Parse(records [][]string) []YnabRecord {
	if records[0][2] != "ご本人" && records[0][2] != "ご家族" {
		return nil
	}

	if !strings.HasPrefix(records[0][5], "'") {
		return nil
	}

	parsed := make([]YnabRecord, 0)
	for _, row := range records {
		date, err := convertDate("2006/1/2", "2006-01-02", row[0])
		if err != nil {
			panic(err)
		}
		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: flipSign(row[7]),
			payee:  row[1],
		})
	}
	return parsed
}
