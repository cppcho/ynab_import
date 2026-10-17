package main

import (
	"fmt"
	"io"
	"mime/quotedprintable"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type Suica struct{}

func (p Suica) Name() string {
	return "suica"
}

// ParseMHTML parses a Mobile Suica MHTML file saved from the website
func (p Suica) ParseMHTML(filePath string) (*ParseResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Check if this is a Mobile Suica MHTML
	if !strings.Contains(content, "mobilesuica.com") {
		return nil, nil
	}

	// Extract year from the MHTML Date header
	year := extractYearFromMHTML(content)

	// Find the HTML part (quoted-printable encoded, Shift_JIS)
	html, err := extractHTMLFromMHTML(content)
	if err != nil {
		return nil, err
	}

	return p.parseMHTMLTransactions(html, year)
}

// Parse implements the Parser interface for CSV compatibility
func (p Suica) Parse(records [][]string) (*ParseResult, error) {
	// Suica uses MHTML format, not CSV
	return nil, nil
}

// extractYearFromMHTML extracts the year from the Date header in MHTML
func extractYearFromMHTML(content string) int {
	// Look for: Date: Sat, 21 Mar 2026 10:23:31 +0900
	re := regexp.MustCompile(`(?m)^Date:\s+.+\s+(\d{4})\s+`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		year, err := strconv.Atoi(matches[1])
		if err == nil {
			return year
		}
	}
	return time.Now().Year()
}

// extractHTMLFromMHTML extracts and decodes the HTML content from MHTML
func extractHTMLFromMHTML(content string) (string, error) {
	// Find the HTML part after Content-Transfer-Encoding: quoted-printable
	// and Content-Type: text/html
	idx := strings.Index(content, "Content-Transfer-Encoding: quoted-printable")
	if idx == -1 {
		return "", fmt.Errorf("no quoted-printable content found in MHTML")
	}

	// Find the blank line that separates headers from body
	bodyStart := strings.Index(content[idx:], "\r\n\r\n")
	if bodyStart == -1 {
		bodyStart = strings.Index(content[idx:], "\n\n")
		if bodyStart == -1 {
			return "", fmt.Errorf("no body found after content headers")
		}
		bodyStart += 2
	} else {
		bodyStart += 4
	}
	bodyContent := content[idx+bodyStart:]

	// Find end of this MIME part (next boundary)
	boundaryIdx := strings.Index(bodyContent, "------MultipartBoundary")
	if boundaryIdx != -1 {
		bodyContent = bodyContent[:boundaryIdx]
	}

	// Decode quoted-printable
	qpReader := quotedprintable.NewReader(strings.NewReader(bodyContent))
	decoded, err := io.ReadAll(qpReader)
	if err != nil {
		return "", fmt.Errorf("failed to decode quoted-printable: %w", err)
	}

	// Decode Shift_JIS to UTF-8
	utf8Reader := transform.NewReader(
		strings.NewReader(string(decoded)),
		japanese.ShiftJIS.NewDecoder(),
	)
	utf8Bytes, err := io.ReadAll(utf8Reader)
	if err != nil {
		return "", fmt.Errorf("failed to decode Shift_JIS: %w", err)
	}

	return string(utf8Bytes), nil
}

// tagContentRe matches content inside <font ...>...</font> tags
var tagContentRe = regexp.MustCompile(`<font[^>]*>(.*?)</font>`)

// trRe matches <tr>...</tr> blocks
var trRe = regexp.MustCompile(`(?s)<tr>(.*?)</tr>`)

// tdRe matches <td ...>...</td> blocks
var tdRe = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)

func (p Suica) parseMHTMLTransactions(html string, year int) (*ParseResult, error) {
	var validRecords []YnabRecord
	var skippedRows []SkippedRow

	// Find all table rows
	trMatches := trRe.FindAllStringSubmatch(html, -1)

	rowNum := 0
	for _, trMatch := range trMatches {
		rowContent := trMatch[1]

		// Extract all <td> cells
		tdMatches := tdRe.FindAllStringSubmatch(rowContent, -1)
		if len(tdMatches) != 8 {
			continue // Not a data row (header has 8 cols too but we'll filter by date)
		}

		rowNum++

		// Extract text from each cell (strip HTML tags)
		cells := make([]string, 8)
		for i, td := range tdMatches {
			cells[i] = extractCellText(td[1])
		}

		// cells[0] = checkbox (ignore)
		// cells[1] = date (MM/DD)
		// cells[2] = type1 (入, 物販, ｵｰﾄ, etc.)
		// cells[3] = location1
		// cells[4] = type2 (出, 　, etc.)
		// cells[5] = location2
		// cells[6] = balance
		// cells[7] = amount

		// Parse date
		dateParts := strings.Split(cells[1], "/")
		if len(dateParts) != 2 {
			continue
		}
		month, err := strconv.Atoi(dateParts[0])
		if err != nil || month < 1 || month > 12 {
			continue
		}
		day, err := strconv.Atoi(dateParts[1])
		if err != nil || day < 1 || day > 31 {
			continue
		}

		// Parse amount
		amountStr := strings.ReplaceAll(strings.TrimSpace(cells[7]), ",", "")
		if amountStr == "" {
			continue
		}
		amount, err := strconv.Atoi(amountStr)
		if err != nil {
			continue
		}

		// Skip positive amounts (auto-charge, deposits)
		if amount > 0 {
			continue
		}

		// A leading ＊ marks a fare Suica adjusted (e.g. a transfer discount);
		// the transaction itself is an ordinary one of the same type.
		transType := strings.TrimPrefix(strings.TrimSpace(cells[2]), "＊")
		var payee, memo string

		if transType == "物販" {
			payee = "物販"
		} else if transType == "購" || transType == "ﾊﾞｽ等" {
			// Ticket bought at a station, or a bus fare; cells[3] names the
			// station or the bus operator
			payee = "交通"
			memo = strings.TrimSpace(cells[3])
		} else if transType == "入" {
			payee = "交通"
			fromStation := strings.TrimSpace(cells[3])
			toStation := strings.TrimSpace(cells[5])
			if fromStation != "" && toStation != "" {
				memo = fmt.Sprintf("%s -> %s", fromStation, toStation)
			}
		} else if transType == "ｵｰﾄ" {
			// Skip auto-charge
			continue
		} else {
			// Unknown transaction type, skip
			skippedRows = append(skippedRows, SkippedRow{
				RowNumber: rowNum,
				RawData:   cells[1:],
				Reason:    fmt.Sprintf("unknown transaction type: %s", transType),
			})
			continue
		}

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

// extractCellText extracts text content from an HTML cell, stripping tags
func extractCellText(cellHTML string) string {
	// Try to extract from <font> tag first
	fontMatch := tagContentRe.FindStringSubmatch(cellHTML)
	if fontMatch != nil {
		return strings.TrimSpace(fontMatch[1])
	}
	// Strip all HTML tags as fallback
	tagRe := regexp.MustCompile(`<[^>]*>`)
	text := tagRe.ReplaceAllString(cellHTML, "")
	return strings.TrimSpace(text)
}
