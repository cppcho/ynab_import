package main

import (
	"reflect"
)

type Smbc struct{}

func (p Smbc) Name() string {
	return "smbc"
}

func (p Smbc) Parse(records [][]string) []YnabRecord {
	if !reflect.DeepEqual(records[0], []string{"年月日", "お引出し", "お預入れ", "お取り扱い内容", "残高", "メモ", "ラベル"}) {
		return nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[1:] {
		amount := row[2]
		if row[1] != "" {
			amount = flipSign(row[1])
		}
		date, err := convertDate("2006/1/2", "2006-01-02", row[0])
		if err != nil {
			panic(err)
		}
		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: amount,
			payee:  row[3],
		})
	}
	return parsed
}
