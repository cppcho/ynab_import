package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const CSV_DIR_IN = "/Users/cppcho/Downloads"
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

var parsers []Parser = []Parser{Smbc{}, Rakuten{}, Epos{}, View{}, Saison{}, RakutenCard{}, Sbi{}, SmbcCard{}, SmbcCard2{}}

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
	files, err := os.ReadDir(CSV_DIR_IN)
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
			srcPath := path.Join(CSV_DIR_IN, file.Name())
			fmt.Printf("Parsing %v ...", srcPath)

			rawRecords := readCsvToRawRecords(srcPath)

			var parsed []YnabRecord
			for _, parser := range parsers {
				parsed = parser.Parse(rawRecords)
				if parsed != nil {
					fmt.Printf("Matched parser %v\n", parser.Name())
					dstPath := path.Join(outputDir, parser.Name()+"_"+file.Name())
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
