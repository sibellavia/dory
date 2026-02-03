# hook-guard plugin

Capabilities:

- `hooks`: `before_remove`, `after_create`

Behavior:

- Blocks removal of decision entries (IDs starting with `D`)
- Allows all other removals

Usage:

```bash
dory plugin install ./examples/plugins/hook-guard --enable
dory plugin doctor hook-guard
```
