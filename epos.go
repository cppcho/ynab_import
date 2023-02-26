package main

import (
	"reflect"
)

type Epos struct{}

func (p Epos) Name() string {
	return "epos"
}

func (p Epos) Parse(records [][]string) []YnabRecord {
	if !reflect.DeepEqual(records[0], []string{"種別（ショッピング、キャッシング、その他）", "ご利用年月日", "ご利用場所", "ご利用内容", "ご利用金額", "お支払金額（キャッシングでは利息を含みます）", "支払区分"}) {
		return nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[1:] {
		if row[1] == "" || row[6] == "" {
			continue
		}
		parsed = append(parsed, YnabRecord{
			date:   convertDate("2006年01月02日", "2006-01-02", row[1]),
			amount: flipSign(row[5]),
			payee:  row[2],
		})
	}
	return parsed
}
