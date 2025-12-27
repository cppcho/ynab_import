package main

type Saison struct{}

func (p Saison) Name() string {
	return "saison"
}

func (p Saison) Parse(records [][]string) []YnabRecord {
	if len(records) <= 4 {
		return nil
	}
	if !(records[0][0] == "カード名称" && records[3][0] == "利用日") {
		return nil
	}
	parsed := make([]YnabRecord, 0)
	for _, row := range records[4:] {
		date, err := convertDate("2006/01/02", "2006-01-02", row[0])
		if err != nil {
			panic(err)
		}
		parsed = append(parsed, YnabRecord{
			date:   date,
			amount: flipSign(row[5]),
			payee:  row[1],
		})
	}
	return parsed
}
