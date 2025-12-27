package main

import (
	"flag"
	"fmt"
	"log"
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

type Parser interface {
	Name() string
	Parse(records [][]string) []YnabRecord
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

func processFile(filePath, outputDir string) error {
	fmt.Printf("Parsing %v ...", filePath)

	rawRecords, err := readCsvToRawRecords(filePath)
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	var parsed []YnabRecord
	for _, parser := range parsers {
		parsed = parser.Parse(rawRecords)
		if parsed != nil {
			fmt.Printf("Matched parser %v\n", parser.Name())
			fileName := path.Base(filePath)
			dstPath := path.Join(outputDir, parser.Name()+"_"+fileName)
			err = writeRecordsToCsv(parsed, dstPath)
			if err != nil {
				return fmt.Errorf("failed to write CSV: %w", err)
			}
			fmt.Printf("Write csv to %v\n", dstPath)
			return nil
		}
	}
	fmt.Println("No matched parser")
	return nil
}

func processDirectory(inputDir, outputDir string) error {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("failed to read input directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") {
			srcPath := path.Join(inputDir, file.Name())
			if err := processFile(srcPath, outputDir); err != nil {
				log.Printf("Error processing %s: %v", srcPath, err)
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
				if strings.HasSuffix(event.Name, ".csv") {
					fmt.Printf("\nDetected change: %s\n", event.Name)
					if err := processFile(event.Name, outputDir); err != nil {
						log.Printf("Error processing %s: %v", event.Name, err)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func main() {
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
	err := os.MkdirAll(timestampedOutputDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	if *watch {
		// Watch mode
		if err := watchMode(*inputDir, timestampedOutputDir); err != nil {
			log.Fatalf("Watch mode error: %v", err)
		}
	} else {
		// One-time processing mode
		if err := processDirectory(*inputDir, timestampedOutputDir); err != nil {
			log.Fatalf("Processing error: %v", err)
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
