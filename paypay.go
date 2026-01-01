package main

import (
	"reflect"
	"strings"
)

type PayPay struct{}

func (p PayPay) Name() string {
	return "paypay"
}

func (p PayPay) Parse(records [][]string) (*ParseResult, error) {
	// Handle empty records
	if len(records) == 0 {
		return nil, nil // Not my format
	}

	// PayPay CSV may have UTF-8 BOM in the first column
	expectedHeader := []string{"取引日", "出金金額（円）", "入金金額（円）", "海外出金金額", "通貨", "変換レート（円）", "利用国", "取引内容", "取引先", "取引方法", "支払い区分", "利用者", "取引番号"}
	expectedHeaderWithBOM := []string{"\ufeff取引日", "出金金額（円）", "入金金額（円）", "海外出金金額", "通貨", "変換レート（円）", "利用国", "取引内容", "取引先", "取引方法", "支払い区分", "利用者", "取引番号"}

	if !reflect.DeepEqual(records[0], expectedHeader) && !reflect.DeepEqual(records[0], expectedHeaderWithBOM) {
		return nil, nil
	}

	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	for i, row := range records[1:] {
		// Extract date part from datetime (2025/12/27 12:00:46 -> 2025/12/27)
		datePart := strings.Split(row[0], " ")[0]
		date, err := convertDate("2006/1/2", "2006-01-02", datePart)
		if err != nil {
			skippedRows = append(skippedRows, SkippedRow{
				RowNumber: i + 2, // +2 for header and 0-index
				RawData:   row,
				Reason:    err.Error(),
			})
			continue
		}

		// Handle withdrawal (出金金額（円）) vs deposit (入金金額（円）)
		// Withdrawals should be negative, deposits should be positive
		amount := row[2] // Default to deposit
		if row[1] != "" && row[1] != "-" {
			amount = flipSign(row[1]) // Withdrawal (make negative)
		}

		validRecords = append(validRecords, YnabRecord{
			date:   date,
			amount: amount,
			payee:  row[8], // 取引先 (merchant/counterparty)
		})
	}

	return &ParseResult{
		ValidRecords: validRecords,
		SkippedRows:  skippedRows,
	}, nil
}
