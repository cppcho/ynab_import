# ynab_import

A Go-based CSV conversion tool that transforms bank and credit card transaction exports from various Japanese financial institutions into YNAB (You Need A Budget) compatible format.

## Features

- **11 Financial Institution Support** - Supports major Japanese banks, credit cards, and transit IC cards
- **Automatic Encoding Detection** - Handles both UTF-8 and Shift_JIS encoded CSVs
- **Batch Processing** - Processes all CSV files in a directory at once
- **Watch Mode** - Continuously monitor directory for new or changed CSV files
- **Automatic Parser Matching** - Identifies the correct parser based on CSV headers
- **Timestamped Output** - Organizes converted files in dated directories
- **YNAB-Ready Format** - Outputs standardized CSV format for direct YNAB import

## Supported Financial Institutions

| Institution | Japanese Name | Type | Format |
|-------------|---------------|------|--------|
| SMBC Bank | 三井住友銀行 | Bank | CSV |
| Rakuten Bank | 楽天銀行 | Bank | CSV |
| SBI Bank | 住信SBIネット銀行 | Bank | CSV |
| SBI Shinsei Bank | SBI新生銀行 | Bank | CSV |
| SMBC Card | 三井住友カード | Credit Card (2 formats) | CSV |
| Rakuten Card | 楽天カード | Credit Card | CSV |
| EPOS Card | エポスカード | Credit Card | CSV |
| VIEW Card | ビューカード | Credit Card | CSV |
| Saison Card | セゾンカード | Credit Card | CSV |
| Mobile Suica | モバイルSuica | Transit IC Card | PDF |

## Requirements

- Go 1.25 or later
- `pdftotext` (from poppler-utils) - Required for PDF parsing (Mobile Suica)
  - macOS: `brew install poppler`
  - Linux: `apt-get install poppler-utils`

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd ynab_import

# Download dependencies
go mod download

# Build the binary
make build
```

The compiled binary will be available at `./bin/ynab_import`.

## Quick Start

```bash
# Place your bank/credit card CSV exports and PDFs in ~/Downloads
# Run the converter
./bin/ynab_import

# Find converted files in ~/Desktop/YYYYMMDD_output/
```

## Usage

### Basic Usage

```bash
./bin/ynab_import
```

By default, the tool:
- Reads CSV and PDF files from `~/Downloads`
- Outputs converted files to `~/Desktop/YYYYMMDD_output/`

### Custom Directories

Use command-line flags to specify custom input/output directories:

```bash
./bin/ynab_import -input ~/Documents/bank_exports -output ~/Documents/ynab_ready
```

### Environment Variables

You can also configure directories using environment variables:

```bash
export CSV_DIR_IN=~/Documents/bank_exports
export CSV_DIR=~/Documents/ynab_ready
./bin/ynab_import
```

### Watch Mode

Watch mode allows the tool to continuously monitor the input directory for new or changed CSV files and automatically convert them:

```bash
# Start watch mode
./bin/ynab_import -w

# Or with custom directories
./bin/ynab_import -w -input ~/Documents/bank_exports -output ~/Documents/ynab_ready

# Alternative flag syntax
./bin/ynab_import --watch
```

In watch mode, the tool will:
1. Process all existing CSV and PDF files in the input directory on startup
2. Continue running and monitor the directory for changes
3. Automatically process any new CSV or PDF files added to the directory
4. Automatically re-process files when they are modified
5. Press `Ctrl+C` to stop watching

This is useful for scenarios like:
- **Automated workflows**: Set up watch mode to run as a background service
- **Continuous imports**: Automatically convert files as they're downloaded
- **Real-time processing**: Process transactions as soon as bank exports are saved

### Command-Line Flags

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `-input` | `CSV_DIR_IN` | `~/Downloads` | Directory containing input CSV and PDF files |
| `-output` | `CSV_DIR` | `~/Desktop` | Base directory for output files |
| `-w`, `--watch` | - | `false` | Watch mode: continuously monitor input directory for new or changed files |

## Output Format

The tool converts all transactions to YNAB's standard CSV format:

```csv
Date,Payee,Memo,Amount
2024-01-15,Grocery Store,Shopping,3500
2024-01-16,Restaurant,Dinner,-4200
```

- **Date**: YYYY-MM-DD format
- **Payee**: Merchant or transaction description
- **Memo**: Additional transaction details
- **Amount**: Numeric amount (positive for income, negative for expenses)

### Output Directory Structure

```
~/Desktop/
└── 20241227_output/
    ├── smbc_transactions.csv
    ├── rakuten_card_december.csv
    └── epos_2024.csv
