package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const CSV_DIR = "/Users/cppcho/Desktop"

type YnabRecord struct {
	date   string
	payee  string
	memo   string
	amount string
}

type Parser interface {
	Name() string
	Parse(records [][]string) []YnabRecord
}

var parsers []Parser = []Parser{Smbc{}, Rakuten{}, Epos{}, View{}, Saison{}, RakutenCard{}, Sbi{}, SmbcCard{}}

func flipSign(str string) string {
	str = strings.Replace(str, ",", "", -1)
	val, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("err: invalid str for flipSign " + str)
		return str
	}
	return strconv.Itoa(val * -1)
}

func main() {
	files, err := os.ReadDir(CSV_DIR)
	if err != nil {
		panic(err)
	}

	// Create output dir (e.g. ~/Desktop/20060102_output)
	now := time.Now().UTC().Format("20060102")
	outputDir := path.Join(CSV_DIR, now+"_output")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") {
			srcPath := path.Join(CSV_DIR, file.Name())
			fmt.Printf("Parsing %v ...", srcPath)

			rawRecords := readCsvToRawRecords(srcPath)

			var parsed []YnabRecord
			for _, parser := range parsers {
				parsed = parser.Parse(rawRecords)
				if parsed != nil {
					fmt.Printf("Matched parser %v\n", parser.Name())
					dstPath := path.Join(outputDir, parser.Name()+".csv")
					writeRecordsToCsv(parsed, dstPath)
					fmt.Printf("Write csv to %v\n", dstPath)
					break
				}
			}
			if parsed == nil {
				fmt.Println("No matched parser")
			}
		}
	}
}

// 2006-01-02T15:04:05
func convertDate(fromLayout, toLayout, value string) string {
	date, err := time.Parse(fromLayout, value)
	if err != nil {
		panic(err)
	}
	return date.Format(toLayout)
}

func parseCsv(path string) [][]string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	det := chardet.NewTextDetector()
	detResult, err := det.DetectBest(data)

	fmt.Println(path, detResult.Charset)
	var reader io.Reader = bytes.NewReader(data)
	if detResult.Charset == "Shift_JIS" {
		reader = transform.NewReader(reader, japanese.ShiftJIS.NewDecoder())
	}
	csvReader := csv.NewReader(reader)

	// header
	headers, err := csvReader.Read()
	fmt.Println(len(headers), headers)

	output := make([][]string, 0)
	output = append(output, []string{"Date", "Payee", "Memo", "Amount"})

	if len(headers) == 7 && headers[0] == "年月日" {
		// smbc bank
		// 年月日 お引出し お預入れ お取り扱い内容 残高 メモ ラベル
		for {
			row := make([]string, 4)
			rec, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			row[0] = rec[0]
			row[2] = rec[3]
			if rec[1] == "" {
				row[3] = rec[2]
			} else {
				row[3] = "-" + rec[1]
			}
			output = append(output, row)
		}
	} else if len(headers) == 4 && headers[3] == "入出金内容" {
		// Rakuten bank
		// 取引日 入出金(円) 取引後残高(円) 入出金内容
		for {
			row := make([]string, 4)
			rec, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			date, err := time.Parse("20060102", rec[0])
			if err != nil {
				panic(err)
			}
			row[0] = date.Format("2006-01-02")
			row[3] = rec[1]
			row[2] = rec[3]
			output = append(output, row)
		}
	} else if len(headers) == 7 && headers[1] == "ご利用年月日" {
		// Epos
		// 種別（ショッピング、キャッシング、その他） ご利用年月日 ご利用場所 ご利用内容 ご利用金額 お支払金額（キャッシングでは利息を含みます） 支払区分
		for {
			row := make([]string, 4)
			rec, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			if rec[1] == "" {
				break
			}
			date, err := time.Parse("2006年01月02日", rec[1])
			if err != nil {
				panic(err)
			}
			row[0] = date.Format("2006-01-02")
			row[2] = rec[2]
			row[3] = "-" + rec[5]
			output = append(output, row)
		}
	} else {
		fmt.Println("skip")
	}
	return output
}
