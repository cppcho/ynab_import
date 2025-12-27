package main

type Saison struct{}

func (p Saison) Name() string {
	return "saison"
}

func (p Saison) Parse(records [][]string) (*ParseResult, error) {
	if len(records) <= 4 {
		return nil, nil
	}
	if !(records[0][0] == "カード名称" && records[3][0] == "利用日") {
		return nil, nil
	}

	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	for i, row := range records[4:] {
		date, err := convertDate("2006/01/02", "2006-01-02", row[0])
		if err != nil {
			skippedRows = append(skippedRows, SkippedRow{
				RowNumber: i + 5, // +5 for 4 header rows and 0-index
				RawData:   row,
				Reason:    err.Error(),
			})
			continue
		}
		validRecords = append(validRecords, YnabRecord{
			date:   date,
			amount: flipSign(row[5]),
			payee:  row[1],
		})
	}

	return &ParseResult{
		ValidRecords: validRecords,
		SkippedRows:  skippedRows,
	}, nil
}
