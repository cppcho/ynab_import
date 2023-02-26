package main

import (
	"reflect"
)

type Rakuten struct{}

func (p Rakuten) Name() string {
	return "rakuten"
}

func (p Rakuten) Parse(records [][]string) []YnabRecord {
	if !reflect.DeepEqual(records[0], []string{"取引日", "入出金(円)", "取引後残高(円)", "入出金内容"}) {
		return nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[1:] {
		parsed = append(parsed, YnabRecord{
			date:   convertDate("20060102", "2006-01-02", row[0]),
			amount: row[1],
			payee:  row[3],
		})
	}
	return parsed
}
