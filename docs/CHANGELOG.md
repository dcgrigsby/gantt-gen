# Changelog

## [Unreleased] - 2025-11-14

### Fixed
- **Critical**: Fixed parser slice reallocation bug that caused silent data loss when parsing projects with many tasks
- **Critical**: Fixed finish-to-finish dependency logic - now correctly constrains end dates
- **Critical**: Fixed start-to-finish dependency logic - now correctly constrains end dates
- Long task names now truncate with ellipsis instead of overflowing SVG column

### Added
- Comprehensive project validation that catches:
  - Duplicate task names
  - Task names exceeding 200 characters
  - Dependencies on non-existent tasks
  - Calendar references to non-existent calendars
  - Projects exceeding 1000 tasks
- Edge case tests for parser (empty documents, Unicode, very long names)
- Edge case tests for resolver (zero duration, multiple dependencies)
- Full pipeline integration tests

### Changed
- Parser now uses indices instead of pointers for safer slice handling
- Dependency resolution tracks start and end constraints separately
- SVG renderer estimates character width more accurately for truncation

### Security
- Added input validation limits to prevent DoS attacks
- Task name length limits prevent excessive memory usage