```

Output files are named: `{parser_name}_{original_filename}`

## Development

### Building

```bash
# Build binary to ./bin/ynab_import
make build

# Run directly without building
make run
# or
go run .
```

### Testing

```bash
# Run all tests
make test

# Generate coverage report (opens in browser)
make coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint
```

### Development Workflow

Before completing any code change:

1. Update or add tests
2. Run `make test` - all tests must pass
3. Run `make fmt` - format code
4. Run `make lint` - fix any issues
5. Run `make build` - verify successful compilation

See [CLAUDE.md](CLAUDE.md) for detailed development guidelines.

## Adding New Parsers

To add support for a new financial institution:

1. **Create parser file**: `institution.go`
2. **Implement Parser interface**:
   ```go
   type YourParser struct{}

   func (p YourParser) Name() string {
       return "InstitutionName"
   }

   func (p YourParser) Parse(records [][]string) []YnabRecord {
       // Check if CSV matches your format
       if !matchesMyFormat(records) {
           return nil  // Not my format
       }

       // Parse and convert records
       // ...
       return ynabRecords
   }
   ```
3. **Register parser**: Add to `parsers` slice in `main.go:27`
4. **Create test file**: `institution_test.go` with comprehensive tests
5. **Add test data**: Sample CSV in `testdata/parsers/`
6. **Verify quality**: Run `make test`, `make fmt`, `make lint`, `make build`

Key utilities available:
- `flipSign(amount)` - Reverse transaction sign
- `convertDate(date, fromLayout)` - Convert date to YYYY-MM-DD

See existing parsers (e.g., `smbc.go`, `rakuten.go`) for examples.

## Project Structure

```
ynab_import/
├── main.go              # Core application logic and parser registry
├── csv.go               # CSV reading/writing with encoding detection
├── smbc.go              # SMBC Bank parser
├── rakuten.go           # Rakuten Bank parser
├── epos.go              # EPOS Card parser
├── sbi.go               # SBI Bank parser
├── shinsei.go           # SBI Shinsei Bank parser
├── rakuten_card.go      # Rakuten Card parser
├── smbc_card.go         # SMBC Card parser (format 1)
├── smbc_card2.go        # SMBC Card parser (format 2)
├── view.go              # VIEW Card parser
├── saison.go            # Saison Card parser
├── suica.go             # Mobile Suica parser (PDF)
├── *_test.go            # Test files
├── testdata/            # Test CSV and PDF samples
├── Makefile             # Build automation
├── go.mod               # Go module definition
├── CLAUDE.md            # Detailed development guidelines
└── TODO.md              # Known issues and planned improvements
```

## How It Works

1. **Scan Input Directory** - Finds all CSV and PDF files in the input directory
2. **Try Parsers** - For each file, tries each registered parser in sequence
3. **Match Format** - Parsers check file headers/content to identify their format
4. **Parse & Convert** - Matching parser converts records to YNAB format
5. **Handle Encoding** - Automatically detects and converts Shift_JIS to UTF-8 (for CSVs)
6. **Extract PDF Text** - Extracts text from PDF files (for transit IC cards like Suica)
7. **Write Output** - Saves converted CSV to timestamped output directory
