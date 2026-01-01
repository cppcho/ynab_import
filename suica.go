package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Suica struct{}

func (p Suica) Name() string {
	return "suica"
}

// ParsePDF extracts text from PDF and parses Suica transactions
func (p Suica) ParsePDF(filePath string) (*ParseResult, error) {
	// Extract text from PDF
	text, err := extractPDFText(filePath)
	if err != nil {
		return nil, err
	}

	// Check if this is a Suica PDF by looking for the header (with flexible spacing)
	if !strings.Contains(text, "Ｓｕｉｃａ") || !strings.Contains(text, "残高ご利用明細") {
		return nil, nil // Not a Suica PDF
	}

	// Extract year from filename or use current year
	year := extractYearFromFilename(filePath)
	if year == 0 {
		year = time.Now().Year()
	}

	// Parse transactions from text
	return p.parseTransactions(text, year)
}

// Parse implements the Parser interface for CSV compatibility
func (p Suica) Parse(records [][]string) (*ParseResult, error) {
	// Suica uses PDF format, not CSV
	return nil, nil
}

func extractPDFText(filePath string) (string, error) {
	// Use pdftotext command-line tool with -layout option to preserve table structure
	cmd := exec.Command("pdftotext", "-layout", filePath, "-")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if pdftotext is not installed
		if err.Error() == "exec: \"pdftotext\": executable file not found in $PATH" {
			return "", fmt.Errorf("pdftotext not found: please install poppler-utils (brew install poppler on macOS, apt-get install poppler-utils on Linux)")
		}
		return "", fmt.Errorf("failed to extract PDF text: %w (stderr: %s)", err, stderr.String())
	}

	return out.String(), nil
}

func extractYearFromFilename(filename string) int {
	// Filename format: JE000000000000000_20251028_20260101110125.pdf
	// Extract date portion: 20251028
	re := regexp.MustCompile(`_(\d{8})_`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) > 1 {
		dateStr := matches[1]
		if len(dateStr) >= 4 {
			year, err := strconv.Atoi(dateStr[:4])
			if err == nil {
				return year
			}
		}
	}
	return 0
}

func (p Suica) parseTransactions(text string, year int) (*ParseResult, error) {
	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	lines := strings.Split(text, "\n")

	// pdftotext -layout gives us one transaction per line:
	// Format: "月 日 種別 利用駅 種別 利用駅 残高 入金・利用額"
	// Example: "12      27   入       京王橋本     出      調布           \14,173         -314"
	// Example: "12      27   物販                                   \14,487         -170"

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and headers
		if line == "" || strings.Contains(line, "モバイル") || strings.Contains(line, "残高履歴") {
			continue
		}

		// Split by whitespace
		fields := strings.Fields(line)

		if len(fields) < 4 {
			continue
		}

		// Check if first field is month (1-12)
		month, err := strconv.Atoi(fields[0])
		if err != nil || month < 1 || month > 12 {
			continue
		}

		// Check if second field is day (1-31)
		day, err := strconv.Atoi(fields[1])
		if err != nil || day < 1 || day > 31 {
			continue
		}

		// Third field should be transaction type
		transType := fields[2]

		// Last field should be amount (starts with + or -)
		amountStr := fields[len(fields)-1]
		if !strings.HasPrefix(amountStr, "+") && !strings.HasPrefix(amountStr, "-") {
			continue
		}

		// Parse amount
		amountStr = strings.Replace(amountStr, ",", "", -1)
		amount, err := strconv.Atoi(amountStr)
		if err != nil {
			continue
		}

		// Skip positive amounts (auto-charge, deposits)
		if amount > 0 {
			continue
		}

		var payee, memo string

		if transType == "物販" {
			payee = "物販"
			memo = ""
		} else if transType == "入" {
			payee = "交通"
			// Extract station names
			// Fields format: [month, day, 入, from_stations..., 出, to_stations..., balance, amount]
			// Find the index of "出"
			exitIndex := -1
			for j := 3; j < len(fields); j++ {
				if fields[j] == "出" {
					exitIndex = j
					break
				}
			}

			if exitIndex > 3 {
				// From station is everything between 入 (field 2) and 出
				fromStation := strings.Join(fields[3:exitIndex], "")

				// To station is everything between 出 and balance (starts with \)
				toStationFields := []string{}
				for k := exitIndex + 1; k < len(fields); k++ {
					if strings.HasPrefix(fields[k], "\\") {
						break
					}
					toStationFields = append(toStationFields, fields[k])
				}
				toStation := strings.Join(toStationFields, "")

				if fromStation != "" && toStation != "" {
					memo = fmt.Sprintf("%s -> %s", fromStation, toStation)
				}
			}
		} else if transType == "ｵｰﾄ" {
			// Skip auto-charge
			continue
		} else {
			// Unknown transaction type, skip
			continue
		}

		// Format date
		date := fmt.Sprintf("%04d-%02d-%02d", year, month, day)

		validRecords = append(validRecords, YnabRecord{
			date:   date,
			payee:  payee,
			memo:   memo,
			amount: fmt.Sprintf("%d", amount),
		})
	}

	return &ParseResult{
		ValidRecords: validRecords,
		SkippedRows:  skippedRows,
	}, nil
}
