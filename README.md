# Gantt Chart Generator

A Go CLI tool that generates beautiful Gantt charts from markdown files.

## Features

- 📝 Write project plans in readable markdown
- 🔗 Link tasks to external tools (Jira, GitHub issues, etc.)
- 📅 Flexible date formats (ISO 8601 or natural language)
- 🔄 Dependency management (finish-to-start, start-to-start, etc.)
- 📆 Calendar support (weekends, holidays, business days)
- 🎨 SVG output format
- ⚡ Fast and standalone (no dependencies at runtime)

## Installation

```bash
go install github.com/yourusername/gantt-gen@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/gantt-gen.git
cd gantt-gen
go build
```

## Usage

```bash
gantt-gen input.md output.svg
```

## Validation

The tool validates your input and provides helpful error messages:

```bash
$ gantt-gen project.md output.svg
Validation error: duplicate task name: Implementation
```

Common validation errors:
- Duplicate task names (must be unique)
- Task names over 200 characters
- Dependencies on non-existent tasks
- Calendar references to non-existent calendars

## Limitations

- Maximum 1000 tasks per project
- Maximum 200 characters per task name (longer names are truncated in output)
- Dependency logic uses simplified constraints (see docs/format.md)

## Markdown Format

See [docs/format.md](docs/format.md) for complete format specification.

### Quick Example

```markdown
# My Project

## Design Phase

| Property | Value |
|----------|-------|
| Start | 2024-01-01 |
| Duration | 5d |
| Link | https://jira.com/PROJ-123 |

## Implementation

| Property | Value |
|----------|-------|
| Duration | 10d |

| Depends On | Type |
|------------|------|
| Design Phase | finish-to-start |
```

## Examples

See `examples/sample-project.md` for a complete example project.

Generate the example:

```bash
gantt-gen examples/sample-project.md examples/output.svg
```

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build
```

## License

MIT License - see LICENSE file for details.
