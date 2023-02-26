package main

type View struct{}

func (p View) Name() string {
	return "view"
}

func (p View) Parse(records [][]string) []YnabRecord {
	if len(records) <= 6 {
		return nil
	}
	if !(records[0][0] == "会員番号" && records[4][0] == "ご利用年月日") {
		return nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[6:] {
		parsed = append(parsed, YnabRecord{
			date:   convertDate("2006/01/02", "2006-01-02", row[0]),
			amount: flipSign(row[4]),
			payee:  row[1],
		})
	}
	return parsed
}
