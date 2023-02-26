package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func writeRecordsToCsv(records []YnabRecord, outputPath string) {
	f, err := os.Create(outputPath)
	defer f.Close()

	if err != nil {
		panic(err)
	}

	w := csv.NewWriter(f)
	w.Write([]string{"Date", "Payee", "Memo", "Amount"})
	for _, record := range records {
		if record.date == "" || record.amount == "" {
			continue
		}
		err = w.Write([]string{record.date, record.payee, record.memo, record.amount})
	}
	if err != nil {
		panic(err)
	}
	w.Flush()
}

func printCsv(records [][]string, outputPath string) {
	for _, record := range records {
		fmt.Println(strings.Join(record, ", "))
	}
}

func readCsvToRawRecords(path string) [][]string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	det := chardet.NewTextDetector()
	detResult, err := det.DetectBest(data)

	var reader io.Reader = bytes.NewReader(data)
	if detResult.Charset == "Shift_JIS" {
		reader = transform.NewReader(reader, japanese.ShiftJIS.NewDecoder())
	}
	csvReader := csv.NewReader(reader)

	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}
	return records
}
