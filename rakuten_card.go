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
		parsed = append(parsed, YnabRecord{
			date:   convertDate("2006/01/02", "2006-01-02", row[0]),
			amount: flipSign(row[6]),
			memo:   row[1],
		})
	}
	return parsed
}
