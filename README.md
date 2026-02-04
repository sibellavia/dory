<h1 align="center">🐠</h1>

<b><p align="center">A knowledge store for coding agents that persists across sessions.</center></p></b>

Modern coding agents are capable: they read code, grep what they need, and build knowledge on the fly. However, they forget between sessions. Even with /compact, knowledge gets summarized, some details are lost and there's no more trace of causality, historical evolution, fixes learned the hard way, and other aspects that may be retained over the time.

Dory captures what would otherwise disappear: learnings from experience you can't grep for, the rationale behind decisions that isn't in the code, project-specific details learned through several sessions and hours, and session state so the next agent knows where you left off. To do so, Dory stores structured knowledge in Doryfile: an append-only storage format consisting of two files: knowledge.dory (the event log) and index.yaml (a snapshot with items and a log offset). 

One may ask: why should I use Dory instead of writing one single markdown file? First of all, I built Dory to meet my needs... and to have fun and play with agents! Besides of that: for small, short-lived projects a single markdown file works fine. Dory helps when knowledge grows. Doryfile is queryable, this means you can just do `dory list --tag auth` instead of scrolling through one big file. You can create relations and reference items with each other with `refs`, with the possibility of even having some fancy visualizations directly on the terminal. Agents can load the index and fetch full content only when needed, categorizing type, topic, severity for each item. And should you need something specific, Dory offers a minimalist yet powerful plugin system.

I made Dory for myself, hope you enjoy it as much as I do!

## Quick Start

```bash
dory init                 # once per project

# During work, capture what matters:

dory create "Why we chose Postgres over SQLite" --kind decision --tag database
dory create "Auth tokens expire silently—always check error response" --tag auth

# At session end, leave a note for the next agent:

dory context --goal "Add user auth" --progress "Login done" --next "Implement logout"
```

For the full command reference, see [DORY.md](./content/DORY.md).

## What you can do

Query by tag instead of scrolling through a giant file:

```bash
$ dory list --tag architecture
D-01KGJ5XFHD1R0B  decision  architecture  Complete Phase 1 plugin lifecycle...
D-01KGJ5XFHD3T05  decision  architecture  Adopt single-file storage format...
D-01KGJ5XFHD5PQ5  decision  architecture  Defer graph export formats...
```

Visualize how decisions connect:

```bash
$ dory show D-01KGJ5... --graph

     ┌────────────┐  ┌────────────┐
     │D-01KGJ5XFHD│  │D-01KGJ5XFHD│
     └────────────┘  └────────────┘
           │              │
           └──────┬───────┘
                  ▼
 ╔══════════════════════════════════════════════════╗
 ║           D-01KGJ5XFHD1R0BDD9NVCARY3BH           ║
 ║  Complete Phase 1 plugin lifecycle and proto...  ║
 ╚══════════════════════════════════════════════════╝
                      │
         ┌────────────┼────────────┐
         ▼            ▼            ▼
   ┌────────────┐┌────────────┐┌────────────┐
   │D-01KGJ5XFHD││D-01KGJ5XFHD││D-01KGJ5XFHD│
   └────────────┘└────────────┘└────────────┘
```

Get a session briefing with what matters:

```bash
$ dory context
╭─────────────────────────────────────────╮
│  CONTEXT: my-project                    │
╰─────────────────────────────────────────╯

SESSION STATE
  Goal: Add user authentication
  Progress: Login endpoint done, JWT working
  Next: Implement logout, add refresh tokens

CRITICAL/HIGH LESSONS (3)
  L-01KGJ5...: [critical] Never store tokens in localStorage
  L-01KGJ5...: [high] Auth tokens expire silently—check response
```

## What your agent can do

Here's what a typical agent session looks like with Dory:

```bash
# Session start — agent reads context
$ cat .dory/index.yaml
session:
  goal: "Add password reset flow"
  progress: "Email service integrated, templates ready"
  next: "Implement reset endpoint, add rate limiting"
  updated: "2026-02-03T18:30:00Z"

items:
  L-01KGJ5XFHD7520: { type: lesson, oneliner: "SendGrid requires domain verification first", ... }
  D-01KGJ5XFHDSMFY: { type: decision, oneliner: "Use JWT tokens for reset links, 15min expiry", ... }

# Agent works, encounters an issue, records it
$ dory create "Rate limiter must be per-user, not per-IP—VPN users share IPs" --tag auth
Created L-01KGJX7...

# Agent makes an architectural choice
$ dory create "Store reset tokens in Redis with TTL instead of DB" --kind decision --tag auth
Created D-01KGJX8...

# Session end, agent updates context
$ dory context --progress "Reset endpoint done, rate limiting done" --next "Add password validation rules, write tests"
Updated session context

# Other agents work on it, or a new session starts, and everything is already recorded
```

## Agent integration

When you run `dory init`, Dory automatically appends instructions to your `CLAUDE.md` or `AGENTS.md` if they exist, and will create a `DORY.md` with instructions on how to use Dory.
Agents will know to check `.dory/index.yaml` at session start and update context before ending.

## Plugins

Dory has a minimal plugin system for when you need something specific. Plugins can add custom commands, hooks, and item types. They run as separate processes, so a broken plugin won't corrupt your knowledge base.

```bash
dory plugin list              # see available plugins
dory plugin enable my-plugin  # enable for this project
dory plugin run my-plugin     # run a plugin command
```

See [PROTOCOL.md](./docs/plugins.md) for details.

## Install

**With Go (any OS):**
```bash
go install github.com/sibellavia/dory/cmd/dory@latest
```

**Download binary (macOS/Linux/Windows):**

Download from [releases](https://github.com/sibellavia/dory/releases), extract, and add to your PATH:
```bash
# macOS (Apple Silicon)
curl -sL https://github.com/sibellavia/dory/releases/latest/download/dory_Darwin_arm64.tar.gz | tar xz
sudo mv dory /usr/local/bin/

# macOS (Intel)
curl -sL https://github.com/sibellavia/dory/releases/latest/download/dory_Darwin_x86_64.tar.gz | tar xz
sudo mv dory /usr/local/bin/

# Linux (x86_64)
curl -sL https://github.com/sibellavia/dory/releases/latest/download/dory_Linux_x86_64.tar.gz | tar xz
sudo mv dory /usr/local/bin/
```

**Verify installation:**
```bash
dory --version
dory init && dory context
```
