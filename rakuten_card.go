package main

type RakutenCard struct{}

func (p RakutenCard) Name() string {
	return "rakuten_card"
}

func (p RakutenCard) Parse(records [][]string) []YnabRecord {
	if len(records[0]) != 10 || records[0][9] != "新規サイン" {
		return nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[1:] {
		date, err := convertDate("2006/01/02", "2006-01-02", row[0])
		if err != nil {
			panic(err)
		}
		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: flipSign(row[6]),
			payee:  row[1],
		})
	}
	return parsed
}
