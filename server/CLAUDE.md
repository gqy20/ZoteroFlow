# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Run
- `make build` - Build the binary to `bin/zoteroflow2`
- `make run` - Build and run the binary
- `make dev` - Run directly with `go run .`
- `go run test_mineru.go` - Run MinerU integration tests

### Testing
- `make test` - Run tests with race detection
- `make test-coverage` - Run tests with coverage report (generates `coverage.html`)
- `go test -v ./...` - Verbose test output

### Code Quality
- `make fmt` - Format Go code
- `make lint` - Run golangci-lint
- `make vet` - Run go vet
- `make check` - Run fmt, lint, and test sequentially
- `make quick` - Quick fmt and vet check

### Dependencies
- `make deps` - Download dependencies and tidy go.mod
- `make mod-upgrade` - Upgrade all dependencies

## Architecture

### Core Components
1. **ZoteroDB** (`core/zotero.go`) - SQLite database accessor for Zotero
   - Reads Zotero's SQLite database in read-only mode
   - Extracts PDF metadata and file paths
   - Handles Zotero's storage system (storage:XXXXXX.pdf format)

2. **MinerUClient** (`core/mineru.go`) - HTTP client for MinerU PDF parsing API
   - Supports both single and batch processing
   - Handles file upload and result polling
   - Uses named types (FileInfo, ProcessResponse) instead of anonymous structs

3. **PDFParser** (`core/parser.go`) - Coordinates Zotero and MinerU integration
   - Manages caching of parsed results
   - Handles PDF file discovery from Zotero storage

### Configuration
Configuration is managed through environment variables and `.env` file:
- `ZOTERO_DB_PATH` - Path to Zotero's SQLite database
- `ZOTERO_DATA_DIR` - Path to Zotero's storage directory
- `MINERU_API_URL` - MinerU API endpoint
- `MINERU_TOKEN` - MinerU authentication token
- `AI_*` variables - AI model configuration

### Data Flow
1. Load configuration from `.env` and environment
2. Connect to Zotero database (read-only)
3. Create MinerU client
4. Query Zotero for PDF items
5. For each PDF: find file → upload to MinerU → retrieve parsed results
6. Cache results in `CACHE_DIR`

### Key Files
- `main.go` - Entry point with basic integration test
- `test_mineru.go` - Complete MinerU API integration tests
- `config/config.go` - Configuration management with environment variable support

## Development Notes

### Code Standards
- Use English logging (no emojis or Chinese characters)
- Prefer named types over anonymous structs for maintainability
- Follow conventional commit format: `<type>(<scope>): <description>`
- Pre-commit hooks enforce formatting and basic checks
- Pre-push hooks require 80% test coverage

### Git Hooks
The project uses automated git hooks for quality control:
- Pre-commit: formatting, go vet, basic checks
- Pre-push: full test suite with coverage
- Commit-msg: conventional commit format validation

### Testing
- No unit tests currently exist (coverage shows 0%)
- Integration tests in `test_mineru.go` test MinerU API connectivity
- Use `make test-coverage` to generate coverage reports