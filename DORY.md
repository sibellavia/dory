# Dory - Agent Memory

This project uses dory for persistent knowledge across sessions.

## Storage Format

Dory uses a single-file format:
```
.dory/
├── index.yaml      # Project metadata, state, deleted IDs
└── knowledge.dory  # Append-only content (YAML entries separated by ===)
```

## Session Start

At the beginning of every session, read the index:
```bash
cat .dory/index.yaml
```

This gives you project info, current state, and session context.

## Session End

Before ending a session, update the state so the next session knows where you left off:
```bash
dory status --goal "Current goal" --progress "What's done" --next "Next step"
```

This is important for session continuity. The next agent will see this state when they run `cat .dory/index.yaml`.

## Command Help

Every command is self-documenting:
```bash
dory <command> --help
```

For example: `dory import --help`, `dory learn --help`, `dory status --help`

## When to Record Knowledge

**Record a lesson when:**
- You discovered a bug or gotcha worth remembering
- Something didn't work as expected
- You found a fix that wasn't obvious

**Record a decision when:**
- You made an architectural or technical choice
- You chose between alternatives
- Future sessions need to know "why it's this way"

**Record a pattern when:**
- You established a convention
- There's a "right way" to do something in this project

## Examples

### Quick entries (no body)
```bash
# Lesson - something learned
dory learn "DHCP leases expire during long deployments" --topic networking --severity high

# Decision - a choice made
dory decide "Use PostgreSQL over SQLite" --topic database

# Pattern - a convention
dory pattern "All API endpoints return {data, error} envelope" --domain api

# With references
dory learn "Config reload requires restart" --topic deployment --refs D003
```

### Entries with body content

Use heredoc with `--body -` for natural markdown without escaping:

```bash
# Lesson with detailed body
cat << 'EOF' | dory learn "Race condition in queue processor" --topic backend --severity critical --body -
# Race Condition

## Symptom
Jobs processed twice under high load.

## Root Cause
Worker threads not properly synchronized.

## Fix
Added mutex lock around dequeue operation.
EOF

# Decision with rationale
cat << 'EOF' | dory decide "Separate API and worker processes" --topic architecture --body -
# Separate Processes

## Context
Single process was hitting memory limits.

## Decision
Split into API server and background worker.

## Rationale
- Independent scaling
- Crash isolation
- Clearer resource limits
EOF

# Pattern with implementation details
cat << 'EOF' | dory pattern "Database migrations use timestamped filenames" --domain database --body -
# Migration Naming

Format: YYYYMMDD_HHMMSS_description.sql

Example: 20260201_143052_add_users_table.sql

Run with: ./scripts/migrate.sh
EOF
```

### Linking with Refs

Use `--refs` to connect related knowledge:

```bash
# Lesson learned from a decision
dory learn "Connection pool exhaustion under load" --topic database --severity high --refs D005

# Decision based on multiple lessons
dory decide "Add circuit breaker for external APIs" --topic resilience --refs L003,L007

# Pattern that implements multiple decisions
dory pattern "Retry with exponential backoff" --domain api --refs D002,D008
```

### Session State

Update state before ending your session so the next agent knows where you left off:

```bash
# Update progress
dory status --goal "Implement user auth" --progress "Login endpoint done" --next "Add JWT validation"

# Multiple next steps
dory status \
  --progress "Fixed database connection pooling" \
  --next "Add connection timeout config" \
  --next "Write tests for edge cases" \
  --next "Update documentation"

# Note a blocker
dory status --blocker "Waiting for API credentials from client"
```

## Retrieval

```bash
# Get full content of an item
dory show D001

# Get all knowledge for a topic
dory recall networking

# List all items
dory list

# Filter by type
dory list --type lesson

# Filter by topic
dory list --topic database

# Filter by severity
dory list --severity critical
```

### Date Range Filtering

Filter items by creation date:

```bash
# Items created on or after a date
dory list --since 2026-01-25

# Items created on or before a date
dory list --until 2026-01-31

# Items within a date range
dory list --since 2026-01-01 --until 2026-01-31

# Combine with other filters
dory list --topic database --since 2026-01-15
dory list --type lesson --severity high --since 2026-01-01
```

## Importing Existing Knowledge

Migrate existing markdown files (like LESSONS_LEARNED.md, DECISIONS.md, etc.) into dory.

### Quick Reference

| File Format | Command |
|-------------|---------|
| Single document | `dory import FILE.md --type lesson --topic <topic>` |
| Has YAML frontmatter | `dory import FILE.md` (metadata extracted automatically) |
| Numbered list (1), 2)...) | `dory import FILE.md --type lesson --topic <topic> --split` |

### Import as Single Entry

Import the entire file as one knowledge item:

```bash
dory import LESSONS_LEARNED.md --type lesson --topic infrastructure
dory import ARCHITECTURE.md --type decision --topic backend
dory import CONVENTIONS.md --type pattern --topic api
```

### Import with Severity

For lessons, specify severity:

```bash
dory import critical-bugs.md --type lesson --topic backend --severity critical
dory import gotchas.md --type lesson --topic api --severity high
```

### Split Numbered Lists

If your file has numbered items, split them into separate entries:

```markdown
# Lessons Learned

1) SSH access failed until correct key used
- Symptom: SSH denied even after adding a key.
- Fix: Use the correct key file.

2) Config reload requires restart
- Symptom: Changes not applied.
- Fix: Restart the service after config changes.
```

```bash
dory import LESSONS_LEARNED.md --type lesson --topic infra --split
# Creates L001: SSH access failed until correct key used
# Creates L002: Config reload requires restart
```

### Frontmatter Support

If your file has YAML frontmatter, dory extracts metadata automatically:

```markdown
---
type: decision
topic: caching
severity: high
---
# Use Redis for session storage

We need fast session lookups...
```

```bash
dory import decisions/redis.md  # type and topic from frontmatter
```

## Maintenance

```bash
# Edit existing item in $EDITOR
dory edit D001

# Remove item (marks as deleted)
dory remove L003

# Compact - physically remove deleted items and reclaim space
dory compact

# Clear all knowledge (fresh start)
dory reset

# Full reset (reinitialize completely)
dory reset --full
```

## Export

Export knowledge for inclusion in other files:

```bash
# Export all knowledge as markdown
dory export

# Export by topic
dory export --topic architecture

# Export specific items
dory export D001 D002 L001

# Append to a file
dory export --topic api --append CLAUDE.md
```
