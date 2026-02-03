# postmortem-validator plugin

Capabilities:

- `types`: `postmortem`

Validation rules:

1. `topic` is required
2. `body` is required
3. `body` must include:
   - `## Timeline`
   - `## Action Items`

Usage:

```bash
dory plugin install ./examples/plugins/postmortem-validator --enable
dory type create postmortem "Cache outage" --topic infra --body $'# Postmortem\n\n## Timeline\n...\n\n## Action Items\n...'
```
