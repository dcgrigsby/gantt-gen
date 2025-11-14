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
