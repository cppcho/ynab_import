package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessFile(t *testing.T) {
	// Create a temporary output directory
	outputDir := t.TempDir()

	tests := []struct {
		name        string
		filePath    string
		shouldError bool
		shouldMatch bool
	}{
		{
			name:        "valid SMBC CSV",
			filePath:    "testdata/parsers/smbc_valid.csv",
			shouldError: false,
			shouldMatch: true,
		},
		{
			name:        "valid Rakuten CSV",
			filePath:    "testdata/parsers/rakuten_valid.csv",
			shouldError: false,
			shouldMatch: true,
		},
		{
			name:        "non-existent file",
			filePath:    "testdata/nonexistent.csv",
			shouldError: true,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processFile(tt.filePath, outputDir)
			if tt.shouldError {
				if err == nil {
					t.Errorf("processFile(%q) expected error, got nil", tt.filePath)
				}
			} else {
				if err != nil {
					t.Errorf("processFile(%q) unexpected error: %v", tt.filePath, err)
				}

				// Check if output file was created when a parser matches
				if tt.shouldMatch {
					files, err := os.ReadDir(outputDir)
					if err != nil {
						t.Fatalf("Failed to read output directory: %v", err)
					}

					found := false
					for _, file := range files {
						if strings.HasSuffix(file.Name(), ".csv") {
							found = true
							break
						}
					}

					if !found {
						t.Errorf("processFile(%q) expected output CSV file to be created", tt.filePath)
					}
				}
			}
		})
	}
}

func TestProcessDirectory(t *testing.T) {
	// Create a temporary input directory
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Copy real test CSV files to the input directory
	testFiles := []struct {
		src  string
		dest string
	}{
		{"testdata/parsers/smbc_valid.csv", "smbc_test.csv"},
		{"testdata/parsers/rakuten_valid.csv", "rakuten_test.csv"},
	}

	for _, tf := range testFiles {
		srcData, err := os.ReadFile(tf.src)
		if err != nil {
			t.Fatalf("Failed to read source file %s: %v", tf.src, err)
		}
		destPath := filepath.Join(inputDir, tf.dest)
		if err := os.WriteFile(destPath, srcData, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.dest, err)
		}
	}

	// Create a non-CSV file (should be ignored)
	txtPath := filepath.Join(inputDir, "notacsv.txt")
	if err := os.WriteFile(txtPath, []byte("text content"), 0644); err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(inputDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Process the directory
	err := processDirectory(inputDir, outputDir)
	if err != nil {
		t.Errorf("processDirectory() unexpected error: %v", err)
	}

	// Verify that output files were created
	outputFiles, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	csvCount := 0
	for _, file := range outputFiles {
		if strings.HasSuffix(file.Name(), ".csv") {
			csvCount++
		}
	}

	if csvCount != 2 {
		t.Errorf("Expected 2 output CSV files, got %d", csvCount)
	}
}

func TestProcessDirectoryNonExistent(t *testing.T) {
	outputDir := t.TempDir()
	err := processDirectory("/nonexistent/directory", outputDir)
	if err == nil {
		t.Error("processDirectory() expected error for non-existent directory, got nil")
	}
}
