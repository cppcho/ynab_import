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

func writeRecordsToCsv(records []YnabRecord, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	err = w.Write([]string{"Date", "Payee", "Memo", "Amount"})
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.date == "" || record.amount == "" {
			continue
		}
		err = w.Write([]string{record.date, record.payee, record.memo, flipSign(flipSign(record.amount))})
		if err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func printCsv(records [][]string, outputPath string) {
	for _, record := range records {
		fmt.Println(strings.Join(record, ", "))
	}
}

func readCsvToRawRecords(path string) ([][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	det := chardet.NewTextDetector()
	detResult, err := det.DetectBest(data)
	if err != nil {
		return nil, err
	}

	var reader io.Reader = bytes.NewReader(data)
	if detResult.Charset == "Shift_JIS" {
		reader = transform.NewReader(reader, japanese.ShiftJIS.NewDecoder())
	}
	csvReader := csv.NewReader(reader)

	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}
