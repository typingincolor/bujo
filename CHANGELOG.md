# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### Features
- **Quarterly Habit View**: Added 90-day quarterly view for habit tracker. Press `w` in habit view to cycle between Week (7 days) → Month (30 days) → Quarter (90 days) views (#151)
- **Search Context**: Search results now display ancestry chains showing parent hierarchy (e.g., "↳ Project A > Phase 1") for better context. Long chains are automatically truncated (#138)
- **Undo Functionality**: Press `u` in TUI to undo the last mark done/undone operation. Supports one level of undo (#126)
- **URL Opening**: Press `o` in TUI to open URLs from selected entry in default browser. Cross-platform support for macOS, Linux, and Windows (#145)
- **Location Setting via Command Palette**: Added "Set Location" command to command palette (`Ctrl+P`) for quick location changes from TUI (#153)
- **Parent Flag Support**: CLI `add` command now fully supports `--parent` flag for adding child entries directly (#139)

#### Technical Improvements
- **BackupService Refactoring**: Extracted `BackupRepository` interface following hexagonal architecture principles. Service is now stateless with configuration passed as parameters (#115)
- **Code Deduplication**: Created shared `internal/dateutil` package eliminating duplicate date parsing logic between CLI and TUI (#114)
- **Domain Immutability**: `Goal.MarkDone()` and `Goal.MarkActive()` now use value receivers and return new instances instead of mutating (#112)
- **Test Coverage**: Increased domain layer test coverage from 91.2% to 99.3% with comprehensive tests for parsers, search, goals, and habits (#118)

### Fixed
- **Stdout/Stderr Separation**: Fixed `goal` and `deleted` commands to output data to stdout instead of stderr per 12-factor app principles (#117)
- **Undo Race Condition**: Fixed race condition where undo state wasn't cleared on operation failure
- **Static Analysis**: Fixed if-else chains flagged by staticcheck (converted to switch statements)
- **Code Cleanup**: Removed unused `formatSummaryPeriod()` and `navigateSummaryPeriod()` functions identified by golangci-lint

### Changed
- **Magic Numbers**: Extracted habit view day counts (7, 30, 90) as named constants (`HabitDaysWeek`, `HabitDaysMonth`, `HabitDaysQuarter`)
- **Undo Operations**: Removed unused `UndoOpDelete`, `UndoOpEdit`, and `UndoOpAdd` enum values to maintain code clarity
- **Search Ancestry Truncation**: Limited ancestry chains to 3 ancestors with 40-character maximum per entry to prevent UI overflow

### Internal
- **Comprehensive Test Suite**: Added 259 lines of tests across dateutil, domain, and TUI layers
- **Date Parsing Tests**: Added full test coverage for `ParsePast()` and `ParseFuture()` functions
- **URL Extraction Tests**: Added 11 test cases for URL detection and extraction
- **Immutability Tests**: Verified that domain object mutations return new instances

## [Previous Releases]

See [GitHub Releases](https://github.com/typingincolor/bujo/releases) for earlier versions.
