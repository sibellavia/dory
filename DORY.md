# Dory - Agent Memory

This project uses `dory` for persistent knowledge across sessions.

## Session Start

At the beginning of every session, read the index:
```bash
cat .dory/index.yaml
```

This gives you all lessons, decisions, patterns, and current session state.

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

### Lessons
```bash
# Bug discovery
dory learn "DHCP leases expire during long deployments" --topic networking --severity high

# Gotcha
dory learn "Must escape quotes in YAML heredocs" --topic yaml --severity normal

# With reference to related decision
dory learn "Config reload requires service restart" --topic deployment --severity high --refs D003

# Detailed lesson with body
dory learn "Race condition in queue processor" --topic backend --severity critical --body "# Race Condition

## Symptom
Jobs processed twice under high load.

## Root Cause
Worker threads not properly synchronized.

## Fix
Added mutex lock around dequeue operation."
```

### Decisions
```bash
# Architecture choice
dory decide "Use PostgreSQL over SQLite" --topic database

# With rationale in body
dory decide "Separate API and worker processes" --topic architecture --body "# Separate Processes

## Context
Single process was hitting memory limits.

## Decision
Split into API server and background worker.

## Rationale
- Independent scaling
- Crash isolation
- Clearer resource limits"

# Referencing a lesson that led to this decision
dory decide "Use static IPs for control plane" --topic networking --refs L001
```

### Patterns
```bash
# Convention
dory pattern "All API endpoints return {data, error} envelope" --domain api

# Pattern that implements a decision
dory pattern "Use context.WithTimeout for all DB queries" --domain database --refs D005

# With implementation details
dory pattern "Database migrations use timestamped filenames" --domain database --body "# Migration Naming

Format: YYYYMMDD_HHMMSS_description.sql

Example: 20260201_143052_add_users_table.sql

Run with: ./scripts/migrate.sh"
```

### Linking with Refs

Use `--refs` to connect related knowledge. This creates edges in the index for quick traversal.

```bash
# Lesson learned from a decision
dory learn "Connection pool exhaustion under load" --topic database --severity high --refs D005

# Decision based on multiple lessons
dory decide "Add circuit breaker for external APIs" --topic resilience --refs L003,L007

# Pattern that implements multiple decisions
dory pattern "Retry with exponential backoff" --domain api --refs D002,D008

# Decision that supersedes another
dory decide "Use Redis for caching instead of memcached" --topic caching --refs D004

# Lesson referencing both a decision and a pattern
dory learn "Must flush cache after schema migration" --topic database --severity critical --refs D012,P003
```

The index.yaml shows edges for quick lookup:
```yaml
edges:
  L005:
    - D005
  D010:
    - L003
    - L007
  P004:
    - D002
    - D008
```

### Session State
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
```

## Maintenance

```bash
# Edit existing item
dory edit D001

# Remove item
dory remove L003

# Clear all knowledge (fresh start)
dory reset

# Rebuild index from files
dory rebuild
```
