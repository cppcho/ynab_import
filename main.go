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
	Parse(records [][]string) ([]YnabRecord, error)
}

var parsers []Parser = []Parser{Smbc{}, Rakuten{}, Epos{}, View{}, Saison{}, RakutenCard{}, Sbi{}, SmbcCard{}, SmbcCard2{}}

func flipSign(str string) string {
	// Remove commas
	str = strings.Replace(str, ",", "", -1)

	// Handle empty strings
	if str == "" {
		return "0"
	}

	// Try parsing as integer
	val, err := strconv.Atoi(str)
	if err != nil {
		// Try parsing as float and convert to int
		floatVal, floatErr := strconv.ParseFloat(str, 64)
		if floatErr != nil {
			fmt.Printf("err: invalid str for flipSign: %q (not a number)\n", str)
			return "0"
		}
		val = int(floatVal)
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
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Define CLI flags with environment variable defaults
	inputDir := flag.String("input", getEnvOrDefault("CSV_DIR_IN", "~/Downloads"), "Input directory containing CSV files (env: CSV_DIR_IN, default: ~/Downloads)")
	outputDir := flag.String("output", getEnvOrDefault("CSV_DIR", "~/Desktop"), "Output directory for converted CSV files (env: CSV_DIR, default: ~/Desktop)")
	flag.Parse()

	// Expand ~ in paths
	*inputDir = expandHomeDir(*inputDir)
	*outputDir = expandHomeDir(*outputDir)

	files, err := os.ReadDir(*inputDir)
	if err != nil {
		return fmt.Errorf("failed to read input directory %q: %w", *inputDir, err)
	}

	// Create output dir (e.g. ~/Desktop/20060102_output)
	now := time.Now().UTC().Format("20060102")
	timestampedOutputDir := path.Join(*outputDir, now+"_output")
	if err := os.MkdirAll(timestampedOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %q: %w", timestampedOutputDir, err)
	}

	// Track errors but continue processing
	var errors []error
	successCount := 0

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") {
			srcPath := path.Join(*inputDir, file.Name())
			fmt.Printf("Parsing %v ...", srcPath)

			if err := processCSV(srcPath, timestampedOutputDir, file.Name()); err != nil {
				fmt.Printf(" ERROR: %v\n", err)
				errors = append(errors, fmt.Errorf("%s: %w", file.Name(), err))
			} else {
				successCount++
			}
		}
	}

	// Report summary
	if len(errors) > 0 {
		fmt.Printf("\nCompleted with %d success(es) and %d error(s)\n", successCount, len(errors))
		return fmt.Errorf("encountered %d error(s) during processing", len(errors))
	}

	if successCount > 0 {
		fmt.Printf("\nSuccessfully processed %d file(s)\n", successCount)
	}
	return nil
}

func processCSV(srcPath, outputDir, fileName string) error {
	rawRecords, err := readCsvToRawRecords(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	for _, parser := range parsers {
		parsed, err := parser.Parse(rawRecords)

		// Error occurred during parsing
		if err != nil {
			return fmt.Errorf("parser %s failed: %w", parser.Name(), err)
		}

		// No match (not this parser's format)
		if parsed == nil {
			continue
		}

		// Match found - write output
		fmt.Printf("Matched parser %v\n", parser.Name())
		dstPath := path.Join(outputDir, parser.Name()+"_"+fileName)

		if err := writeRecordsToCsv(parsed, dstPath); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}

		fmt.Printf("Write csv to %v\n", dstPath)
		return nil // Success
	}

	fmt.Println("No matched parser")
	return nil // Not an error - just no parser matched
}

// 2006-01-02T15:04:05
func convertDate(fromLayout, toLayout, value string) (string, error) {
	date, err := time.Parse(fromLayout, value)
	if err != nil {
		return "", err
	}
	return date.Format(toLayout), nil
}
