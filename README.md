Dory is a knowledge store that gives coding agents persistent memory across sessions.

Modern coding agents are capable! They read code, grep codebases, and build mental models on the fly. But they forget between sessions. Even with compact mode, knowledge gets summarized and lost over time.

Dory captures what would otherwise disappear: learnings from experience you can't grep for, the rationale behind decisions that isn't in the code, project-specific gotchas learned through several sessions and hours, and session state so the next agent knows where you left off.

Dory stores structured knowledge in plain text files (YAML + Markdown) that are token-efficient (load only what's needed), human-editable, and git-friendly. This tool was made for myself, to experiment with agents, and to use it in complex projects that require extended knowledge across different coding sessions.

## Install

```bash
go install github.com/sibellavia/dory/cmd/dory@latest
```

Or download a binary from [releases](https://github.com/sibellavia/dory/releases).

## Quick start

```bash
dory init                        # Initialize in your project
cat .dory/index.yaml             # Read state at session start

# Record knowledge as you work
dory learn "Timeout must be > 30s for large uploads" --topic api --severity high
dory decide "Use Redis for session storage" --topic backend
dory pattern "All handlers return {data, error}" --domain api

# Update session state before ending
dory status --goal "Add file uploads" --progress "Endpoint done" --next "Add size limits"
```

## Why not just KNOWLEDGE.md?

For small projects, a single markdown file works fine. Dory helps when knowledge grows:

- **Queryable**: `dory recall auth` instead of scrolling through one big file
- **Incremental**: load the index, fetch full content only when needed
- **Structured**: categorized by type, topic, severity
- **Linked**: items reference each other with refs/edges
- **Session state**: goal, progress, next steps tracked automatically


## How it works

Dory creates a `.dory/` directory in your project that stores everything as plain text files. The design tries to prioritize token efficiency: agents load a lightweight index at session start, then fetch full content only when needed.

```
.dory/
├── index.yaml       # Index + session state (always loaded)
└── knowledge/       # Full content files
    ├── L001.eng     # Lessons
    ├── D001.eng     # Decisions
    └── P001.eng     # Patterns
```

Knowledge is organized into three types, each serving a distinct purpose. **Lessons** capture things you learned the hard way (bugs, gotchas, and fixes worth remembering). **Decisions** record architectural and technical choices along with their rationale, so future sessions understand why things are the way they are. **Patterns** document established conventions and "how we do things here."

| Type | Prefix | Purpose | Key Fields |
|------|--------|---------|------------|
| Lesson | L | Something learned (bugs, gotchas) | topic, severity |
| Decision | D | Architectural/technical choices | topic |
| Pattern | P | Established conventions | domain |

All types support `refs` to link related items (e.g., `--refs D001,L002`). These create edges in the index for quick traversal.

Lessons support severity levels to help agents prioritize what matters. Critical lessons are things that will break if ignored. High severity saves significant debugging time. Normal is good general knowledge, and low covers edge cases.

| Level | When to Use |
|-------|-------------|
| `critical` | Must know or things break |
| `high` | Important, saves significant time |
| `normal` | Good to know |
| `low` | Minor/edge case |

## Commands

### Adding Knowledge

```bash
# Quick lesson
dory learn "Brief description" --topic <topic> --severity <level>

# Lesson with full content
dory learn "Title" --topic <topic> --body "# Full markdown content..."

# With references to related items
dory learn "Title" --topic <topic> --refs D001,L002

# From stdin
cat content.md | dory learn "Title" --topic <topic> --body -

# Open editor (omit arguments)
dory learn --topic <topic>
```

Same flags work for `decide` and `pattern`.

### Retrieving Knowledge

```bash
cat .dory/index.yaml    # Index + state (for session start)
dory recall <topic>     # All knowledge for a topic
dory show <id>          # Full content for an item
dory list               # List all items
dory list --topic net   # Filter by topic
dory list --type lesson # Filter by type
dory list --since 2026-01-25  # Items from date onward
dory list --since 2026-01-01 --until 2026-01-31  # Date range
dory topics             # List topics with counts
```

### Session State

```bash
dory status \
  --goal "Current goal" \
  --progress "What's done" \
  --blocker "Current blocker" \
  --next "Next step 1" \
  --next "Next step 2"
```

### Import

```bash
dory import doc.md --type lesson --topic api        # Import markdown file
dory import doc.md                                   # Use frontmatter metadata
dory import lessons.md --type lesson --topic infra --split  # Split numbered items
```

### Export

```bash
dory export                      # Export all as markdown
dory export --topic architecture # Export by topic
dory export D001 D002            # Export specific items
dory export --append CLAUDE.md   # Append to file
```

### Maintenance

```bash
dory edit <id>     # Edit in $EDITOR
dory remove <id>   # Delete (with confirmation)
dory rebuild       # Rebuild index from files
dory reset         # Clear all knowledge (keep config)
dory reset --full  # Full reset (reinitialize)
```

### Output Formats

All commands support `--json` and `--yaml` flags for machine-readable output.

## Agent Integration

When you run `dory init`, it automatically appends instructions to `CLAUDE.md` and/or `AGENTS.md` if they exist in your project.

You can also manually copy [DORY.md](DORY.md) to your project and include it in your agent's context.

To export knowledge into your agent instructions:
```bash
dory export --append CLAUDE.md
```

## Use cases

### Knowledge that compounds across sessions

Session 1 - Building an API, things go wrong:

```bash
# Discovered a gotcha the hard way
dory learn "Connection pool exhausts under load" --topic database --severity critical

# Made an architecture decision based on that
dory decide "Limit pool size to 20, add queue" --topic database --refs L001

# End of session
dory status --progress "Fixed connection issues" --next "Load test again"
```

Session 2 - New agent picks up where we left off:

```bash
cat .dory/index.yaml              # Read state and index
dory recall database              # What do we know about database?
dory list --type lesson --since 2026-01-20   # Recent lessons

# Continue working... establish a pattern
dory pattern "All DB calls use context timeout" --domain database --refs D001
```

The lesson informed a decision, which became a pattern. Future sessions won't repeat the debugging.

### Reviewing recent activity

```bash
# What happened this week?
dory list --since 2026-01-27

# Critical lessons from the last month
dory list --type lesson --severity critical --since 2026-01-01

# Decisions made before a deadline
dory list --type decision --until 2026-01-31
```

### Migrating existing knowledge

```bash
# Import a markdown file
dory import docs/lessons.md --type lesson --topic api

# File with frontmatter (type/topic extracted automatically)
dory import decisions/auth.md

# Split a file with numbered items into separate entries
dory import LESSONS_LEARNED.md --type lesson --topic infra --split
```

## License

MIT