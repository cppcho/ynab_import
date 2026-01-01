package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type YnabRecord struct {
	date   string
	payee  string
	memo   string
	amount string
}

type ParseResult struct {
	ValidRecords []YnabRecord
	SkippedRows  []SkippedRow
}

type SkippedRow struct {
	RowNumber int
	RawData   []string
	Reason    string
}

type Parser interface {
	Name() string
	Parse(records [][]string) (*ParseResult, error)
}

var parsers []Parser = []Parser{Smbc{}, Rakuten{}, Epos{}, View{}, Saison{}, RakutenCard{}, Sbi{}, SmbcCard{}, SmbcCard2{}, Shinsei{}, Suica{}, PayPay{}}

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
			fmt.Fprintf(os.Stderr, "Error: invalid value for flipSign: %q (not a number)\n", str)
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

func processFile(filePath, outputDir string) error {
	// Check if this is a PDF file
	if strings.HasSuffix(filePath, ".pdf") {
		return processPDFFile(filePath, outputDir)
	}

	fmt.Printf("Parsing %v ...", filePath)

	rawRecords, err := readCsvToRawRecords(filePath)
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	fileName := path.Base(filePath)
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
		fmt.Printf(" Matched parser %v\n", parser.Name())
		dstPath := path.Join(outputDir, parser.Name()+"_"+fileName)

		if err := writeRecordsToCsv(parsed.ValidRecords, dstPath); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}

		// Display statistics
		fmt.Printf("Converted %d row(s)", len(parsed.ValidRecords))
		if len(parsed.SkippedRows) > 0 {
			fmt.Printf(", skipped %d row(s)", len(parsed.SkippedRows))
		}
		fmt.Printf("\n")

		// Display skipped rows with details
		if len(parsed.SkippedRows) > 0 {
			for _, skipped := range parsed.SkippedRows {
				fmt.Fprintf(os.Stderr, "Skipped row %d: %v (reason: %s)\n",
					skipped.RowNumber, skipped.RawData, skipped.Reason)
			}
		}

		fmt.Printf("Wrote to %v\n", dstPath)
		return nil // Success
	}

	fmt.Println(" No matched parser")
	return nil // Not an error - just no parser matched
}

func processPDFFile(filePath, outputDir string) error {
	fmt.Printf("Parsing %v ...", filePath)

	fileName := path.Base(filePath)
	baseName := strings.TrimSuffix(fileName, ".pdf") + ".csv"

	// Try Suica parser (currently only PDF parser)
	suicaParser := Suica{}
	parsed, err := suicaParser.ParsePDF(filePath)

	// Error occurred during parsing
	if err != nil {
		return fmt.Errorf("parser %s failed: %w", suicaParser.Name(), err)
	}

	// No match (not this parser's format)
	if parsed == nil {
		fmt.Println(" No matched parser")
		return nil
	}

	// Match found - write output
	fmt.Printf(" Matched parser %v\n", suicaParser.Name())
	dstPath := path.Join(outputDir, suicaParser.Name()+"_"+baseName)

	if err := writeRecordsToCsv(parsed.ValidRecords, dstPath); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Display statistics
	fmt.Printf("Converted %d row(s)", len(parsed.ValidRecords))
	if len(parsed.SkippedRows) > 0 {
		fmt.Printf(", skipped %d row(s)", len(parsed.SkippedRows))
	}
	fmt.Printf("\n")

	// Display skipped rows with details
	if len(parsed.SkippedRows) > 0 {
		for _, skipped := range parsed.SkippedRows {
			fmt.Fprintf(os.Stderr, "Skipped row %d: %v (reason: %s)\n",
				skipped.RowNumber, skipped.RawData, skipped.Reason)
		}
	}

	fmt.Printf("Wrote to %v\n", dstPath)
	return nil // Success
}

func processDirectory(inputDir, outputDir string) error {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("failed to read input directory %q: %w", inputDir, err)
	}

	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".csv") || strings.HasSuffix(file.Name(), ".pdf")) {
			srcPath := path.Join(inputDir, file.Name())
			if err := processFile(srcPath, outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", srcPath, err)
			}
		}
	}
	return nil
}

func watchMode(inputDir, outputDir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Process existing files first
	fmt.Println("Processing existing files...")
	if err := processDirectory(inputDir, outputDir); err != nil {
		return err
	}

	// Start watching
	err = watcher.Add(inputDir)
	if err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	fmt.Printf("Watching %s for new or changed CSV files... (Press Ctrl+C to stop)\n", inputDir)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			// Process on write or create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				if strings.HasSuffix(event.Name, ".csv") || strings.HasSuffix(event.Name, ".pdf") {
					fmt.Printf("\nDetected change: %s\n", event.Name)
					if err := processFile(event.Name, outputDir); err != nil {
						fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", event.Name, err)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
		}
	}
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
	watch := flag.Bool("w", false, "Watch mode: continuously monitor input directory for new or changed CSV files")
	flag.BoolVar(watch, "watch", false, "Watch mode: continuously monitor input directory for new or changed CSV files")
	flag.Parse()

	// Expand ~ in paths
	*inputDir = expandHomeDir(*inputDir)
	*outputDir = expandHomeDir(*outputDir)

	// Create output dir (e.g. ~/Desktop/20060102_output)
	now := time.Now().UTC().Format("20060102")
	timestampedOutputDir := path.Join(*outputDir, now+"_output")
	if err := os.MkdirAll(timestampedOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %q: %w", timestampedOutputDir, err)
	}

	if *watch {
		// Watch mode
		return watchMode(*inputDir, timestampedOutputDir)
	}

	// One-time processing mode
	files, err := os.ReadDir(*inputDir)
	if err != nil {
		return fmt.Errorf("failed to read input directory %q: %w", *inputDir, err)
	}

	// Track errors but continue processing
	var errors []error
	successCount := 0

	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".csv") || strings.HasSuffix(file.Name(), ".pdf")) {
			srcPath := path.Join(*inputDir, file.Name())

			if err := processFile(srcPath, timestampedOutputDir); err != nil {
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

// 2006-01-02T15:04:05
func convertDate(fromLayout, toLayout, value string) (string, error) {
	date, err := time.Parse(fromLayout, value)
	if err != nil {
		return "", err
	}
	return date.Format(toLayout), nil
}
