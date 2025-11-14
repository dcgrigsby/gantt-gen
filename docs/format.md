# Gantt Chart Markdown Format

This document describes the markdown format for gantt-gen.

## Overview

The format uses standard markdown with special table structures to define tasks, dependencies, and calendars.

## Structure

### Project Name (H1)

```markdown
# Project Name
```

The first H1 heading becomes the project name.

### Tasks (H2, H3, H4, etc.)

```markdown
## Task Name
### Subtask Name
```

Heading levels 2 and below create tasks. The level determines visual hierarchy and color coding.

### Milestones (Bold Text)

```markdown
**Milestone Name**
```

Bold text creates zero-duration milestone markers.

### Property Tables

Define task properties:

```markdown
| Property | Value |
|----------|-------|
| Start | 2024-01-01 |
| Duration | 5d |
| Link | https://example.com |
| Calendar | US-2024 |
```

**Properties:**
- `Start`: Explicit start date (any common format)
- `End`: Explicit end date (for date ranges)
- `Date`: Explicit date (for milestones)
- `Duration`: Duration (e.g., `5d` for days, `2w` for weeks)
- `Link`: URL to external resource (e.g., Jira ticket)
- `Calendar`: Calendar name to use for this task

### Dependency Tables

Define task dependencies:

```markdown
| Depends On | Type |
|------------|------|
| Task A | finish-to-start |
| Task B | start-to-start |
```

**Dependency Types:**
- `finish-to-start`: Task starts when dependency finishes (default)
- `start-to-start`: Task starts when dependency starts
- `finish-to-finish`: Task finishes when dependency finishes
- `start-to-finish`: Task finishes when dependency starts

### Calendar Tables

Define working calendars:

```markdown
## Calendar: US-2024

| Type | Value |
|------|-------|
| Default | true |
| Weekends | Sat, Sun |
| Holiday | 2024-01-01 |
| Holiday | 2024-07-04 |
```

**Calendar Properties:**
- `Default`: Set to `true` to make this the default calendar
- `Weekends`: Comma-separated weekend days
- `Holiday`: Holiday dates (can have multiple rows)

## Timing Rules

Tasks can specify timing in three ways:

1. **Explicit dates**: Use `Start` + `Duration` or `Start` + `End`
2. **Dependency-based**: Use `Depends On` + `Duration`
3. **Milestone date**: Use `Date` for fixed milestones

## Examples

See `examples/sample-project.md` for a complete example.
