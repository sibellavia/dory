Dory is a _lightweight_ knowledge store that gives coding agents persistent memory across sessions.

Coding agents (Claude Code, Codex, etc.) lose context between sessions. When a new session starts, the agent doesn't remember:

- Lessons learned: bugs fixed, gotchas discovered
- Decisions made: why the architecture is the way it is
- Patterns established: conventions, or _how we do things here_
- Session state: where we left off, what's next

Dory stores structured knowledge in plain text files (YAML + Markdown) that are token-efficient (load only what's needed), human-editable, and git-friendly.

This tool was made for myself, to experiment with agents, and to use it in complex projects that require extended knowledge across different coding sessions.

Here's a quick start:

```bash
# Initialize in your project
dory init

# Record a lesson
dory learn "DHCP unreliable through firewall" --topic networking --severity high

# Record a decision
dory decide "Use static IPs for control plane" --topic networking --rationale "DHCP unreliable"

# Record a pattern
dory pattern "All infra VMs use static IPs" --domain networking

# Update session state
dory status --goal "Get cluster running" --progress "Control plane up" --next "Add workers"

# Get context at session start
cat .dory/index.yaml
```

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
| Decision | D | Architectural/technical choices | topic, rationale |
| Pattern | P | Established conventions | domain |

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

### Maintenance

```bash
dory edit <id>     # Edit in $EDITOR
dory remove <id>   # Delete (with confirmation)
dory rebuild       # Rebuild index from files
```

### Output Formats

All commands support `--json` and `--yaml` flags for machine-readable output.

## Agent Integration

Copy [DORY.md](DORY.md) to your project and include it in your agent's context (e.g., reference it in AGENTS.md/CLAUDE.md or the main file you usually give to your coding agent).

## License

MIT