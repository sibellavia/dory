# Dory

Persistent memory for coding agents. Store lessons, decisions, and conventions across sessions.

## Quick Start

```bash
cat .dory/index.yaml         # Session start: read current state
dory context                  # Or use this for formatted context
# ... work ...
dory context --goal "X" --progress "Y" --next "Z"  # Session end: save state
```

## Commands

### Create

```bash
dory create "Title" --tag <tag>                    # Lesson (default)
dory create "Title" --kind decision --tag <tag>    # Decision
dory create "Title" --kind convention --tag <tag>  # Convention
dory create "Title" --tag <tag> --severity high    # With severity (lessons only)
dory create "Title" --tag <tag> --refs L-xxx,D-yyy # With references

# With body
cat << 'EOF' | dory create "Title" --tag api --body -
# Details
Content here.
EOF
```

### List

```bash
dory list                         # All items
dory list --tag api               # By tag
dory list --type lesson           # By type
dory list --severity critical     # By severity
dory list --since 2026-01-01      # By date
dory list --tags                  # Show all tags with counts
```

### Show

```bash
dory show <id>              # Content
dory show <id> --refs       # Content + relationships
dory show <id> --expand     # Content + connected items
dory show <id> --graph      # Visual graph
```

### Edit

```bash
# Agent mode
dory edit <id> --tag new-tag --severity high
echo 'tag: api' | dory edit <id> --apply -
dory edit <id> --patch '{"severity":"critical"}'

# Human mode (opens $EDITOR)
dory edit <id>
```

### Context (Session State)

```bash
# Read
dory context                      # Current state + recent items
dory context --tag auth           # Include auth-related items
dory context --full               # Include all items

# Write (updates state, returns context)
dory context --goal "Add auth" --progress "50%" --next "Add logout"
dory context --blocker "Waiting for API keys"
```

### Other

```bash
dory remove <id> --force    # Delete item
dory import file.md --type lesson --tag api
dory export --tag api
dory compact                # Reclaim space from deleted items
```

## Types

| Type | Prefix | Use When |
|------|--------|----------|
| lesson | L- | Learned something (bug, gotcha, fix) |
| decision | D- | Made a choice (architecture, trade-off) |
| convention | P- | Established a standard (pattern, convention) |

## Severity (lessons only)

`critical` > `high` > `normal` > `low`

## Output Formats

```bash
dory list --json            # JSON
dory list --yaml            # YAML
dory --agent list           # Agent mode (YAML, no prompts)
```

## Storage

```
.dory/
├── index.yaml      # Metadata, state, snapshot
└── knowledge.dory  # Append-only entries
```
