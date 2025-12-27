package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func expandHomeDir(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if len(path) == 1 {
		return homeDir
	}
	return homeDir + path[1:]
}

func main() {
	// Define CLI flags with environment variable defaults
	inputDir := flag.String("input", getEnvOrDefault("CSV_DIR_IN", "~/Downloads"), "Input directory containing CSV files (env: CSV_DIR_IN, default: ~/Downloads)")
	outputDir := flag.String("output", getEnvOrDefault("CSV_DIR", "~/Desktop"), "Output directory for converted CSV files (env: CSV_DIR, default: ~/Desktop)")
	flag.Parse()

	// Expand ~ in paths
	*inputDir = expandHomeDir(*inputDir)
	*outputDir = expandHomeDir(*outputDir)

	files, err := os.ReadDir(*inputDir)
	if err != nil {
		panic(err)
	}

	// Create output dir (e.g. ~/Desktop/20060102_output)
	now := time.Now().UTC().Format("20060102")
	timestampedOutputDir := path.Join(*outputDir, now+"_output")
	err = os.MkdirAll(timestampedOutputDir, 0755)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") {
			srcPath := path.Join(*inputDir, file.Name())
			fmt.Printf("Parsing %v ...", srcPath)

			rawRecords, err := readCsvToRawRecords(srcPath)
			if err != nil {
				panic(err)
			}

			var parsed []YnabRecord
			for _, parser := range parsers {
				parsed = parser.Parse(rawRecords)
				if parsed != nil {
					fmt.Printf("Matched parser %v\n", parser.Name())
					dstPath := path.Join(timestampedOutputDir, parser.Name()+"_"+file.Name())
					err = writeRecordsToCsv(parsed, dstPath)
					if err != nil {
						panic(err)
					}
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
func convertDate(fromLayout, toLayout, value string) (string, error) {
	date, err := time.Parse(fromLayout, value)
	if err != nil {
		return "", err
	}
	return date.Format(toLayout), nil
}
