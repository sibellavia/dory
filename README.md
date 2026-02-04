<center>🐠</center>

<b><center>Dory is a lightweight, simple knowledge store that supports coding agents across sessions.</center></b>

Modern coding agents are capable: they read code, grep what they need, and build knowledge on the fly. However, they forget between sessions. Even with /compact, knowledge gets summarized, some details are lost and there's no more trace of causality, historical evolution, fixes learned the hard way, and other aspects that may be retained over the time.

Dory captures what would otherwise disappear: learnings from experience you can't grep for, the rationale behind decisions that isn't in the code, project-specific details learned through several sessions and hours, and session state so the next agent knows where you left off. To do so, Dory stores structured knowledge in Doryfile: an append-only storage format consisting of two files: knowledge.dory (the event log) and index.yaml (a snapshot with items and a log offset for O(1) access). 

One may ask: why shuld I use Dory instead of writing one single markdown file? First of all, I built Dory to meet my needs... and to have fun and play with agents! Besides of that: for small, short-lived projects a single markdown file works fine. Dory helps when knowledge grows. Doryfile is queryable, this means you can just do `dory list --tag auth` instead of scrolling through one big file. You can create relations and reference items with each other with `refs`, with the possibility of even having some fancy visualizations directly on the terminal. Agents can load the index and fetch full content only when needed, categorizing type, topic, severity for each item. And should you need something specific, Dory offers a minimalist yet powerful plugin system.

I hope you will enjoy Dory as I do :)

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
