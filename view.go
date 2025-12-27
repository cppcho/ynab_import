package main

type View struct{}

func (p View) Name() string {
	return "view"
}

func (p View) Parse(records [][]string) (*ParseResult, error) {
	if len(records) <= 6 {
		return nil, nil
	}
	if !(records[0][0] == "会員番号" && records[4][0] == "ご利用年月日") {
		return nil, nil
	}

	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	for i, row := range records[6:] {
		date, err := convertDate("2006/01/02", "2006-01-02", row[0])
		if err != nil {
			skippedRows = append(skippedRows, SkippedRow{
				RowNumber: i + 7, // +7 for 6 header rows and 0-index
				RawData:   row,
				Reason:    err.Error(),
			})
			continue
		}
		validRecords = append(validRecords, YnabRecord{
			date:   date,
			amount: flipSign(row[4]),
			payee:  row[1],
		})
	}

	return &ParseResult{
		ValidRecords: validRecords,
		SkippedRows:  skippedRows,
	}, nil
}
