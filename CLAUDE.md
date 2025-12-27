# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CSV conversion tool that transforms bank and credit card transaction CSVs from various Japanese financial institutions into YNAB (You Need A Budget) compatible format.

## Commands

**Build and run:**
```bash
go build
./ynab_import
```

**Run directly:**
```bash
go run .
```

**Install dependencies:**
```bash
go mod download
```

**Update dependencies:**
```bash
go mod tidy
```

## Development Guidelines

**CRITICAL: Code Quality Requirements**

Before completing ANY code change, you MUST:

1. **Update Tests**: Write or update tests for all code changes
   ```bash
   make test
   ```
   All tests must pass before considering the change complete.

2. **Format Code**: Ensure code is properly formatted
   ```bash
   make fmt
   ```

3. **Run Linter**: Ensure no linting errors
   ```bash
   make lint
   ```

4. **Verify Build**: Ensure the project builds successfully
   ```bash
   make build
   ```

5. **Update Documentation**: Update documentation when making changes:
   - **README.md** (user-facing): Update when adding new features, supported institutions, or changing usage
   - **CLAUDE.md** (developer-facing): Update when changing architecture, design patterns, or development workflows

**Workflow for Every Change:**
1. Make code changes
2. Update relevant tests (or add new tests)
3. Run `make test` and fix any failures
4. Run `make fmt` to format code
5. Run `make lint` and fix any issues
6. Run `make build` to verify successful compilation
7. Update documentation (README.md and/or CLAUDE.md) if necessary
8. Only then is the change complete

## Documentation

This project maintains two documentation files:

**README.md** - User-facing documentation:
- Installation and setup instructions
- Usage examples and command-line flags
- List of supported financial institutions
- Quick start guide for end users
- Output format specification

**CLAUDE.md** (this file) - Developer-facing documentation:
- Architecture and design patterns
- Code quality requirements and workflow
- Parser plugin pattern details
- Guidelines for adding new parsers
- Internal implementation details

**When to update each:**
- Update **README.md** when: Adding supported institutions, changing CLI flags, modifying usage patterns, adding features visible to end users
- Update **CLAUDE.md** when: Changing architecture, adding internal utilities, modifying parser patterns, updating development workflow

## Architecture

### Parser Plugin Pattern

The codebase uses a plugin-style architecture where each financial institution has its own parser implementation:

- **Parser Interface** (main.go:22-25): Defines `Name()` and `Parse()` methods
- **Parser Registry** (main.go:27): All parsers are registered in a slice and tried sequentially
- **Parser Implementations**: Each `*_*.go` file (smbc.go, rakuten.go, epos.go, etc.) implements the Parser interface

### Processing Flow

1. **Input**: Reads all CSV files from `CSV_DIR_IN` (/Users/cppcho/Downloads)
2. **Matching**: For each CSV, tries each parser in sequence until one matches
3. **Parsing**: Matching is done by checking CSV headers using `reflect.DeepEqual`
4. **Output**: Writes converted CSVs to timestamped directory in `CSV_DIR` (~/Desktop/YYYYMMDD_output)

### Key Components

**main.go**:
- `YnabRecord` struct: Standard output format (date, payee, memo, amount)
- `flipSign()`: Utility to reverse transaction sign (income/expense)
- `convertDate()`: Date format conversion between layouts
- Main loop: Directory scanning, parser matching, output generation

**csv.go**:
- `readCsvToRawRecords()`: Handles Shift_JIS encoding detection and conversion (common in Japanese bank CSVs)
- `writeRecordsToCsv()`: Outputs YNAB-compatible format with headers

**Parser files** (smbc.go, rakuten.go, etc.):
- Each parser identifies its CSV format by checking headers
- Returns `nil` if CSV doesn't match (signals "not my format")
- Returns `[]YnabRecord` if CSV matches and is successfully parsed

### Adding New Parsers

1. Create new file `<institution>.go`
2. Implement Parser interface with unique header check
3. Add parser to registry in main.go:27
4. Handle institution-specific date formats and column mappings
5. Use `flipSign()` if amount signs need reversing
6. **Create test file** `<institution>_test.go` with comprehensive tests
7. **Add test data**: Sample CSV in `testdata/parsers/<institution>_valid.csv`
8. **Run quality checks**: `make test`, `make fmt`, `make lint`, `make build` (all must pass)
9. **Update README.md**: Add the institution to the "Supported Financial Institutions" table
10. **Update CLAUDE.md**: Only if the parser introduces new patterns or special handling

### Important Details

- **Encoding**: Japanese CSVs often use Shift_JIS encoding, handled automatically in csv.go:44-65
- **Date Formats**: Different institutions use different formats (2006/1/2 vs 20060102). Use `convertDate()` to normalize to YYYY-MM-DD
- **Output Format**: YNAB expects columns: Date, Payee, Memo, Amount
- **Sign Convention**: Use `flipSign()` to convert debit/credit conventions between institutions and YNAB
- **Hardcoded Paths**: CSV_DIR_IN and CSV_DIR in main.go are currently hardcoded to user-specific paths
