# Index

```yaml
version: 1
project: dory
description: Knowledge memory CLI for coding agents
lessons:
    L001:
        oneliner: Rebuild loses oneliners if not stored in frontmatter
        topic: architecture
        severity: high
        created: 2026-01-31T21:49:52.87501+01:00
    L002:
        oneliner: Use /dory not dory in .gitignore to avoid ignoring cmd/dory
        topic: git
        severity: normal
        created: 2026-01-31T21:49:52.981773+01:00
decisions:
    D001:
        oneliner: Use Cobra for CLI framework
        topic: architecture
        created: 2026-01-31T21:49:45.987925+01:00
    D002:
        oneliner: Store oneliner in .eng frontmatter
        topic: architecture
        created: 2026-01-31T21:49:46.081821+01:00
    D003:
        oneliner: YAML frontmatter + Markdown body for .eng files
        topic: file-format
        created: 2026-01-31T21:49:46.194131+01:00
    D004:
        oneliner: Separate index.yaml from knowledge files
        topic: architecture
        created: 2026-01-31T21:49:46.288691+01:00
    D005:
        oneliner: Name the tool Dory after Finding Nemo character
        topic: branding
        created: 2026-01-31T21:49:46.479352+01:00
patterns:
    P001:
        oneliner: All commands support --json and --yaml output flags
        domain: cli
        created: 2026-01-31T21:49:55.464717+01:00
    P002:
        oneliner: Use --body flag or stdin for programmatic content input
        domain: cli
        created: 2026-01-31T21:49:55.574983+01:00
topics:
    - architecture
    - file-format
    - branding
    - git
    - cli
```

# State

```yaml
last_updated: 2026-01-31T21:53:20.581317+01:00
goal: Explore token-efficient formats and relation structures
progress: Core implementation complete, renamed to Dory
next:
    - Prototype refs field for lightweight linking
    - Add edges section to index.yaml for causal chains
    - Consider compact mode for agent consumption
    - Add tests
```
