# Software Development Project

## Calendar: US-2024

| Type | Value |
|------|-------|
| Default | true |
| Weekends | Sat, Sun |
| Holiday | 2024-01-01 |
| Holiday | 2024-07-04 |
| Holiday | 2024-12-25 |

## Design Phase

| Property | Value |
|----------|-------|
| Start | 2024-01-02 |
| Duration | 10d |
| Link | https://jira.example.com/PROJ-101 |

## Implementation

| Property | Value |
|----------|-------|
| Duration | 15d |
| Link | https://jira.example.com/PROJ-102 |

| Depends On | Type |
|------------|------|
| Design Phase | finish-to-start |

### Backend Development

| Property | Value |
|----------|-------|
| Duration | 10d |
| Link | https://jira.example.com/PROJ-103 |

| Depends On | Type |
|------------|------|
| Implementation | start-to-start |

### Frontend Development

| Property | Value |
|----------|-------|
| Duration | 12d |
| Link | https://jira.example.com/PROJ-104 |

| Depends On | Type |
|------------|------|
| Implementation | start-to-start |

## Testing

| Property | Value |
|----------|-------|
| Duration | 5d |
| Link | https://jira.example.com/PROJ-105 |

| Depends On | Type |
|------------|------|
| Backend Development | finish-to-start |
| Frontend Development | finish-to-start |

**Code Complete Milestone**

| Property | Value |
|----------|-------|
| Link | https://jira.example.com/PROJ-200 |

| Depends On | Type |
|------------|------|
| Testing | finish-to-start |

## Deployment

| Property | Value |
|----------|-------|
| Duration | 2d |
| Link | https://jira.example.com/PROJ-106 |

| Depends On | Type |
|------------|------|
| Code Complete Milestone | finish-to-start |

**Launch Milestone**

| Property | Value |
|----------|-------|
| Date | 2024-03-01 |
| Link | https://jira.example.com/PROJ-201 |
