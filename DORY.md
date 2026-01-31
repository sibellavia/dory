# Dory - Agent Memory

This project uses dory for persistent knowledge across sessions.

## Session Start

Run `dory brief` to load project knowledge and current state.

## Recording Knowledge

### Lessons (something you learned)
```bash
dory learn "<what you learned>" --topic <topic> --severity <critical|high|normal|low>
```

### Decisions (architectural/technical choices)
```bash
dory decide "<decision>" --topic <topic> --rationale "<why>"
```

### Patterns (established conventions)
```bash
dory pattern "<pattern>" --domain <domain>
```

## Writing Detailed Content

Use `--body` flag for full markdown:
```bash
dory learn "Title" --topic <topic> --severity <level> --body "# Full markdown content..."
```

Or pipe content:
```bash
cat <<'EOF' | dory learn "Title" --topic <topic> --body -
# Detailed Lesson

## Symptom
What you observed...

## Root Cause
Why it happened...

## Fix
How to solve it...
EOF
```

## Session End

Update status before ending:
```bash
dory status \
  --goal "<current goal>" \
  --progress "<what's done>" \
  --blocker "<current blocker, if any>" \
  --next "<next step>"
```

## Retrieval Commands

| Command | Purpose |
|---------|---------|
| `dory brief` | Get index + state (session bootstrap) |
| `dory recall <topic>` | Get all knowledge for a topic |
| `dory show <id>` | Get full content for an item |
| `dory list` | List all items |
| `dory list --topic <t>` | Filter by topic |
| `dory topics` | List topics with counts |

## Output Formats

Add `--json` or `--yaml` for machine-readable output:
```bash
dory brief --yaml
dory list --json
```
